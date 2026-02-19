package enricher

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type IEnricher interface {
	// Enrich starts a workflow to enrich rowKeys with the target columnsMetadata.
	// rowData is an optional map from composite row key to source column values used
	// by the imputation stage. Pass nil when imputation is not needed.
	Enrich(ctx context.Context, jobID, userID, stripeCustomerID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata, rowData map[string]map[string]string) error
	GetProgress(ctx context.Context, jobID string) (*models.JobProgress, error)
	Cancel(ctx context.Context, jobID string) error
	GetResults(ctx context.Context, jobID string, offset, limit int) ([]*models.EnrichmentResult, error)
	GetRowsProgress(ctx context.Context, jobID string, params state.RowsQueryParams) (*models.RowsProgressResponse, error)
}
