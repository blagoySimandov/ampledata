package api

import (
	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/gorilla/mux"
)

func SetupRoutes(enrHandler *EnrichHandler, authHandler *auth.Auth) *mux.Router {
	r := mux.NewRouter()

	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)

	r.HandleFunc("/login", authHandler.HandleLogin).Methods("GET")
	r.HandleFunc("/callback", authHandler.HandleCallback).Methods("GET")
	r.HandleFunc("/logout", authHandler.HandleLogout).Methods("GET")

	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(authHandler.AuthMiddleware)

	apiRouter.HandleFunc("/enrich", enrHandler.EnrichKeys).Methods("POST")
	apiRouter.HandleFunc("/jobs/{jobID}/progress", enrHandler.GetJobProgress).Methods("GET")
	apiRouter.HandleFunc("/jobs/{jobID}/cancel", enrHandler.CancelJob).Methods("POST")
	apiRouter.HandleFunc("/jobs/{jobID}/results", enrHandler.GetJobResults).Methods("GET")
	apiRouter.HandleFunc("/enrichment-signed-url", enrHandler.UploadFileForEnrichment).Methods("POST")

	return r
}
