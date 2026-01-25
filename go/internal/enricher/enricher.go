package enricher

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type IEnricher interface {
	Enrich(ctx context.Context, jobID, userID, stripeCustomerID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error
	GetProgress(ctx context.Context, jobID string) (*models.JobProgress, error)
	Cancel(ctx context.Context, jobID string) error
	GetResults(ctx context.Context, jobID string, offset, limit int) ([]*models.EnrichmentResult, error)
	GetRowsProgress(ctx context.Context, jobID string, params state.RowsQueryParams) (*models.RowsProgressResponse, error)
}
