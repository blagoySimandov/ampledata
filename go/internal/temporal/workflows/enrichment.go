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
	JobID            string
	UserID           string
	StripeCustomerID string
	RowKey           string
	ColumnsMetadata  []*models.ColumnMetadata
	QueryPatterns    []string
	EntityType       string
	RetryCount       int
	PreviousAttempts []*models.EnrichmentAttempt
	MaxRetries       int
	// SourceData holds the values of existing CSV columns for this row.
	// When non-empty, the imputation stage runs before web-search enrichment.
	SourceData map[string]string
}

type EnrichmentWorkflowOutput struct {
	RowKey         string
	ExtractedData  map[string]interface{}
	Confidence     map[string]*models.FieldConfidenceInfo
	Sources        []string
	Success        bool
	Error          string
	IterationCount int
}

// imputationConfidenceThreshold is the minimum score for an imputed field to be
// considered "done" — i.e., it will not be re-fetched from the web.
const imputationConfidenceThreshold = 0.8

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

	// ── Stage 0: Column Imputation ────────────────────────────────────────────
	// Run only on the first attempt (imputed results are carried into retries).
	// Columns imputed with high confidence are removed from ColumnsMetadata so
	// the expensive web-search pipeline only processes what is still missing.
	imputedData := make(map[string]interface{})
	imputedConfidence := make(map[string]*models.FieldConfidenceInfo)

	if input.RetryCount == 0 && len(input.SourceData) > 0 {
		event.StartStage("IMPUTATION")
		var imputeOutput activities.ImputeOutput
		workflow.ExecuteActivity(ctx, "Impute", activities.ImputeInput{
			JobID:           input.JobID,
			RowKey:          input.RowKey,
			SourceData:      input.SourceData,
			ColumnsMetadata: input.ColumnsMetadata,
		}).Get(ctx, &imputeOutput)

		if imputeOutput.ImputedData != nil {
			imputedData = imputeOutput.ImputedData
		}
		if imputeOutput.Confidence != nil {
			imputedConfidence = imputeOutput.Confidence
		}

		event.CompleteStage("IMPUTATION", map[string]interface{}{
			"imputed_count": len(imputedData),
		})

		workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageImputed,
			Data:   nil,
		}).Get(ctx, nil)

		// If every target column was imputed with sufficient confidence, we are done.
		if allImputedWithHighConfidence(input.ColumnsMetadata, imputedData, imputedConfidence) {
			output.ExtractedData = imputedData
			output.Confidence = imputedConfidence
			output.Sources = []string{}
			output.Success = true

			workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
				JobID:  input.JobID,
				RowKey: input.RowKey,
				Stage:  models.StageCompleted,
				Data: &models.StateUpdate{
					ExtractedData: imputedData,
					Confidence:    imputedConfidence,
				},
			}).Get(ctx, nil)

			event.EmitSuccess(ctx)

			if input.RetryCount == 0 {
				var reportErr error
				workflow.ExecuteActivity(ctx, "ReportUsage", activities.ReportUsageInput{
					StripeCustomerID: input.StripeCustomerID,
					Credits:          len(output.ExtractedData),
				}).Get(ctx, &reportErr)
				if reportErr != nil {
					event.SetMetadata("billing_error", reportErr.Error())
				}
			}

			return output, nil
		}

		// Remove well-imputed columns so the web-search pipeline only handles
		// the remaining ones.
		input.ColumnsMetadata = filterNotImputed(input.ColumnsMetadata, imputedData, imputedConfidence)
	}

	// ── Stage 1: Pattern Regeneration (on retry) ─────────────────────────────
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

	// ── Stage 2: SERP Fetch ───────────────────────────────────────────────────
	event.StartStage(models.StageSerpFetched)
	var serpOutput activities.SerpFetchOutput
	err := workflow.ExecuteActivity(ctx, "SerpFetch", activities.SerpFetchInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		ColumnsMetadata: input.ColumnsMetadata,
		QueryPatterns:   queryPatterns,
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

	// ── Stage 3: Decision Making ──────────────────────────────────────────────
	event.StartStage(models.StageDecisionMade)
	var decisionOutput activities.DecisionOutput
	err = workflow.ExecuteActivity(ctx, "MakeDecision", activities.DecisionInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		SerpData:        serpOutput.SerpData,
		ColumnsMetadata: input.ColumnsMetadata,
		EntityType:      input.EntityType,
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

	// ── Stage 4: Crawl ────────────────────────────────────────────────────────
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

	// ── Stage 5: Extraction ───────────────────────────────────────────────────
	event.StartStage(models.StageEnriched)
	var extractOutput activities.ExtractOutput
	err = workflow.ExecuteActivity(ctx, "Extract", activities.ExtractInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		Decision:        decisionOutput.Decision,
		CrawlResults:    crawlOutput.CrawlResults,
		ColumnsMetadata: input.ColumnsMetadata,
		EntityType:      input.EntityType,
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

	// Merge imputed data with web-search extracted data.
	// Web-search results take precedence for columns that appear in both.
	mergedData, mergedConfidence := mergeImputedWithExtracted(imputedData, imputedConfidence, extractOutput.ExtractedData, extractOutput.Confidence)

	extractedFieldCount := len(mergedData)
	event.CompleteStage(models.StageEnriched, map[string]interface{}{
		"fields_extracted": extractedFieldCount,
		"confidence":       mergedConfidence,
	})

	enrichedData := models.StateUpdate{}
	if len(mergedData) > 0 {
		enrichedData.ExtractedData = mergedData
	}
	if len(mergedConfidence) > 0 {
		enrichedData.Confidence = mergedConfidence
	}
	if crawlOutput.CrawlResults != nil && crawlOutput.CrawlResults.Sources != nil {
		enrichedData.Sources = crawlOutput.CrawlResults.Sources
	}

	workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageEnriched,
		Data:   &enrichedData,
	}).Get(ctx, nil)

	output.ExtractedData = mergedData
	output.Confidence = mergedConfidence
	output.Sources = crawlOutput.CrawlResults.Sources
	output.Success = true

	// ── Stage 6: Feedback Analysis ────────────────────────────────────────────
	var feedbackOutput activities.FeedbackAnalysisOutput
	workflow.ExecuteActivity(ctx, "AnalyzeFeedback", activities.FeedbackAnalysisInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		ExtractedData:   mergedData,
		Confidence:      mergedConfidence,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(ctx, &feedbackOutput)

	if feedbackOutput.NeedsFeedback && input.RetryCount < input.MaxRetries {
		currentAttempt := &models.EnrichmentAttempt{
			AttemptNumber:        input.RetryCount + 1,
			QueryPatterns:        queryPatterns,
			LowConfidenceColumns: feedbackOutput.LowConfidenceColumns,
			MissingColumns:       feedbackOutput.MissingColumns,
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
			JobID:            input.JobID,
			UserID:           input.UserID,
			StripeCustomerID: input.StripeCustomerID,
			RowKey:           input.RowKey,
			ColumnsMetadata:  filteredMetadata,
			QueryPatterns:    input.QueryPatterns,
			EntityType:       input.EntityType,
			RetryCount:       input.RetryCount + 1,
			PreviousAttempts: previousAttempts,
			MaxRetries:       input.MaxRetries,
			// SourceData is intentionally omitted on retries: imputation already ran
			// in the first pass, and its results are merged into mergedData above.
		}

		retryOutput, err := EnrichmentWorkflow(ctx, retryInput)
		if err != nil {
			return output, err
		}

		if output.ExtractedData == nil {
			output.ExtractedData = make(map[string]interface{})
		}
		if output.Confidence == nil {
			output.Confidence = make(map[string]*models.FieldConfidenceInfo)
		}

		if retryOutput.ExtractedData != nil {
			for k, v := range retryOutput.ExtractedData {
				output.ExtractedData[k] = v
			}
		}
		if retryOutput.Confidence != nil {
			for k, v := range retryOutput.Confidence {
				output.Confidence[k] = v
			}
		}
		if retryOutput.Sources != nil {
			output.Sources = append(output.Sources, retryOutput.Sources...)
		}

		output.IterationCount = retryOutput.IterationCount

		if input.RetryCount == 0 {
			var reportErr error
			workflow.ExecuteActivity(ctx, "ReportUsage", activities.ReportUsageInput{
				StripeCustomerID: input.StripeCustomerID,
				Credits:          len(output.ExtractedData),
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
			Credits:          len(output.ExtractedData),
		}).Get(ctx, &reportErr)
		if reportErr != nil {
			event.SetMetadata("billing_error", reportErr.Error())
		}
	}

	return output, nil
}

// allImputedWithHighConfidence returns true when every target column has been
// imputed with a confidence score at or above imputationConfidenceThreshold.
func allImputedWithHighConfidence(
	targetCols []*models.ColumnMetadata,
	imputedData map[string]interface{},
	imputedConfidence map[string]*models.FieldConfidenceInfo,
) bool {
	if len(imputedData) == 0 {
		return false
	}
	for _, col := range targetCols {
		conf, ok := imputedConfidence[col.Name]
		if !ok || conf.Score < imputationConfidenceThreshold {
			return false
		}
		if _, exists := imputedData[col.Name]; !exists {
			return false
		}
	}
	return true
}

// filterNotImputed returns only the columns that were NOT successfully imputed
// with high confidence, so the web-search pipeline focuses on what is still missing.
func filterNotImputed(
	cols []*models.ColumnMetadata,
	imputedData map[string]interface{},
	imputedConfidence map[string]*models.FieldConfidenceInfo,
) []*models.ColumnMetadata {
	var remaining []*models.ColumnMetadata
	for _, col := range cols {
		conf, hasConf := imputedConfidence[col.Name]
		_, hasData := imputedData[col.Name]
		if hasData && hasConf && conf.Score >= imputationConfidenceThreshold {
			continue
		}
		remaining = append(remaining, col)
	}
	return remaining
}

// mergeImputedWithExtracted combines imputed and web-search extracted data.
// Web-search results take precedence when a column appears in both.
func mergeImputedWithExtracted(
	imputedData map[string]interface{},
	imputedConfidence map[string]*models.FieldConfidenceInfo,
	extractedData map[string]interface{},
	extractedConfidence map[string]*models.FieldConfidenceInfo,
) (map[string]interface{}, map[string]*models.FieldConfidenceInfo) {
	merged := make(map[string]interface{})
	mergedConf := make(map[string]*models.FieldConfidenceInfo)

	for k, v := range imputedData {
		merged[k] = v
		if c, ok := imputedConfidence[k]; ok {
			mergedConf[k] = c
		}
	}
	for k, v := range extractedData {
		merged[k] = v
		if c, ok := extractedConfidence[k]; ok {
			mergedConf[k] = c
		}
	}

	if len(merged) == 0 {
		return nil, nil
	}
	return merged, mergedConf
}
