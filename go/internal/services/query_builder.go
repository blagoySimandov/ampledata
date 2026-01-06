package services

import (
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type QueryBuilder interface {
	Build(entity string) string
}

type QueryBuilderFactory interface {
	Create(columnsMetadata []*models.ColumnMetadata, entityType *string) QueryBuilder
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

type DefaultQueryBuilderFactory struct{}

func NewQueryBuilderFactory() QueryBuilderFactory {
	return &DefaultQueryBuilderFactory{}
}

func (f *DefaultQueryBuilderFactory) Create(columnsMetadata []*models.ColumnMetadata, entityType *string) QueryBuilder {
	return NewQueryBuilder(columnsMetadata, entityType)
}
