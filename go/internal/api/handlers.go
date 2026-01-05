package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"slices"

	"github.com/blagoySimandov/ampledata/go/internal/enricher"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type EnrichHandler struct {
	enricher *enricher.Enricher
}

func NewEnrichHandler(enr *enricher.Enricher) *EnrichHandler {
	return &EnrichHandler{
		enricher: enr,
	}
}

func (h *EnrichHandler) EnrichKeys(w http.ResponseWriter, r *http.Request) {
	// Example: Get the authenticated user from the request context
	user, ok := GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.EnrichmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobID := uuid.New().String()

	// You can now use the user information for logging, auditing, etc.
	log.Printf("User %s (%s) started enrichment job %s", user.Email, user.ID, jobID)

	go h.enricher.Enrich(context.Background(), jobID, req.RowKeys, req.ColumnsMetadata)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.EnrichmentResponse{
		JobID:   jobID,
		Message: "Enrichment started",
	})
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
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	id := generateJobId(ext[0])
	url, err := generateSignedURL(id, reqBody.ContentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := SignedURLResponse{
		URL:   url,
		JobID: id,
	}
	json.NewEncoder(w).Encode(response)
	return
}
