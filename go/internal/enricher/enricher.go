package enricher

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type EnrichmentRunner interface {
	Run(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error
}

type Enricher struct {
	runner       EnrichmentRunner
	stateManager *state.StateManager
}

func NewEnricher(runner EnrichmentRunner, stateManager *state.StateManager) *Enricher {
	return &Enricher{
		runner:       runner,
		stateManager: stateManager,
	}
}

func (e *Enricher) Enrich(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error {
	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	e.stateManager.RegisterCancelFunc(jobID, cancel)

	return e.runner.Run(jobCtx, jobID, rowKeys, columnsMetadata)
}

func (e *Enricher) GetProgress(ctx context.Context, jobID string) (*models.JobProgress, error) {
	return e.stateManager.Progress(ctx, jobID)
}

func (e *Enricher) Cancel(ctx context.Context, jobID string) error {
	return e.stateManager.Cancel(ctx, jobID)
}

func (e *Enricher) GetResults(ctx context.Context, jobID string, offset, limit int) ([]*models.EnrichmentResult, error) {
	completedRows, err := e.stateManager.Store().GetRowsAtStage(ctx, jobID, models.StageCompleted, offset, limit)
	if err != nil {
		return nil, err
	}

	results := make([]*models.EnrichmentResult, len(completedRows))
	for i, row := range completedRows {
		sources := []string{}
		if row.CrawlResults != nil {
			sources = row.CrawlResults.Sources
		}

		results[i] = &models.EnrichmentResult{
			Key:           row.Key,
			ExtractedData: row.ExtractedData,
			Confidence:    row.Confidence,
			Sources:       sources,
			Error:         row.Error,
		}
	}

	return results, nil
}
