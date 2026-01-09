package enricher

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

// IEnricher defines the interface for enrichment operations
// This interface is implemented by TemporalEnricher
type IEnricher interface {
	Enrich(ctx context.Context, jobID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata) error
	GetProgress(ctx context.Context, jobID string) (*models.JobProgress, error)
	Cancel(ctx context.Context, jobID string) error
	GetResults(ctx context.Context, jobID string, offset, limit int) ([]*models.EnrichmentResult, error)
}
