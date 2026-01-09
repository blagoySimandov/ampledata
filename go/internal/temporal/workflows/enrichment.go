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
	JobID           string
	RowKey          string
	ColumnsMetadata []*models.ColumnMetadata
	QueryPatterns   []string
	RetryCount      int
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

	event.StartStage(models.StageSerpFetched)
	var serpOutput activities.SerpFetchOutput
	err := workflow.ExecuteActivity(ctx, "SerpFetch", activities.SerpFetchInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		ColumnsMetadata: input.ColumnsMetadata,
		QueryPatterns:   input.QueryPatterns,
	}).Get(ctx, &serpOutput)
	if err != nil {
		output.Error = fmt.Sprintf("SERP fetch failed: %v", err)
		event.FailStage(models.StageSerpFetched, err)
		event.EmitError(ctx, models.StageSerpFetched, err)

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
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		SerpData:        serpOutput.SerpData,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(ctx, &decisionOutput)
	if err != nil {
		output.Error = fmt.Sprintf("Decision making failed: %v", err)
		event.FailStage(models.StageDecisionMade, err)
		event.EmitError(ctx, models.StageDecisionMade, err)

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
		event.EmitError(ctx, models.StageCrawled, err)

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
	}).Get(ctx, &extractOutput)
	if err != nil {
		output.Error = fmt.Sprintf("Extraction failed: %v", err)
		event.FailStage(models.StageEnriched, err)
		event.EmitError(ctx, models.StageEnriched, err)

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
	if crawlOutput.CrawlResults != nil && crawlOutput.CrawlResults.Sources != nil {
		enrichedData.Sources = crawlOutput.CrawlResults.Sources
	}

	workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageEnriched,
		Data:   &enrichedData,
	}).Get(ctx, nil)

	output.ExtractedData = extractOutput.ExtractedData
	output.Confidence = extractOutput.Confidence
	output.Sources = crawlOutput.CrawlResults.Sources
	output.Success = true

	var feedbackOutput activities.FeedbackAnalysisOutput
	workflow.ExecuteActivity(ctx, "AnalyzeFeedback", activities.FeedbackAnalysisInput{
		JobID:           input.JobID,
		RowKey:          input.RowKey,
		ExtractedData:   extractOutput.ExtractedData,
		Confidence:      extractOutput.Confidence,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(ctx, &feedbackOutput)

	workflow.ExecuteActivity(ctx, "UpdateState", activities.StateUpdateInput{
		JobID:  input.JobID,
		RowKey: input.RowKey,
		Stage:  models.StageCompleted,
		Data:   nil,
	}).Get(ctx, nil)

	event.EmitSuccess(ctx, models.StageCompleted)

	return output, nil
}
