package templates

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/uptrace/bun"
)

type TemplatesRepo struct {
	db *bun.DB
}

func NewTemplatesRepo(db *bun.DB) *TemplatesRepo {
	return &TemplatesRepo{db: db}
}

func (r *TemplatesRepo) ListTemplates(ctx context.Context, userId string) ([]*models.TemplateDB, error) {
	var templates []*models.TemplateDB

	err := r.db.NewSelect().
		Model(&templates).
		Where("owned_by = ? OR owned_by IS NULL", userId).
		Scan(ctx)
	return templates, err
}
