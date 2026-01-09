package workflows

import (
	"fmt"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/activities"
)

// JobWorkflowInput contains all the input needed to process a job
type JobWorkflowInput struct {
	JobID           string
	RowKeys         []string
	ColumnsMetadata []*models.ColumnMetadata
	EntityType      *string
}

// JobWorkflowOutput contains the overall job result
type JobWorkflowOutput struct {
	JobID          string
	TotalRows      int
	SuccessfulRows int
	FailedRows     int
	CompletedAt    time.Time
}

// JobWorkflow orchestrates the enrichment of all rows in a job
// It spawns child workflows for each row and tracks overall progress
func JobWorkflow(ctx workflow.Context, input JobWorkflowInput) (*JobWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting job workflow",
		"jobID", input.JobID,
		"totalRows", len(input.RowKeys))

	// Configure activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)

	output := &JobWorkflowOutput{
		JobID:          input.JobID,
		TotalRows:      len(input.RowKeys),
		SuccessfulRows: 0,
		FailedRows:     0,
	}

	// Step 1: Initialize job in database
	logger.Info("Initializing job", "jobID", input.JobID)

	err := workflow.ExecuteActivity(activityCtx, "InitializeJob", input.JobID, input.RowKeys).Get(activityCtx, nil)
	if err != nil {
		logger.Error("Failed to initialize job", "error", err)
		return nil, fmt.Errorf("job initialization failed: %w", err)
	}

	// Step 2: Generate query patterns
	logger.Info("Generating query patterns", "jobID", input.JobID)

	var patternsOutput activities.GeneratePatternsOutput
	err = workflow.ExecuteActivity(activityCtx, "GeneratePatterns", activities.GeneratePatternsInput{
		JobID:           input.JobID,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(activityCtx, &patternsOutput)
	if err != nil {
		logger.Warn("Pattern generation failed, using fallback", "error", err)
		// Use fallback pattern
		patternsOutput.Patterns = []string{"%entity"}
	}

	logger.Info("Query patterns generated",
		"jobID", input.JobID,
		"patternCount", len(patternsOutput.Patterns))

	// Step 3: Start child workflows for each row
	// We process rows in parallel using child workflows
	logger.Info("Starting enrichment for all rows", "rowCount", len(input.RowKeys))

	// Configure child workflow options
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 10 * time.Minute,
		WorkflowTaskTimeout:      1 * time.Minute,
		ParentClosePolicy:        *enumspb.PARENT_CLOSE_POLICY_TERMINATE.Enum(), // Cancel children if parent is cancelled
	}
	childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)

	// Start all child workflows
	childFutures := make([]workflow.ChildWorkflowFuture, 0, len(input.RowKeys))

	for _, rowKey := range input.RowKeys {
		// Check if workflow is cancelled
		if workflow.GetInfo(ctx).GetCurrentHistoryLength() > 0 && ctx.Err() != nil {
			logger.Info("Job workflow cancelled, stopping row processing")
			break
		}

		childInput := EnrichmentWorkflowInput{
			JobID:           input.JobID,
			RowKey:          rowKey,
			ColumnsMetadata: input.ColumnsMetadata,
			QueryPatterns:   patternsOutput.Patterns,
			RetryCount:      0,
		}

		// Start child workflow
		childWorkflow := workflow.ExecuteChildWorkflow(childCtx, EnrichmentWorkflow, childInput)
		childFutures = append(childFutures, childWorkflow)

		logger.Debug("Started enrichment workflow for row", "rowKey", rowKey)
	}

	logger.Info("All enrichment workflows started", "count", len(childFutures))

	// Step 4: Wait for all child workflows to complete
	for i, future := range childFutures {
		var rowOutput EnrichmentWorkflowOutput
		err := future.Get(ctx, &rowOutput)

		if err != nil {
			logger.Error("Row enrichment workflow failed",
				"rowIndex", i,
				"error", err)
			output.FailedRows++
		} else if rowOutput.Success {
			output.SuccessfulRows++
			logger.Debug("Row enrichment completed",
				"rowKey", rowOutput.RowKey,
				"fieldsExtracted", len(rowOutput.ExtractedData))
		} else {
			output.FailedRows++
			logger.Warn("Row enrichment unsuccessful",
				"rowKey", rowOutput.RowKey,
				"error", rowOutput.Error)
		}
	}

	logger.Info("All row enrichments completed",
		"successful", output.SuccessfulRows,
		"failed", output.FailedRows)

	// Step 5: Mark job as completed
	logger.Info("Completing job", "jobID", input.JobID)

	err = workflow.ExecuteActivity(activityCtx, "CompleteJob", input.JobID).Get(activityCtx, nil)
	if err != nil {
		logger.Error("Failed to complete job", "error", err)
		// Don't fail the workflow, job is essentially done
	}

	output.CompletedAt = workflow.Now(ctx)

	logger.Info("Job workflow completed",
		"jobID", input.JobID,
		"totalRows", output.TotalRows,
		"successful", output.SuccessfulRows,
		"failed", output.FailedRows)

	return output, nil
}
