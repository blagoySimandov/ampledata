package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/activities"
)

type EnrichmentWorkflowInput struct {
	JobID                string
	UserID               string
	StripeCustomerID     string
	RowKey               string
	ColumnsMetadata      []*models.ColumnMetadata
	QueryPatterns        []string
	KeyColumnDescription string
	AllowedDomains       []string
	RetryCount           int
	PreviousAttempts     []*models.EnrichmentAttempt
	MaxRetries           int
}

type EnrichmentWorkflowOutput struct {
	RowKey            string
	ExtractedData     map[string]interface{}
	Confidence        map[string]*models.FieldConfidenceInfo
	Sources           []string
	ExtractionHistory []*models.ExtractionHistoryEntry
	Success           bool
	Error             string
	IterationCount    int
}

func mergeSources(serpSources, crawlSources []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range serpSources {
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range crawlSources {
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// mergeBestConfidence merges retry results into the base, keeping whichever
// value has the higher confidence score for each field. Fields that only exist
// in one side are always included.
func mergeBestConfidence(
	base map[string]interface{},
	baseConf map[string]*models.FieldConfidenceInfo,
	retry map[string]interface{},
	retryConf map[string]*models.FieldConfidenceInfo,
) (map[string]interface{}, map[string]*models.FieldConfidenceInfo) {
	if base == nil {
		base = make(map[string]interface{})
	}
	if baseConf == nil {
		baseConf = make(map[string]*models.FieldConfidenceInfo)
	}
	for k, v := range retry {
		current := baseConf[k]
		candidate := retryConf[k]
		if current == nil || (candidate != nil && candidate.Score > current.Score) {
			base[k] = v
			if candidate != nil {
				baseConf[k] = candidate
			}
		}
	}
	return base, baseConf
}

func EnrichmentWorkflow(ctx workflow.Context, input EnrichmentWorkflowInput) (*EnrichmentWorkflowOutput, error) {
	info := workflow.GetInfo(ctx)
	event := logger.NewEnrichmentEvent(input.JobID, input.RowKey, "")
	event.SetWorkflowInfo(info.WorkflowExecution.ID, info.WorkflowExecution.RunID)
	event.SetMetadata("retry_count", input.RetryCount)

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	output := &EnrichmentWorkflowOutput{
		RowKey:         input.RowKey,
		Success:        false,
		IterationCount: input.RetryCount + 1,
	}

	// Generate patterns if this is a retry with feedback
	queryPatterns := input.QueryPatterns
	if input.RetryCount > 0 && len(input.PreviousAttempts) > 0 {
		event.StartStage("PATTERN_REGENERATION")
		var patternsOutput activities.GeneratePatternsOutput
		err := workflow.ExecuteActivity(ctx, "GeneratePatternsWithFeedback", activities.GeneratePatternsWithFeedbackInput{
			JobID:            input.JobID,
			ColumnsMetadata:  input.ColumnsMetadata,
			PreviousAttempts: input.PreviousAttempts,
		}).Get(ctx, &patternsOutput)
		if err != nil {
			// TODO: maybe stop the workflow if pattern generation fails ?
			event.FailStage("PATTERN_REGENERATION", err)
		} else {
			queryPatterns = patternsOutput.Patterns
			event.CompleteStage("PATTERN_REGENERATION", map[string]interface{}{
				"pattern_count": len(patternsOutput.Patterns),
			})
		}
	}

	event.StartStage(models.StageSerpFetched)
	var serpOutput activities.SerpFetchOutput
	err := workflow.ExecuteActivity(ctx, "SerpFetch", activities.SerpFetchInput{
		JobID:          input.JobID,
		RowKey:         input.RowKey,
		ColumnsMetadata: input.ColumnsMetadata,
		QueryPatterns:  queryPatterns,
		AllowedDomains: input.AllowedDomains,
	}).Get(ctx, &serpOutput)
	if err != nil {
		output.Error = fmt.Sprintf("SERP fetch failed: %v", err)
		event.FailStage(models.StageSerpFetched, err)
		event.EmitError(ctx, err)

		workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageFailed,
			Data: &models.StateUpdate{
				Error: &output.Error,
			},
		}).Get(ctx, nil)

		return output, nil
	}
	event.CompleteStage(models.StageSerpFetched, map[string]interface{}{
		"result_count": len(serpOutput.SerpData.Results),
	})

	workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageSerpFetched,
		Data:   nil,
	}).Get(ctx, nil)

	event.StartStage(models.StageDecisionMade)
	var decisionOutput activities.DecisionOutput
	err = workflow.ExecuteActivity(ctx, "MakeDecision", activities.DecisionInput{
		JobID:            input.JobID,
		RowKey:           input.RowKey,
		SerpData:         serpOutput.SerpData,
		ColumnsMetadata:  input.ColumnsMetadata,
		KeyColumnDescription:      input.KeyColumnDescription,
		PreviousAttempts: input.PreviousAttempts,
	}).Get(ctx, &decisionOutput)
	if err != nil {
		output.Error = fmt.Sprintf("Decision making failed: %v", err)
		event.FailStage(models.StageDecisionMade, err)
		event.EmitError(ctx, err)

		workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageFailed,
			Data: &models.StateUpdate{
				Error: &output.Error,
			},
		}).Get(ctx, nil)

		return output, nil
	}
	event.CompleteStage(models.StageDecisionMade, nil)

	workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageDecisionMade,
		Data:   nil,
	}).Get(ctx, nil)

	event.StartStage(models.StageCrawled)
	var crawlOutput activities.CrawlOutput
	err = workflow.ExecuteActivity(ctx, "Crawl", activities.CrawlInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		SerpData:        serpOutput.SerpData,
		Decision:        decisionOutput.Decision,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(ctx, &crawlOutput)
	if err != nil {
		output.Error = fmt.Sprintf("Crawling failed: %v", err)
		event.FailStage(models.StageCrawled, err)
		event.EmitError(ctx, err)

		workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageFailed,
			Data: &models.StateUpdate{
				Error: &output.Error,
			},
		}).Get(ctx, nil)

		return output, nil
	}
	event.CompleteStage(models.StageCrawled, map[string]interface{}{
		"crawled_urls": len(crawlOutput.CrawlResults.Sources),
	})

	workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageCrawled,
		Data:   nil,
	}).Get(ctx, nil)

	event.StartStage(models.StageEnriched)
	var extractOutput activities.ExtractOutput
	err = workflow.ExecuteActivity(ctx, "Extract", activities.ExtractInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		Decision:        decisionOutput.Decision,
		CrawlResults:    crawlOutput.CrawlResults,
		ColumnsMetadata: input.ColumnsMetadata,
		KeyColumnDescription:      input.KeyColumnDescription,
	}).Get(ctx, &extractOutput)
	if err != nil {
		output.Error = fmt.Sprintf("Extraction failed: %v", err)
		event.FailStage(models.StageEnriched, err)
		event.EmitError(ctx, err)

		workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageFailed,
			Data: &models.StateUpdate{
				Error: &output.Error,
			},
		}).Get(ctx, nil)

		return output, nil
	}

	extractedFieldCount := 0
	if extractOutput.ExtractedData != nil {
		extractedFieldCount = len(extractOutput.ExtractedData)
	}
	event.CompleteStage(models.StageEnriched, map[string]interface{}{
		"fields_extracted": extractedFieldCount,
		"confidence":       extractOutput.Confidence,
	})

	enrichedData := models.StateUpdate{}
	if extractOutput.ExtractedData != nil {
		enrichedData.ExtractedData = extractOutput.ExtractedData
	}
	if extractOutput.Confidence != nil {
		enrichedData.Confidence = extractOutput.Confidence
	}
	allSources := mergeSources(decisionOutput.Decision.SourceURLs, crawlOutput.CrawlResults.Sources)
	if len(allSources) > 0 {
		enrichedData.Sources = allSources
	}

	// Always record this attempt in the extraction history, regardless of confidence
	historyEntry := &models.ExtractionHistoryEntry{
		AttemptNumber: input.RetryCount + 1,
		ExtractedData: extractOutput.ExtractedData,
		Confidence:    extractOutput.Confidence,
		Sources:       allSources,
		Reasoning:     extractOutput.Reasoning,
	}
	enrichedData.ExtractionHistory = []*models.ExtractionHistoryEntry{historyEntry}

	workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageEnriched,
		Data:   &enrichedData,
	}).Get(ctx, nil)

	output.ExtractedData = extractOutput.ExtractedData
	output.Confidence = extractOutput.Confidence
	output.Sources = allSources
	output.ExtractionHistory = []*models.ExtractionHistoryEntry{historyEntry}
	output.Success = true

	var feedbackOutput activities.FeedbackAnalysisOutput
	workflow.ExecuteActivity(ctx, "AnalyzeFeedback", activities.FeedbackAnalysisInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		ExtractedData:   extractOutput.ExtractedData,
		Confidence:      extractOutput.Confidence,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(ctx, &feedbackOutput)

	if feedbackOutput.NeedsFeedback && input.RetryCount < input.MaxRetries {
		currentAttempt := &models.EnrichmentAttempt{
			AttemptNumber:        input.RetryCount + 1,
			QueryPatterns:        queryPatterns,
			LowConfidenceColumns: feedbackOutput.LowConfidenceColumns,
			MissingColumns:       feedbackOutput.MissingColumns,
			CrawledURLs:          crawlOutput.CrawlResults.Sources,
		}

		previousAttempts := append(input.PreviousAttempts, currentAttempt)

		problematicColumns := make(map[string]bool)
		for _, col := range feedbackOutput.LowConfidenceColumns {
			problematicColumns[col] = true
		}
		for _, col := range feedbackOutput.MissingColumns {
			problematicColumns[col] = true
		}

		filteredMetadata := []*models.ColumnMetadata{}
		for _, col := range input.ColumnsMetadata {
			if problematicColumns[col.Name] {
				filteredMetadata = append(filteredMetadata, col)
			}
		}

		retryInput := EnrichmentWorkflowInput{
			JobID:                input.JobID,
			UserID:               input.UserID,
			StripeCustomerID:     input.StripeCustomerID,
			RowKey:               input.RowKey,
			ColumnsMetadata:      filteredMetadata,
			QueryPatterns:        input.QueryPatterns,
			KeyColumnDescription: input.KeyColumnDescription,
			AllowedDomains:       input.AllowedDomains,
			RetryCount:           input.RetryCount + 1,
			PreviousAttempts:     previousAttempts,
			MaxRetries:           input.MaxRetries,
		}

		retryOutput, err := EnrichmentWorkflow(ctx, retryInput)
		if err != nil {
			return output, err
		}

		// Merge: keep the highest-confidence value per field across this attempt
		// and all retry attempts. The retry only re-ran problematic columns, but
		// its LLM may still return values for other fields — only upgrade when
		// confidence actually improves.
		output.ExtractedData, output.Confidence = mergeBestConfidence(
			output.ExtractedData, output.Confidence,
			retryOutput.ExtractedData, retryOutput.Confidence,
		)
		// Deduplicate: the retry may select the same SERP URLs or crawl targets as the
		// initial attempt if the same sources are relevant to the missing columns.
		// This shouldn't usually happen since the feedback loop generates different query
		// patterns specifically to find better sources for the missing/low-confidence columns.
		output.Sources = mergeSources(output.Sources, retryOutput.Sources)
		output.ExtractionHistory = append(output.ExtractionHistory, retryOutput.ExtractionHistory...)

		output.IterationCount = retryOutput.IterationCount

		// Persist the best-merged result. The retry's own UpdateState already wrote
		// its values to the DB; we overwrite here with the correctly merged data.
		bestMerged := &models.StateUpdate{
			ExtractedData: output.ExtractedData,
			Confidence:    output.Confidence,
			Sources:       output.Sources,
		}
		workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageCompleted,
			Data:   bestMerged,
		}).Get(ctx, nil)

		event.EmitSuccess(ctx)

		if input.RetryCount == 0 {
			var reportErr error
			workflow.ExecuteActivity(ctx, "ReportUsage", activities.ReportUsageInput{
				StripeCustomerID: input.StripeCustomerID,
				Credits:          len(input.ColumnsMetadata),
			}).Get(ctx, &reportErr)
			if reportErr != nil {
				event.SetMetadata("billing_error", reportErr.Error())
			}
		}

		return output, nil
	}

	workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageCompleted,
		Data:   nil,
	}).Get(ctx, nil)

	event.EmitSuccess(ctx)

	if input.RetryCount == 0 {
		var reportErr error
		workflow.ExecuteActivity(ctx, "ReportUsage", activities.ReportUsageInput{
			StripeCustomerID: input.StripeCustomerID,
			Credits:          len(input.ColumnsMetadata),
		}).Get(ctx, &reportErr)
		if reportErr != nil {
			event.SetMetadata("billing_error", reportErr.Error())
		}
	}

	return output, nil
}
