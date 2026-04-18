package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/google/uuid"
)

var WHITELISTED_CONTENT_TYPES = []SignedURLRequestContentType{
	Textcsv,
	Applicationjson,
}

func (s *Server) UploadFileForEnrichment(ctx context.Context, req UploadFileForEnrichmentRequestObject) (UploadFileForEnrichmentResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok || u == nil {
		return UploadFileForEnrichment401JSONResponse{Message: "Unauthorized"}, nil
	}
	if !slices.Contains(WHITELISTED_CONTENT_TYPES, req.Body.ContentType) {
		return UploadFileForEnrichment400JSONResponse{Message: fmt.Sprintf("invalid content type: %s", req.Body.ContentType)}, nil
	}
	if req.Body.Length <= 0 {
		return UploadFileForEnrichment400JSONResponse{Message: "invalid length"}, nil
	}
	var headers []string
	if req.Body.Headers != nil {
		headers = *req.Body.Headers
	}
	sourceID, url, err := s.sourcesService.CreateUploadSource(ctx, u.ID, string(req.Body.ContentType), headers)
	if err != nil {
		log.Printf("Failed to create upload source: %v", err)
		return UploadFileForEnrichment500JSONResponse{Message: "Failed to create source"}, nil
	}
	return UploadFileForEnrichment200JSONResponse{Url: url, SourceId: sourceID}, nil
}

func (s *Server) CreateSampleSource(ctx context.Context, req CreateSampleSourceRequestObject) (CreateSampleSourceResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok || u == nil {
		return CreateSampleSource401JSONResponse{Message: "Unauthorized"}, nil
	}
	result, err := s.sourcesService.CreateSampleSource(ctx, u.ID)
	if err != nil {
		log.Printf("Failed to create sample source: %v", err)
		return CreateSampleSource500JSONResponse{Message: "Failed to create sample source"}, nil
	}
	detail := toAPISourceDetail(result.Source, result.Jobs)
	return CreateSampleSource200JSONResponse(detail), nil
}

func (s *Server) ListSources(ctx context.Context, req ListSourcesRequestObject) (ListSourcesResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return ListSources401JSONResponse{Message: "Unauthorized"}, nil
	}
	offset, limit := paginationParams(req.Params.Offset, req.Params.Limit, 50)
	results, err := s.sourcesService.ListSources(ctx, u.ID, offset, limit)
	if err != nil {
		log.Printf("Failed to retrieve sources: %v", err)
		return ListSources500JSONResponse{Message: "Failed to retrieve sources"}, nil
	}
	return ListSources200JSONResponse{Sources: toAPISourceSummaries(results), TotalCount: len(results)}, nil
}

func (s *Server) GetSource(ctx context.Context, req GetSourceRequestObject) (GetSourceResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return GetSource401JSONResponse{Message: "Unauthorized"}, nil
	}
	result, err := s.sourcesService.GetSource(ctx, uuid.UUID(req.SourceID), u.ID)
	if err != nil {
		return toGetSourceError(err), nil
	}
	return GetSource200JSONResponse(toAPISourceDetail(result.Source, result.Jobs)), nil
}

func (s *Server) GetSourceData(ctx context.Context, req GetSourceDataRequestObject) (GetSourceDataResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return GetSourceData401JSONResponse{Message: "Unauthorized"}, nil
	}
	result, err := s.sourcesService.GetSourceData(ctx, uuid.UUID(req.SourceID), u.ID)
	if err != nil {
		return toGetSourceDataError(err), nil
	}
	return GetSourceData200JSONResponse{Headers: result.Headers, Rows: result.Rows}, nil
}

func (s *Server) EnrichSource(ctx context.Context, req EnrichSourceRequestObject) (EnrichSourceResponseObject, error) {
	authUser, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return EnrichSource401JSONResponse{Message: "Unauthorized"}, nil
	}
	dbUser, ok := user.GetDBUserFromContext(ctx)
	if !ok {
		return EnrichSource500JSONResponse{Message: "User not found"}, nil
	}
	input := buildEnrichInput(req, authUser.ID, dbUser)
	jobID, err := s.sourcesService.EnrichSource(ctx, input)
	if err != nil {
		return toEnrichSourceError(err), nil
	}
	return EnrichSource200JSONResponse{JobId: jobID}, nil
}

func buildEnrichInput(req EnrichSourceRequestObject, authUserID string, dbUser *models.User) services.EnrichSourceInput {
	var keyColumns []string
	if req.Body.KeyColumns != nil {
		keyColumns = *req.Body.KeyColumns
	}
	return services.EnrichSourceInput{
		SourceID:             uuid.UUID(req.SourceID),
		AuthUserID:           authUserID,
		DBUser:               dbUser,
		KeyColumns:           keyColumns,
		KeyColumnDescription: req.Body.KeyColumnDescription,
		ColumnsMetadata:      toModelColumnMetadataSlice(req.Body.ColumnsMetadata),
	}
}

func paginationParams(offset, limit *int, defaultLimit int) (int, int) {
	o, l := 0, defaultLimit
	if offset != nil {
		o = *offset
	}
	if limit != nil {
		l = *limit
	}
	return o, l
}

func toGetSourceError(err error) GetSourceResponseObject {
	switch {
	case errors.Is(err, services.ErrSourceNotFound):
		return GetSource404JSONResponse{Message: "Source not found"}
	case errors.Is(err, services.ErrSourceForbidden):
		return GetSource403JSONResponse{Message: "Forbidden"}
	default:
		return GetSource500JSONResponse{Message: "Failed to retrieve source"}
	}
}

func toGetSourceDataError(err error) GetSourceDataResponseObject {
	switch {
	case errors.Is(err, services.ErrSourceNotFound):
		return GetSourceData404JSONResponse{Message: "Source not found"}
	case errors.Is(err, services.ErrSourceForbidden):
		return GetSourceData403JSONResponse{Message: "Forbidden"}
	default:
		return GetSourceData500JSONResponse{Message: fmt.Sprintf("Failed to read CSV: %v", err)}
	}
}

func toEnrichSourceError(err error) EnrichSourceResponseObject {
	var validErr services.ValidationError
	switch {
	case errors.Is(err, services.ErrSourceNotFound):
		return EnrichSource404JSONResponse{Message: "Source not found"}
	case errors.Is(err, services.ErrSourceForbidden):
		return EnrichSource403JSONResponse{Message: "Forbidden"}
	case errors.Is(err, services.ErrInsufficientCredits):
		return EnrichSource402JSONResponse{Message: "Insufficient credits to run this job"}
	case errors.As(err, &validErr):
		return EnrichSource400JSONResponse{Message: err.Error()}
	default:
		return EnrichSource500JSONResponse{Message: err.Error()}
	}
}
