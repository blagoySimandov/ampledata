package api

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
)

func (s *Server) ListTemplates(ctx context.Context, req ListTemplatesRequestObject) (ListTemplatesResponseObject, error) {
	authUser, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return ListTemplates401JSONResponse{Message: "Unauthorized"}, nil
	}
	println(authUser.ID)
	templates, err := s.templatesRepo.ListTemplates(ctx, authUser.ID)
	if err != nil {
		return nil, err
	}

	return ListTemplates200JSONResponse(toAPITemplateList(templates)), nil
}
