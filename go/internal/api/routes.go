package api

import (
	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/gorilla/mux"
)

func SetupRoutes(enrHandler *EnrichHandler, keySelectorHandler *KeySelectorHandler, jwtVerifier *auth.JWTVerifier, userService user.Service, billing services.BillingService, userRepo user.Repository) *mux.Router {
	r := mux.NewRouter()

	r.Use(CORSMiddleware().Handler)
	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)

	checkoutHandler := NewCheckoutHandler(billing, userRepo)
	r.HandleFunc("/api/v1/webhooks/stripe", checkoutHandler.HandleWebhook).Methods("POST")

	authenticated := r.PathPrefix("/api/v1").Subrouter()
	authenticated.Use(auth.Middleware(jwtVerifier))
	authenticated.Use(user.UserMiddleware(userService))

	authenticated.HandleFunc("/enrichment-signed-url", enrHandler.UploadFileForEnrichment).Methods("POST", "OPTIONS")
	authenticated.HandleFunc("/jobs", enrHandler.ListJobs).Methods("GET", "OPTIONS")
	authenticated.HandleFunc("/jobs/{jobID}/start", enrHandler.StartJob).Methods("POST", "OPTIONS")
	authenticated.HandleFunc("/jobs/{jobID}/cancel", enrHandler.CancelJob).Methods("POST", "OPTIONS")

	authenticated.HandleFunc("/jobs/{jobID}/progress", enrHandler.GetJobProgress).Methods("GET", "OPTIONS")
	authenticated.HandleFunc("/jobs/{jobID}/results", enrHandler.GetJobResults).Methods("GET", "OPTIONS")
	authenticated.HandleFunc("/jobs/{jobID}/rows", enrHandler.GetRowsProgress).Methods("GET", "OPTIONS")

	authenticated.HandleFunc("/select-key", keySelectorHandler.SelectKey).Methods("POST", "OPTIONS")

	authenticated.HandleFunc("/checkout", checkoutHandler.CreateCheckoutSession).Methods("POST", "OPTIONS")
	authenticated.HandleFunc("/credits/balance", checkoutHandler.GetCreditBalance).Methods("GET", "OPTIONS")

	return r
}
