package workflows

import (
	"fmt"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/activities"
)

type JobWorkflowInput struct {
	JobID           string
	RowKeys         []string
	ColumnsMetadata []*models.ColumnMetadata
	EntityType      *string
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
		JobID:          input.JobID,
		TotalRows:      len(input.RowKeys),
		SuccessfulRows: 0,
		FailedRows:     0,
	}

	err := workflow.ExecuteActivity(activityCtx, "InitializeJob", input.JobID, input.RowKeys).Get(activityCtx, nil)
	if err != nil {
		event.EmitError(ctx, fmt.Errorf("job initialization failed: %w", err))
		return nil, fmt.Errorf("job initialization failed: %w", err)
	}

	var patternsOutput activities.GeneratePatternsOutput
	err = workflow.ExecuteActivity(activityCtx, "GeneratePatterns", activities.GeneratePatternsInput{
		JobID:           input.JobID,
		ColumnsMetadata: input.ColumnsMetadata,
	}).Get(activityCtx, &patternsOutput)
	if err != nil {
		patternsOutput.Patterns = []string{"%entity"}
	}
	event.SetMetadata("pattern_count", len(patternsOutput.Patterns))

	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 10 * time.Minute,
		WorkflowTaskTimeout:      1 * time.Minute,
		ParentClosePolicy:        *enumspb.PARENT_CLOSE_POLICY_TERMINATE.Enum(),
	}
	childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)

	childFutures := make([]workflow.ChildWorkflowFuture, 0, len(input.RowKeys))

	for _, rowKey := range input.RowKeys {
		if workflow.GetInfo(ctx).GetCurrentHistoryLength() > 0 && ctx.Err() != nil {
			break
		}

		entityType := ""
		if input.EntityType != nil {
			entityType = *input.EntityType
		}

		childInput := EnrichmentWorkflowInput{
			JobID:           input.JobID,
			RowKey:          rowKey,
			ColumnsMetadata: input.ColumnsMetadata,
			QueryPatterns:   patternsOutput.Patterns,
			EntityType:      entityType,
			RetryCount:      0,
		}

		childWorkflow := workflow.ExecuteChildWorkflow(childCtx, EnrichmentWorkflow, childInput)
		childFutures = append(childFutures, childWorkflow)
	}

	for _, future := range childFutures {
		var rowOutput EnrichmentWorkflowOutput
		err := future.Get(ctx, &rowOutput)

		if err != nil {
			output.FailedRows++
		} else if rowOutput.Success {
			output.SuccessfulRows++
		} else {
			output.FailedRows++
		}
	}

	event.Completed = output.SuccessfulRows
	event.Failed = output.FailedRows

	err = workflow.ExecuteActivity(activityCtx, "CompleteJob", input.JobID).Get(activityCtx, nil)
	if err != nil {
		event.EmitError(ctx, err)
		return nil, err
	}

	output.CompletedAt = workflow.Now(ctx)
	event.EmitSuccess(ctx)

	return output, nil
}
