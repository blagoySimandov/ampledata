package api

import (
	"context"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
)

func (s *Server) GetMe(ctx context.Context, _ GetMeRequestObject) (GetMeResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok || u == nil {
		return GetMe401JSONResponse{Message: "Unauthorized"}, nil
	}
	return GetMe200JSONResponse{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
	}, nil
}
