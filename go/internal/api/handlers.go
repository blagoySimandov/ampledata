package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"slices"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

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
	return s.createSourceWithSignedURL(ctx, u.ID, string(req.Body.ContentType))
}

func (s *Server) createSourceWithSignedURL(ctx context.Context, userID, contentType string) (UploadFileForEnrichmentResponseObject, error) {
	ext, _ := mime.ExtensionsByType(contentType)
	extension := ".csv"
	if len(ext) > 0 {
		extension = ext[0]
	}
	fileID := generateJobId(extension)
	source, err := s.createCSVSource(ctx, userID, fileID, contentType)
	if err != nil {
		log.Printf("Failed to create source: %v", err)
		return UploadFileForEnrichment500JSONResponse{Message: "Failed to create source"}, nil
	}
	url, err := generateSignedURL(fileID, contentType)
	if err != nil {
		return UploadFileForEnrichment500JSONResponse{Message: err.Error()}, nil
	}
	return UploadFileForEnrichment200JSONResponse{Url: url, SourceId: source.ID}, nil
}

func (s *Server) createCSVSource(ctx context.Context, userID, fileURI, contentType string) (*models.SourceDB, error) {
	metaJSON, err := json.Marshal(&models.CSVSourceMetadata{FileURI: fileURI, ContentType: contentType})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	now := time.Now()
	source := &models.SourceDB{
		UserID:    userID,
		Type:      models.SourceTypeCSVUpload,
		Metadata:  json.RawMessage(metaJSON),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateSource(ctx, source); err != nil {
		return nil, err
	}
	return source, nil
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

func (s *Server) readRowKeys(ctx context.Context, filePath string, keyColumns []string, columnsMetadata []*models.ColumnMetadata) ([]string, error) {
	imputationCols := imputationColumnNames(columnsMetadata)
	var keys []string
	var err error
	if len(imputationCols) > 0 {
		keys, err = s.gcsReader.ReadCompositeKeyFromFileFiltered(ctx, filePath, keyColumns, imputationCols)
	} else {
		keys, err = s.gcsReader.ReadCompositeKeyFromFile(ctx, filePath, keyColumns)
	}
	if err != nil {
		return nil, err
	}
	return deduplicateKeys(keys), nil
}

func deduplicateKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			result = append(result, k)
		}
	}
	return result
}

func imputationColumnNames(cols []*models.ColumnMetadata) []string {
	var names []string
	for _, col := range cols {
		if col.JobType == models.JobTypeImputation {
			names = append(names, col.Name)
		}
	}
	return names
}

func sourceCSVMeta(job *models.Job) (*models.CSVSourceMetadata, bool) {
	if job.Source == nil {
		return nil, false
	}
	meta, ok := job.Source.Metadata.(*models.CSVSourceMetadata)
	return meta, ok
}

func stripeCustomerIDOrEmpty(u *models.User) string {
	if u.StripeCustomerID != nil {
		return *u.StripeCustomerID
	}
	return ""
}


