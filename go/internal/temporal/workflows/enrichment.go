package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/activities"
)

// EnrichmentWorkflowInput contains all the input needed to enrich a single row
type EnrichmentWorkflowInput struct {
	JobID           string
	RowKey          string
	ColumnsMetadata []*models.ColumnMetadata
	QueryPatterns   []string
	RetryCount      int // For tracking feedback loop iterations
}

// EnrichmentWorkflowOutput contains the enrichment result for a single row
type EnrichmentWorkflowOutput struct {
	RowKey         string
	ExtractedData  map[string]interface{}
	Confidence     map[string]*models.FieldConfidenceInfo
	Sources        []string
	Success        bool
	Error          string
	IterationCount int // Number of feedback iterations
}

// EnrichmentWorkflow processes a single row through the enrichment pipeline
// This workflow supports feedback loops for rows with low confidence or missing data
func EnrichmentWorkflow(ctx workflow.Context, input EnrichmentWorkflowInput) (*EnrichmentWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting enrichment workflow",
		"jobID", input.JobID,
		"rowKey", input.RowKey,
		"retryCount", input.RetryCount)

	// Configure activity options with appropriate timeouts
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

	// Stage 1: SERP Fetch
	var serpOutput activities.SerpFetchOutput
	err := workflow.ExecuteActivity(ctx, "SerpFetch", activities.SerpFetchInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		ColumnsMetadata: input.ColumnsMetadata,
		QueryPatterns:   input.QueryPatterns,
	}).Get(ctx, &serpOutput)
	if err != nil {
		output.Error = fmt.Sprintf("SERP fetch failed: %v", err)
		logger.Error("SERP fetch failed", "error", err)

		// Update state to failed
		_ = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageFailed,
			Data: map[string]interface{}{
				"error": output.Error,
			},
		}).Get(ctx, nil)

		return output, nil
	}

	// Update state: SERP fetched
	err = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageSerpFetched,
		Data: map[string]interface{}{
			"serp_data": serpOutput.SerpData,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to update state after SERP fetch", "error", err)
	}

	// Stage 2: Decision Making
	var decisionOutput activities.DecisionOutput
	err = workflow.ExecuteActivity(ctx, "MakeDecision", activities.DecisionInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		SerpData:        serpOutput.SerpData,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(ctx, &decisionOutput)
	if err != nil {
		output.Error = fmt.Sprintf("Decision making failed: %v", err)
		logger.Error("Decision making failed", "error", err)

		_ = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageFailed,
			Data: map[string]interface{}{
				"error": output.Error,
			},
		}).Get(ctx, nil)

		return output, nil
	}

	// Update state: Decision made
	err = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageDecisionMade,
		Data: map[string]interface{}{
			"decision": decisionOutput.Decision,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to update state after decision", "error", err)
	}

	// Stage 3: Crawl (if needed)
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
		logger.Error("Crawling failed", "error", err)

		_ = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageFailed,
			Data: map[string]interface{}{
				"error": output.Error,
			},
		}).Get(ctx, nil)

		return output, nil
	}

	// Update state: Crawled
	err = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageCrawled,
		Data: map[string]interface{}{
			"crawl_results": crawlOutput.CrawlResults,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to update state after crawl", "error", err)
	}

	// Stage 4: Extract
	var extractOutput activities.ExtractOutput
	err = workflow.ExecuteActivity(ctx, "Extract", activities.ExtractInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		Decision:        decisionOutput.Decision,
		CrawlResults:    crawlOutput.CrawlResults,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(ctx, &extractOutput)
	if err != nil {
		output.Error = fmt.Sprintf("Extraction failed: %v", err)
		logger.Error("Extraction failed", "error", err)

		_ = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
			JobID:  input.JobID,
			RowKey: input.RowKey,
			Stage:  models.StageFailed,
			Data: map[string]interface{}{
				"error": output.Error,
			},
		}).Get(ctx, nil)

		return output, nil
	}

	// Update state: Enriched
	err = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageEnriched,
		Data: map[string]interface{}{
			"extracted_data": extractOutput.ExtractedData,
			"confidence":     extractOutput.Confidence,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to update state after extraction", "error", err)
	}

	// Populate output
	output.ExtractedData = extractOutput.ExtractedData
	output.Confidence = extractOutput.Confidence
	output.Sources = crawlOutput.CrawlResults.Sources
	output.Success = true

	// Stage 5: Feedback Analysis (for future feedback loop support)
	var feedbackOutput activities.FeedbackAnalysisOutput
	err = workflow.ExecuteActivity(ctx, "AnalyzeFeedback", activities.FeedbackAnalysisInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		ExtractedData:   extractOutput.ExtractedData,
		Confidence:      extractOutput.Confidence,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(ctx, &feedbackOutput)

	if err != nil {
		logger.Warn("Feedback analysis failed", "error", err)
	} else {
		logger.Info("Feedback analysis completed",
			"needsFeedback", feedbackOutput.NeedsFeedback,
			"avgConfidence", feedbackOutput.AverageConfidence,
			"missingColumns", len(feedbackOutput.MissingColumns),
			"lowConfidenceColumns", len(feedbackOutput.LowConfidenceColumns))

		// Future: Implement feedback loop here
		// If feedbackOutput.NeedsFeedback && input.RetryCount < maxRetries:
		//   - Generate new query patterns based on missing/low-confidence columns
		//   - Use workflow.ExecuteChildWorkflow or continue-as-new to retry
		//   - Pass feedback information to adjust the enrichment strategy
		//
		// Example (not implemented yet):
		// if feedbackOutput.NeedsFeedback && input.RetryCount < 2 {
		//     logger.Info("Feedback needed, will retry with adjusted parameters")
		//     // Could generate new patterns targeting missing columns
		//     // newPatterns := generateFeedbackPatterns(feedbackOutput)
		//     // return continueAsNew with updated input
		// }
	}

	// Update state: Completed
	err = workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageCompleted,
		Data:   nil,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to update state to completed", "error", err)
	}

	logger.Info("Enrichment workflow completed",
		"rowKey", input.RowKey,
		"success", output.Success,
		"fieldsExtracted", len(output.ExtractedData))

	return output, nil
}
