package workflows

import (
	"fmt"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/activities"
)

type JobWorkflowInput struct {
	JobID                string
	UserID               string
	StripeCustomerID     string
	RowKeys              []string
	ColumnsMetadata      []*models.ColumnMetadata
	KeyColumnDescription *string
	MaxRetries           int
}

type JobWorkflowOutput struct {
	JobID          string
	TotalRows      int
	SuccessfulRows int
	FailedRows     int
	CompletedAt    time.Time
}

func JobWorkflow(ctx workflow.Context, input JobWorkflowInput) (*JobWorkflowOutput, error) {
	info := workflow.GetInfo(ctx)
	event := logger.NewJobEvent(input.JobID, "")
	event.SetWorkflowInfo(info.WorkflowExecution.ID, info.WorkflowExecution.RunID)
	event.TotalRows = len(input.RowKeys)

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
		JobID:     input.JobID,
		TotalRows: len(input.RowKeys),
	}

	if err := workflow.ExecuteActivity(activityCtx, "InitializeJob", input.JobID, input.RowKeys).Get(activityCtx, nil); err != nil {
		event.EmitError(ctx, fmt.Errorf("job initialization failed: %w", err))
		return nil, fmt.Errorf("job initialization failed: %w", err)
	}

	var patternsOutput activities.GeneratePatternsOutput
	err := workflow.ExecuteActivity(activityCtx, "GeneratePatterns", activities.GeneratePatternsInput{
		JobID:           input.JobID,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(activityCtx, &patternsOutput)
	if err != nil {
		patternsOutput.Patterns = []string{"%entity"}
	}
	event.SetMetadata("pattern_count", len(patternsOutput.Patterns))

	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 10 * time.Minute,
		WorkflowTaskTimeout:      1 * time.Minute,
		ParentClosePolicy:        *enumspb.PARENT_CLOSE_POLICY_TERMINATE.Enum(),
	})

	keyColumnDescription := ""
	if input.KeyColumnDescription != nil {
		keyColumnDescription = *input.KeyColumnDescription
	}

	sem := &workflowSemaphore{ctx: ctx, limit: config.Load().ConcurrencyRowEnrichmentLimit}

	for _, rowKey := range input.RowKeys {
		sem.Acquire()
		sem.Add(workflow.ExecuteChildWorkflow(childCtx, EnrichmentWorkflow, EnrichmentWorkflowInput{
			JobID:                input.JobID,
			UserID:               input.UserID,
			StripeCustomerID:     input.StripeCustomerID,
			RowKey:               rowKey,
			ColumnsMetadata:      input.ColumnsMetadata,
			QueryPatterns:        patternsOutput.Patterns,
			KeyColumnDescription: keyColumnDescription,
			RetryCount:           0,
			PreviousAttempts:     []*models.EnrichmentAttempt{},
			MaxRetries:           input.MaxRetries,
		}))
	}

	for _, future := range sem.Futures() {
		var rowOutput EnrichmentWorkflowOutput
		if err := future.Get(ctx, &rowOutput); err != nil {
			workflow.GetLogger(ctx).Error("child workflow failed", "error", err)
			output.FailedRows++
		} else if rowOutput.Success {
			output.SuccessfulRows++
		} else if !rowOutput.Cancelled {
			output.FailedRows++
		}
	}

	event.Completed = output.SuccessfulRows
	event.Failed = output.FailedRows

	workflow.ExecuteActivity(activityCtx, "IncrementJobCredits", activities.IncrementJobCreditsInput{
		JobID:   input.JobID,
		Credits: output.SuccessfulRows,
	}).Get(activityCtx, nil)

	if err := workflow.ExecuteActivity(activityCtx, "CompleteJob", input.JobID).Get(activityCtx, nil); err != nil {
		event.EmitError(ctx, err)
		return nil, err
	}

	output.CompletedAt = workflow.Now(ctx)
	event.EmitSuccess(ctx)
	return output, nil
}
