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

// Enrich starts a Temporal workflow to process the job.
// rowData is optional: when non-nil it enables the imputation stage for each row.
func (e *TemporalEnricher) Enrich(ctx context.Context, jobID, userID, stripeCustomerID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata, rowData map[string]map[string]string) error {
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("job-%s", jobID),
		TaskQueue: e.taskQueue,
	}

	input := workflows.JobWorkflowInput{
		JobID:            jobID,
		UserID:           userID,
		StripeCustomerID: stripeCustomerID,
		RowKeys:          rowKeys,
		ColumnsMetadata:  columnsMetadata,
		EntityType:       nil,
		MaxRetries:       e.maxRetries,
		RowData:          rowData,
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

func (e *TemporalEnricher) Cancel(ctx context.Context, jobID string) error {
	// Get workflow ID from state manager
	workflowID, runID := e.stateManager.GetWorkflowID(jobID)
	if workflowID == "" {
		return fmt.Errorf("workflow not found for job %s", jobID)
	}

	err := e.temporalClient.CancelWorkflow(ctx, workflowID, runID)
	if err != nil {
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	return e.stateManager.Cancel(ctx, jobID)
}

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

func (e *TemporalEnricher) GetRowsProgress(ctx context.Context, jobID string, params state.RowsQueryParams) (*models.RowsProgressResponse, error) {
	paginatedRows, err := e.stateManager.Store().GetRowsPaginated(ctx, jobID, params)
	if err != nil {
		return nil, err
	}

	rows := make([]*models.RowProgressItem, len(paginatedRows.Rows))
	for i, row := range paginatedRows.Rows {
		rows[i] = models.ToRowProgressItem(row)
	}

	hasMore := params.Offset+len(rows) < paginatedRows.Total

	return &models.RowsProgressResponse{
		Rows: rows,
		Pagination: &models.PaginationInfo{
			Total:   paginatedRows.Total,
			Offset:  params.Offset,
			Limit:   params.Limit,
			HasMore: hasMore,
		},
	}, nil
}
