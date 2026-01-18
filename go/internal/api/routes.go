package api

import (
	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/gorilla/mux"
)

func SetupRoutes(enrHandler *EnrichHandler, keySelectorHandler *KeySelectorHandler, jwtVerifier *auth.JWTVerifier) *mux.Router {
	r := mux.NewRouter()

	r.Use(CORSMiddleware().Handler)
	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)
	r.Use(auth.Middleware(jwtVerifier))

	r.HandleFunc("/api/v1/enrichment-signed-url", enrHandler.UploadFileForEnrichment).Methods("POST")
	r.HandleFunc("/api/v1/jobs", enrHandler.ListJobs).Methods("GET")
	r.HandleFunc("/api/v1/jobs/{jobID}/start", enrHandler.StartJob).Methods("POST")
	r.HandleFunc("/api/v1/jobs/{jobID}/progress", enrHandler.GetJobProgress).Methods("GET")
	r.HandleFunc("/api/v1/jobs/{jobID}/cancel", enrHandler.CancelJob).Methods("POST")
	r.HandleFunc("/api/v1/jobs/{jobID}/results", enrHandler.GetJobResults).Methods("GET")

	// Key selection endpoint - separate from enrichment
	r.HandleFunc("/api/v1/select-key", keySelectorHandler.SelectKey).Methods("POST")

	return r
}
