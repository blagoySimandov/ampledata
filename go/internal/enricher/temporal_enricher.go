package enricher

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/workflows"
)

// TemporalEnricher uses Temporal workflows for enrichment processing
type TemporalEnricher struct {
	temporalClient client.Client
	stateManager   *state.StateManager
	taskQueue      string
	maxRetries     int
}

// NewTemporalEnricher creates a new enricher that uses Temporal
func NewTemporalEnricher(temporalClient client.Client, stateManager *state.StateManager, taskQueue string, maxRetries int) *TemporalEnricher {
	return &TemporalEnricher{
		temporalClient: temporalClient,
		stateManager:   stateManager,
		taskQueue:      taskQueue,
		maxRetries:     maxRetries,
	}
}

// Enrich starts a Temporal workflow to process the job
func (e *TemporalEnricher) Enrich(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error {
	// Start the JobWorkflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("job-%s", jobID),
		TaskQueue: e.taskQueue,
	}

	input := workflows.JobWorkflowInput{
		JobID:           jobID,
		RowKeys:         rowKeys,
		ColumnsMetadata: columnsMetadata,
		EntityType:      nil, // Can be added if needed
		MaxRetries:      e.maxRetries,
	}

	workflowRun, err := e.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.JobWorkflow, input)
	if err != nil {
		return fmt.Errorf("failed to start workflow: %w", err)
	}

	// Register the workflow for cancellation
	// We store the workflow ID so we can cancel it later
	e.stateManager.RegisterWorkflowID(jobID, workflowRun.GetID(), workflowRun.GetRunID())

	return nil
}

// GetProgress returns the current progress of a job
func (e *TemporalEnricher) GetProgress(ctx context.Context, jobID string) (*models.JobProgress, error) {
	return e.stateManager.Progress(ctx, jobID)
}

// Cancel cancels a running job
func (e *TemporalEnricher) Cancel(ctx context.Context, jobID string) error {
	// Get workflow ID from state manager
	workflowID, runID := e.stateManager.GetWorkflowID(jobID)
	if workflowID == "" {
		return fmt.Errorf("workflow not found for job %s", jobID)
	}

	// Cancel the workflow
	err := e.temporalClient.CancelWorkflow(ctx, workflowID, runID)
	if err != nil {
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	// Update job status in database
	return e.stateManager.Cancel(ctx, jobID)
}

// GetResults returns the enrichment results for a job
func (e *TemporalEnricher) GetResults(ctx context.Context, jobID string, offset, limit int) ([]*models.EnrichmentResult, error) {
	completedRows, err := e.stateManager.Store().GetRowsAtStage(ctx, jobID, models.StageCompleted, offset, limit)
	if err != nil {
		return nil, err
	}

	results := make([]*models.EnrichmentResult, len(completedRows))
	for i, row := range completedRows {
		results[i] = models.ToEnrichmentResult(row)
	}

	return results, nil
}

// WaitForCompletion waits for a job workflow to complete
// This is useful for testing or synchronous processing
func (e *TemporalEnricher) WaitForCompletion(ctx context.Context, jobID string) (*workflows.JobWorkflowOutput, error) {
	workflowID, runID := e.stateManager.GetWorkflowID(jobID)
	if workflowID == "" {
		return nil, fmt.Errorf("workflow not found for job %s", jobID)
	}

	var output workflows.JobWorkflowOutput
	err := e.temporalClient.GetWorkflow(ctx, workflowID, runID).Get(ctx, &output)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow result: %w", err)
	}

	return &output, nil
}
