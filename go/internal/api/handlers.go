package api

import (
	"context"
	"encoding/json"
	"net/http"

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
	var req models.EnrichmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobID := uuid.New().String()

	go h.enricher.EnrichWithMetadata(context.Background(), jobID, req.RowKeys, req.ColumnsMetadata)

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

	results, err := h.enricher.GetResults(r.Context(), jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
