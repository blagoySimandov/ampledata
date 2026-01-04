package api

import (
	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/gorilla/mux"
)

func SetupRoutes(enrHandler *EnrichHandler, authHandler *AuthHandler, authMiddleware *auth.Middleware) *mux.Router {
	r := mux.NewRouter()

	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)
	r.Use(CORSMiddleware)

	r.HandleFunc("/api/auth/authenticate", authHandler.Authenticate).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/refresh", authHandler.RefreshToken).Methods("POST", "OPTIONS")

	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/enrich", enrHandler.EnrichKeys).Methods("POST")
	protected.HandleFunc("/jobs/{jobID}/progress", enrHandler.GetJobProgress).Methods("GET")
	protected.HandleFunc("/jobs/{jobID}/cancel", enrHandler.CancelJob).Methods("POST")
	protected.HandleFunc("/jobs/{jobID}/results", enrHandler.GetJobResults).Methods("GET")
	protected.HandleFunc("/enrichment-signed-url", enrHandler.UploadFileForEnrichment).Methods("POST")

	return r
}
