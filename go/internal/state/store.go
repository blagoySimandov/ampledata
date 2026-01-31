package state

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type RowsQueryParams struct {
	Offset int
	Limit  int
	Stage  string
	Sort   string
}

type PaginatedRows struct {
	Rows  []*models.RowState
	Total int
}

type Store interface {
	CreateJob(ctx context.Context, jobID string, totalRows int, status models.JobStatus) error
	CreatePendingJob(ctx context.Context, jobID, userID, filePath string) error
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
	UpdateJobConfiguration(ctx context.Context, jobID string, keyColumns []string, columnsMetadata []*models.ColumnMetadata, entityType *string) error
	StartJob(ctx context.Context, jobID string, totalRows int) error
	GetJobsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Job, error)
	BulkCreateRows(ctx context.Context, jobID string, rowKeys []string) error

	SaveRowState(ctx context.Context, jobID string, state *models.RowState) error
	GetRowState(ctx context.Context, jobID string, key string) (*models.RowState, error)
	GetRowsAtStage(ctx context.Context, jobID string, stage models.RowStage, offset, limit int) ([]*models.RowState, error)
	GetRowsPaginated(ctx context.Context, jobID string, params RowsQueryParams) (*PaginatedRows, error)

	SetJobStatus(ctx context.Context, jobID string, status models.JobStatus) error
	GetJobStatus(ctx context.Context, jobID string) (models.JobStatus, error)
	GetJobProgress(ctx context.Context, jobID string) (*models.JobProgress, error)
	IncrementJobCost(ctx context.Context, jobID string, costDollars, costCredits int) error

	Close() error
}
