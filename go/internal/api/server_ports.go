package api

import "context"

type ITemplateRepo interface {
	ListTemplates(ctx context.Context) []*Template
}
