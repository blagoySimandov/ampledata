package services

import (
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type QueryBuilder interface {
	Build(entity string) string
}

type DefaultQueryBuilder struct {
	columnsMetadata []*models.ColumnMetadata
	entityType      *string
}

func NewQueryBuilder(columnsMetadata []*models.ColumnMetadata, entityType *string) *DefaultQueryBuilder {
	return &DefaultQueryBuilder{
		columnsMetadata: columnsMetadata,
		entityType:      entityType,
	}
}

func (qb *DefaultQueryBuilder) Build(entity string) string {
	parts := []string{entity}

	if qb.entityType != nil && *qb.entityType != "" {
		parts = append(parts, *qb.entityType)
	}

	for _, col := range qb.columnsMetadata {
		parts = append(parts, col.Name)
	}

	return strings.Join(parts, " ")
}
