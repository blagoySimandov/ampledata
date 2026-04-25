package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/gorilla/mux"
)

type createGoogleSheetsSourceRequest struct {
	SpreadsheetID   string `json:"spreadsheet_id"`
	SpreadsheetURL  string `json:"spreadsheet_url"`
	SpreadsheetName string `json:"spreadsheet_name"`
	SheetName       string `json:"sheet_name"`
}

func (s *Server) HandleCreateGoogleSheetsSource(w http.ResponseWriter, r *http.Request) {
	u, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req createGoogleSheetsSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.SpreadsheetID == "" || req.SheetName == "" {
		http.Error(w, "spreadsheet_id and sheet_name are required", http.StatusBadRequest)
		return
	}
	sourceID, err := s.sourcesService.CreateGoogleSheetsSource(r.Context(), u.ID, req.SpreadsheetID, req.SpreadsheetURL, req.SpreadsheetName, req.SheetName)
	if err != nil {
		log.Printf("Failed to create google sheets source: %v", err)
		http.Error(w, "Failed to create source", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"source_id": sourceID.String()})
}

func (s *Server) HandleListSpreadsheets(w http.ResponseWriter, r *http.Request) {
	u, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	spreadsheets, err := s.sheetsClient.ListSpreadsheets(r.Context(), u.ID)
	if err != nil {
		log.Printf("Failed to list spreadsheets: %v", err)
		http.Error(w, "Failed to list spreadsheets. Make sure Google account is connected.", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, spreadsheets)
}

func (s *Server) HandleListSheetTabs(w http.ResponseWriter, r *http.Request) {
	u, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	spreadsheetID := mux.Vars(r)["spreadsheetId"]
	if spreadsheetID == "" {
		http.Error(w, "spreadsheetId is required", http.StatusBadRequest)
		return
	}
	tabs, err := s.sheetsClient.ListSheetTabs(r.Context(), u.ID, spreadsheetID)
	if err != nil {
		log.Printf("Failed to list sheet tabs: %v", err)
		http.Error(w, "Failed to list sheet tabs", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, tabs)
}
