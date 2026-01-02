package state

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type Store interface {
	CreateJob(ctx context.Context, jobID string, totalRows int, status models.JobStatus) error
	BulkCreateRows(ctx context.Context, jobID string, rowKeys []string) error

	SaveRowState(ctx context.Context, jobID string, state *models.RowState) error
	GetRowState(ctx context.Context, jobID string, key string) (*models.RowState, error)
	GetRowsAtStage(ctx context.Context, jobID string, stage models.RowStage, offset, limit int) ([]*models.RowState, error)

	SetJobStatus(ctx context.Context, jobID string, status models.JobStatus) error
	GetJobStatus(ctx context.Context, jobID string) (models.JobStatus, error)
	GetJobProgress(ctx context.Context, jobID string) (*models.JobProgress, error)

	Close() error
}
