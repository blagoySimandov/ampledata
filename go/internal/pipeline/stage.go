package pipeline

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type Message struct {
	JobID           string
	RowKey          string
	State           *models.RowState
	ColumnsMetadata []*models.ColumnMetadata
	QueryPatterns   []string
	Error           error
}

type Stage interface {
	Run(ctx context.Context, inChan <-chan Message, outChan chan<- Message)
	Name() string
}
