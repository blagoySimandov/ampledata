package api

import (
	"context"
	"fmt"
	"log"
	"mime"
	"slices"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/blagoySimandov/ampledata/go/internal/user"
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
	return s.createJobWithSignedURL(ctx, u.ID, string(req.Body.ContentType))
}

func (s *Server) createJobWithSignedURL(ctx context.Context, userID, contentType string) (UploadFileForEnrichmentResponseObject, error) {
	ext, _ := mime.ExtensionsByType(contentType)
	extension := ".csv"
	if len(ext) > 0 {
		extension = ext[0]
	}
	jobID := generateJobId(extension)
	if err := s.store.CreatePendingJob(ctx, jobID, userID, jobID); err != nil {
		log.Printf("Failed to create pending job: %v", err)
		return UploadFileForEnrichment500JSONResponse{Message: "Failed to create job"}, nil
	}
	url, err := generateSignedURL(jobID, contentType)
	if err != nil {
		return UploadFileForEnrichment500JSONResponse{Message: err.Error()}, nil
	}
	return UploadFileForEnrichment200JSONResponse{Url: url, JobId: jobID}, nil
}

func (s *Server) ListJobs(ctx context.Context, req ListJobsRequestObject) (ListJobsResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return ListJobs401JSONResponse{Message: "Unauthorized"}, nil
	}
	offset, limit := 0, 50
	if req.Params.Offset != nil {
		offset = *req.Params.Offset
	}
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	jobs, err := s.store.GetJobsByUser(ctx, u.ID, offset, limit)
	if err != nil {
		log.Printf("Failed to retrieve jobs: %v", err)
		return ListJobs500JSONResponse{Message: "Failed to retrieve jobs"}, nil
	}
	summaries := make([]JobSummary, len(jobs))
	for i, job := range jobs {
		summaries[i] = toAPIJobSummary(models.ToJobSummary(job))
	}
	return ListJobs200JSONResponse{Jobs: summaries, TotalCount: len(summaries)}, nil
}

func (s *Server) StartJob(ctx context.Context, req StartJobRequestObject) (StartJobResponseObject, error) {
	authUser, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return StartJob401JSONResponse{Message: "Unauthorized"}, nil
	}
	dbUser, ok := user.GetDBUserFromContext(ctx)
	if !ok {
		return StartJob500JSONResponse{Message: "User not found"}, nil
	}
	job, errResp := s.validateJobForStart(ctx, req.JobID, authUser.ID)
	if errResp != nil {
		return errResp, nil
	}
	return s.executeStartJob(ctx, job, dbUser, req.Body)
}

func (s *Server) validateJobForStart(ctx context.Context, jobID, userID string) (*models.Job, StartJobResponseObject) {
	job, err := s.store.GetJob(ctx, jobID)
	if err != nil {
		return nil, StartJob404JSONResponse{Message: "Job not found"}
	}
	if job.UserID != userID {
		return nil, StartJob403JSONResponse{Message: "Forbidden: You do not own this job"}
	}
	if job.Status != models.JobStatusPending {
		return nil, StartJob400JSONResponse{Message: fmt.Sprintf("Job cannot be started: current status is %s", job.Status)}
	}
	return job, nil
}

func (s *Server) executeStartJob(ctx context.Context, job *models.Job, dbUser *models.User, body *StartJobJSONRequestBody) (StartJobResponseObject, error) {
	if dbUser.SubscriptionTier == nil {
		return StartJob402JSONResponse{Message: "Active subscription required"}, nil
	}
	cols := toModelColumnMetadataSlice(body.ColumnsMetadata)
	rowKeys, err := s.readRowKeys(ctx, job.FilePath, body.KeyColumns, cols)
	if err != nil {
		return StartJob400JSONResponse{Message: fmt.Sprintf("Failed to read CSV file: %v", err)}, nil
	}
	if len(rowKeys) == 0 {
		return StartJob400JSONResponse{Message: "No rows found in the specified key column"}, nil
	}
	if err := s.configureAndStartJob(ctx, job.JobID, body, cols, len(rowKeys)); err != nil {
		return StartJob500JSONResponse{Message: err.Error()}, nil
	}
	go s.enricher.Enrich(context.Background(), job.JobID, dbUser.ID, stripeCustomerIDOrEmpty(dbUser), rowKeys, cols)
	return StartJob200JSONResponse{JobId: job.JobID, Message: fmt.Sprintf("Enrichment started with %d rows", len(rowKeys))}, nil
}

func (s *Server) configureAndStartJob(ctx context.Context, jobID string, body *StartJobJSONRequestBody, cols []*models.ColumnMetadata, rowCount int) error {
	if err := s.store.UpdateJobConfiguration(ctx, jobID, body.KeyColumns, cols, body.EntityType); err != nil {
		log.Printf("Failed to update job configuration: %v", err)
		return fmt.Errorf("failed to update job configuration")
	}
	if err := s.store.StartJob(ctx, jobID, rowCount); err != nil {
		log.Printf("Failed to start job: %v", err)
		return fmt.Errorf("failed to start job")
	}
	return nil
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
	if len(imputationCols) > 0 {
		return s.gcsReader.ReadCompositeKeyFromFileFiltered(ctx, filePath, keyColumns, imputationCols)
	}
	return s.gcsReader.ReadCompositeKeyFromFile(ctx, filePath, keyColumns)
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

func stripeCustomerIDOrEmpty(u *models.User) string {
	if u.StripeCustomerID != nil {
		return *u.StripeCustomerID
	}
	return ""
}
