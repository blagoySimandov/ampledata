package api

import (
	"context"
	"fmt"
	"log"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *Server) ListSources(ctx context.Context, req ListSourcesRequestObject) (ListSourcesResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return ListSources401JSONResponse{Message: "Unauthorized"}, nil
	}
	offset, limit := 0, 50
	if req.Params.Offset != nil {
		offset = *req.Params.Offset
	}
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	sources, err := s.store.GetSourcesByUser(ctx, u.ID, offset, limit)
	if err != nil {
		log.Printf("Failed to retrieve sources: %v", err)
		return ListSources500JSONResponse{Message: "Failed to retrieve sources"}, nil
	}
	summaries, err := s.buildSourceSummaries(ctx, sources)
	if err != nil {
		return ListSources500JSONResponse{Message: "Failed to build source summaries"}, nil
	}
	return ListSources200JSONResponse{Sources: summaries, TotalCount: len(summaries)}, nil
}

func (s *Server) buildSourceSummaries(ctx context.Context, sources []*models.Source) ([]SourceSummary, error) {
	summaries := make([]SourceSummary, len(sources))
	for i, src := range sources {
		jobs, err := s.store.GetJobsBySource(ctx, src.ID)
		if err != nil {
			return nil, err
		}
		summary := SourceSummary{
			SourceId:  openapi_types.UUID(src.ID),
			Type:      string(src.Type),
			CreatedAt: src.CreatedAt,
			JobCount:  len(jobs),
		}
		if len(jobs) > 0 {
			status := JobStatus(jobs[0].Status)
			summary.LatestJobStatus = &status
		}
		summaries[i] = summary
	}
	return summaries, nil
}

func (s *Server) GetSource(ctx context.Context, req GetSourceRequestObject) (GetSourceResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return GetSource401JSONResponse{Message: "Unauthorized"}, nil
	}
	source, err := s.store.GetSource(ctx, uuid.UUID(req.SourceID))
	if err != nil {
		return GetSource404JSONResponse{Message: "Source not found"}, nil
	}
	if source.UserID != u.ID {
		return GetSource403JSONResponse{Message: "Forbidden"}, nil
	}
	jobs, err := s.store.GetJobsBySource(ctx, source.ID)
	if err != nil {
		return GetSource500JSONResponse{Message: "Failed to retrieve jobs"}, nil
	}
	return GetSource200JSONResponse(toAPISourceDetail(source, jobs)), nil
}

func (s *Server) GetSourceData(ctx context.Context, req GetSourceDataRequestObject) (GetSourceDataResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return GetSourceData401JSONResponse{Message: "Unauthorized"}, nil
	}
	source, err := s.store.GetSource(ctx, uuid.UUID(req.SourceID))
	if err != nil {
		return GetSourceData404JSONResponse{Message: "Source not found"}, nil
	}
	if source.UserID != u.ID {
		return GetSourceData403JSONResponse{Message: "Forbidden"}, nil
	}

	csvMeta, ok := source.Metadata.(*models.CSVSourceMetadata)
	if !ok {
		return GetSourceData500JSONResponse{Message: "Source metadata not found"}, nil
	}

	result, err := s.gcsReader.ReadCSV(ctx, csvMeta.FileURI)
	if err != nil {
		return GetSourceData500JSONResponse{Message: fmt.Sprintf("Failed to read CSV: %v", err)}, nil
	}

	return GetSourceData200JSONResponse{
		Headers: result.Headers,
		Rows:    result.Rows,
	}, nil
}

func (s *Server) EnrichSource(ctx context.Context, req EnrichSourceRequestObject) (EnrichSourceResponseObject, error) {
	authUser, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return EnrichSource401JSONResponse{Message: "Unauthorized"}, nil
	}
	user, ok := user.GetDBUserFromContext(ctx)
	if !ok {
		return EnrichSource500JSONResponse{Message: "User not found"}, nil
	}
	source, err := s.store.GetSource(ctx, uuid.UUID(req.SourceID))
	if err != nil {
		return EnrichSource404JSONResponse{Message: "Source not found"}, nil
	}
	if source.UserID != authUser.ID {
		return EnrichSource403JSONResponse{Message: "Forbidden"}, nil
	}
	keyColumns, keyColumnDescription, err := s.resolveKeyColumns(ctx, source.ID, req.Body)
	if err != nil {
		return EnrichSource400JSONResponse{Message: err.Error()}, nil
	}
	csvMeta, ok := source.Metadata.(*models.CSVSourceMetadata)
	if !ok {
		return EnrichSource500JSONResponse{Message: "Source metadata not found"}, nil
	}
	cols := toModelColumnMetadataSlice(req.Body.ColumnsMetadata)
	rowKeys, err := s.readRowKeys(ctx, csvMeta.FileURI, keyColumns, cols)
	if err != nil {
		return EnrichSource400JSONResponse{Message: fmt.Sprintf("Failed to read CSV: %v", err)}, nil
	}
	if len(rowKeys) == 0 {
		return EnrichSource400JSONResponse{Message: "No rows found in key column"}, nil
	}
	cellsToBeEnriched := int64(len(rowKeys) * len(cols))
	if !user.CanEnrichCells(cellsToBeEnriched) {
		return EnrichSource402JSONResponse{Message: "Insufficient credits to run this job"}, nil
	}
	jobID := generateJobId(".csv") // TODO: Maybe don't append .csv ? idk
	if err := s.store.CreatePendingJob(ctx, jobID, authUser.ID, source.ID); err != nil {
		return EnrichSource500JSONResponse{Message: "Failed to create job"}, nil
	}
	if err := s.configureAndStartEnrich(ctx, jobID, keyColumns, keyColumnDescription, cols, len(rowKeys)); err != nil {
		return EnrichSource500JSONResponse{Message: err.Error()}, nil
	}
	go s.enricher.Enrich(context.Background(), jobID, user.ID, stripeCustomerIDOrEmpty(user), rowKeys, cols, keyColumnDescription)
	return EnrichSource200JSONResponse{JobId: jobID}, nil
}

func (s *Server) resolveKeyColumns(ctx context.Context, sourceID uuid.UUID, body *EnrichSourceJSONRequestBody) ([]string, *string, error) {
	if body.KeyColumns != nil && len(*body.KeyColumns) > 0 {
		return *body.KeyColumns, body.KeyColumnDescription, nil
	}
	jobs, err := s.store.GetJobsBySource(ctx, sourceID)
	if err != nil || len(jobs) == 0 {
		return nil, nil, fmt.Errorf("key_columns required for first enrichment run")
	}
	mostRecent := jobs[0]
	return mostRecent.KeyColumns, mostRecent.KeyColumnDescription, nil
}

func (s *Server) configureAndStartEnrich(ctx context.Context, jobID string, keyColumns []string, keyColumnDescription *string, cols []*models.ColumnMetadata, rowCount int) error {
	if err := s.store.UpdateJobConfiguration(ctx, jobID, keyColumns, cols, keyColumnDescription); err != nil {
		return fmt.Errorf("failed to update job configuration")
	}
	if err := s.store.StartJob(ctx, jobID, rowCount); err != nil {
		return fmt.Errorf("failed to start job")
	}
	return nil
}
