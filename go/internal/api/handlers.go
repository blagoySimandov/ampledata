package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"slices"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/enricher"
	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/gorilla/mux"
)

type EnrichHandler struct {
	enricher  *enricher.Enricher
	gcsReader *gcs.CSVReader
	store     state.Store
}

func NewEnrichHandler(enr *enricher.Enricher, gcsReader *gcs.CSVReader, store state.Store) *EnrichHandler {
	return &EnrichHandler{
		enricher:  enr,
		gcsReader: gcsReader,
		store:     store,
	}
}

func (h *EnrichHandler) GetJobProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	progress, err := h.enricher.GetProgress(r.Context(), jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

func (h *EnrichHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	if err := h.enricher.Cancel(r.Context(), jobID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Job cancelled"})
}

func (h *EnrichHandler) GetJobResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	offset := 0
	limit := 0

	if offsetStr := r.URL.Query().Get("start"); offsetStr != "" {
		_, err := fmt.Sscanf(offsetStr, "%d", &offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		_, err := fmt.Sscanf(limitStr, "%d", &limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	results, err := h.enricher.GetResults(r.Context(), jobID, offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *EnrichHandler) UploadFileForEnrichment(w http.ResponseWriter, r *http.Request) {
	var reqBody SignedURLRequest
	user, ok := auth.GetUserFromRequest(r)
	if user == nil || !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !slices.Contains(WHITELISTED_CONTENT_TYPES, reqBody.ContentType) {
		http.Error(w, fmt.Errorf("invalid content type: %s", reqBody.ContentType).Error(), http.StatusBadRequest)
		return
	}

	if reqBody.Length <= 0 {
		http.Error(w, "invalid length", http.StatusBadRequest)
		return
	}

	ext, _ := mime.ExtensionsByType(reqBody.ContentType)
	jobID := generateJobId(ext[0])

	if err := h.store.CreatePendingJob(r.Context(), jobID, user.ID, jobID); err != nil {
		log.Printf("Failed to create pending job: %v", err)
		http.Error(w, "Failed to create job", http.StatusInternalServerError)
		return
	}

	url, err := generateSignedURL(jobID, reqBody.ContentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("User %s (%s) created pending job %s", user.Email, user.ID, jobID)

	w.Header().Set("Content-Type", "application/json")
	response := SignedURLResponse{
		URL:   url,
		JobID: jobID,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *EnrichHandler) StartJob(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	jobID := vars["jobID"]

	job, err := h.store.GetJob(r.Context(), jobID)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if job.UserID != user.ID {
		http.Error(w, "Forbidden: You do not own this job", http.StatusForbidden)
		return
	}

	if job.Status != models.JobStatusPending {
		http.Error(w, fmt.Sprintf("Job cannot be started: current status is %s", job.Status), http.StatusBadRequest)
		return
	}

	var req models.StartJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.KeyColumn == "" {
		http.Error(w, "key_column is required", http.StatusBadRequest)
		return
	}

	if len(req.ColumnsMetadata) == 0 {
		http.Error(w, "columns_metadata is required", http.StatusBadRequest)
		return
	}

	rowKeys, err := h.gcsReader.ReadColumnFromFile(r.Context(), job.FilePath, req.KeyColumn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read CSV file: %v", err), http.StatusBadRequest)
		return
	}

	if len(rowKeys) == 0 {
		http.Error(w, "No rows found in the specified key column", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateJobConfiguration(r.Context(), jobID, req.KeyColumn, req.ColumnsMetadata, req.EntityType); err != nil {
		log.Printf("Failed to update job configuration: %v", err)
		http.Error(w, "Failed to update job configuration", http.StatusInternalServerError)
		return
	}

	if err := h.store.StartJob(r.Context(), jobID, len(rowKeys)); err != nil {
		log.Printf("Failed to start job: %v", err)
		http.Error(w, "Failed to start job", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s (%s) started enrichment job %s with %d rows", user.Email, user.ID, jobID, len(rowKeys))

	go h.enricher.Enrich(context.Background(), jobID, rowKeys, req.ColumnsMetadata)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.StartJobResponse{
		JobID:   jobID,
		Message: fmt.Sprintf("Enrichment started with %d rows", len(rowKeys)),
	})
}

func (h *EnrichHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	offset := 0
	limit := 50

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		fmt.Sscanf(offsetStr, "%d", &offset)
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	jobs, err := h.store.GetJobsByUser(r.Context(), user.ID, offset, limit)
	if err != nil {
		log.Printf("Failed to retrieve jobs: %v", err)
		http.Error(w, "Failed to retrieve jobs", http.StatusInternalServerError)
		return
	}

	summaries := make([]*models.JobSummary, len(jobs))
	for i, job := range jobs {
		summaries[i] = &models.JobSummary{
			JobID:     job.JobID,
			Status:    job.Status,
			TotalRows: job.TotalRows,
			FilePath:  job.FilePath,
			CreatedAt: job.CreatedAt,
			StartedAt: job.StartedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.JobListResponse{
		Jobs:       summaries,
		TotalCount: len(summaries),
	})
}
