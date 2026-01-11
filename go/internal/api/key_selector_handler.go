package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
)

type KeySelectorHandler struct {
	keySelector services.KeySelector
	gcsReader   *gcs.CSVReader
	store       state.Store
}

func NewKeySelectorHandler(keySelector services.KeySelector, gcsReader *gcs.CSVReader, store state.Store) *KeySelectorHandler {
	return &KeySelectorHandler{
		keySelector: keySelector,
		gcsReader:   gcsReader,
		store:       store,
	}
}

func (h *KeySelectorHandler) SelectKey(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.SelectKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.JobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	// Verify the job exists and belongs to the user
	job, err := h.store.GetJob(r.Context(), req.JobID)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if job.UserID != user.ID {
		http.Error(w, "Forbidden: You do not own this job", http.StatusForbidden)
		return
	}

	// Read the CSV file to get headers
	csvResult, err := h.gcsReader.ReadCSV(r.Context(), job.FilePath)
	if err != nil {
		log.Printf("Failed to read CSV file: %v", err)
		http.Error(w, "Failed to read CSV file", http.StatusInternalServerError)
		return
	}

	if len(csvResult.Headers) == 0 {
		http.Error(w, "No headers found in CSV file", http.StatusBadRequest)
		return
	}

	// Use the key selector service to select the best key
	result, err := h.keySelector.SelectBestKey(r.Context(), csvResult.Headers, req.ColumnsMetadata)
	if err != nil {
		log.Printf("Failed to select key: %v", err)
		http.Error(w, "Failed to select key", http.StatusInternalServerError)
		return
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	response := models.SelectKeyResponse{
		SelectedKey: result.SelectedKey,
		AllKeys:     result.AllKeys,
		Reasoning:   result.Reasoning,
	}
	json.NewEncoder(w).Encode(response)
}
