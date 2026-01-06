package services

import (
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type IQueryBuilderFactory interface {
	Create(columnsMetadata []*models.ColumnMetadata, entityType *string) IQueryBuilder
}

type IQueryBuilder interface {
	Build(entity string) []string
}

type PatternQueryBuilder struct {
	patterns        []string
	columnsMetadata []*models.ColumnMetadata
}

func NewPatternQueryBuilder(patterns []string, columnsMetadata []*models.ColumnMetadata) *PatternQueryBuilder {
	return &PatternQueryBuilder{
		patterns:        patterns,
		columnsMetadata: columnsMetadata,
	}
}

func (pqb *PatternQueryBuilder) Build(entity string) []string {
	queries := make([]string, len(pqb.patterns))

	for i, pattern := range pqb.patterns {
		queries[i] = pqb.fillPattern(pattern, entity)
	}

	return queries
}

func (pqb *PatternQueryBuilder) fillPattern(pattern string, entity string) string {
	result := strings.ReplaceAll(pattern, "%entity", entity)
	return result
}
