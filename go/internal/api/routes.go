package api

import (
	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/gorilla/mux"
)

func SetupRoutes(enrHandler *EnrichHandler, keySelectorHandler *KeySelectorHandler, jwtVerifier *auth.JWTVerifier, userService user.Service) *mux.Router {
	r := mux.NewRouter()

	r.Use(CORSMiddleware().Handler)
	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)
	r.Use(auth.Middleware(jwtVerifier, userService))

	r.HandleFunc("/api/v1/enrichment-signed-url", enrHandler.UploadFileForEnrichment).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/jobs", enrHandler.ListJobs).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/jobs/{jobID}/start", enrHandler.StartJob).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/jobs/{jobID}/progress", enrHandler.GetJobProgress).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/jobs/{jobID}/cancel", enrHandler.CancelJob).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/jobs/{jobID}/results", enrHandler.GetJobResults).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/jobs/{jobID}/rows", enrHandler.GetRowsProgress).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/v1/select-key", keySelectorHandler.SelectKey).Methods("POST", "OPTIONS")

	return r
}
