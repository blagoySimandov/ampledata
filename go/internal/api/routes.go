package api

import (
	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/gorilla/mux"
)

func SetupRoutes(enrHandler *EnrichHandler, jwtVerifier *auth.JWTVerifier) *mux.Router {
	r := mux.NewRouter()

	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)
	r.Use(auth.Middleware(jwtVerifier))

	r.HandleFunc("/api/v1/enrich", enrHandler.EnrichKeys).Methods("POST")
	r.HandleFunc("/api/v1/jobs/{jobID}/progress", enrHandler.GetJobProgress).Methods("GET")
	r.HandleFunc("/api/v1/jobs/{jobID}/cancel", enrHandler.CancelJob).Methods("POST")
	r.HandleFunc("/api/v1/jobs/{jobID}/results", enrHandler.GetJobResults).Methods("GET")
	r.HandleFunc("/api/v1/enrichment-signed-url", enrHandler.UploadFileForEnrichment).Methods("POST")

	return r
}
