package api

import (
	"github.com/gorilla/mux"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

func SetupRoutes(enrHandler *EnrichHandler, workosClient *usermanagement.Client) *mux.Router {
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)
	r.Use(AuthMiddleware(workosClient))

	// All routes now require authentication
	r.HandleFunc("/api/v1/enrich", enrHandler.EnrichKeys).Methods("POST")
	r.HandleFunc("/api/v1/jobs/{jobID}/progress", enrHandler.GetJobProgress).Methods("GET")
	r.HandleFunc("/api/v1/jobs/{jobID}/cancel", enrHandler.CancelJob).Methods("POST")
	r.HandleFunc("/api/v1/jobs/{jobID}/results", enrHandler.GetJobResults).Methods("GET")
	r.HandleFunc("/api/v1/enrichment-signed-url", enrHandler.UploadFileForEnrichment).Methods("POST")

	return r
}
