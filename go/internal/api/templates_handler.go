package api

import "context"

func (s *Server) ListTemplates(ctx context.Context, req ListTemplatesRequestObject) (ListTemplatesResponseObject, error) {
	return ListTemplates200JSONResponse(toAPITemplateList(s.templatesRepo.ListTemplates(ctx))), nil
}
