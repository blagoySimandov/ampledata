package api

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/state"
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
	sourceID, url, err := s.sourcesService.CreateUploadSource(ctx, u.ID, string(req.Body.ContentType))
	if err != nil {
		log.Printf("Failed to create upload source: %v", err)
		return UploadFileForEnrichment500JSONResponse{Message: "Failed to create source"}, nil
	}
	return UploadFileForEnrichment200JSONResponse{Url: url, SourceId: sourceID}, nil
}

func (s *Server) CancelJob(ctx context.Context, req CancelJobRequestObject) (CancelJobResponseObject, error) {
	if err := s.enricher.Cancel(ctx, req.JobID); err != nil {
		return CancelJob404JSONResponse{Message: err.Error()}, nil
	}
	return CancelJob200JSONResponse{Message: "Job cancelled"}, nil
}

func (s *Server) GetJobProgress(ctx context.Context, req GetJobProgressRequestObject) (GetJobProgressResponseObject, error) {
	progress, err := s.enricher.GetProgress(ctx, req.JobID)
	if err != nil {
		return GetJobProgress404JSONResponse{Message: err.Error()}, nil
	}
	return GetJobProgress200JSONResponse(toAPIJobProgress(progress)), nil
}

func (s *Server) GetJobResults(ctx context.Context, req GetJobResultsRequestObject) (GetJobResultsResponseObject, error) {
	offset, limit := 0, 0
	if req.Params.Start != nil {
		offset = *req.Params.Start
	}
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	results, err := s.enricher.GetResults(ctx, req.JobID, offset, limit)
	if err != nil {
		return GetJobResults500JSONResponse{Message: err.Error()}, nil
	}
	apiResults := make([]EnrichmentResult, len(results))
	for i, r := range results {
		apiResults[i] = toAPIEnrichmentResult(r)
	}
	return GetJobResults200JSONResponse(apiResults), nil
}

func (s *Server) GetRowsProgress(ctx context.Context, req GetRowsProgressRequestObject) (GetRowsProgressResponseObject, error) {
	params := parseRowsParams(req.Params)
	response, err := s.enricher.GetRowsProgress(ctx, req.JobID, params)
	if err != nil {
		return GetRowsProgress500JSONResponse{Message: err.Error()}, nil
	}
	return GetRowsProgress200JSONResponse(toAPIRowsProgressResponse(response)), nil
}

func parseRowsParams(p GetRowsProgressParams) state.RowsQueryParams {
	offset, limit := 0, 50
	stage, sort := "all", "updated_at_desc"
	if p.Offset != nil {
		offset = *p.Offset
	}
	if p.Limit != nil {
		limit = *p.Limit
		if limit > 100 {
			limit = 100
		}
		if limit <= 0 {
			limit = 50
		}
	}
	if p.Stage != nil {
		stage = *p.Stage
	}
	if p.Sort != nil {
		sort = *p.Sort
	}
	return state.RowsQueryParams{Offset: offset, Limit: limit, Stage: stage, Sort: sort}
}
