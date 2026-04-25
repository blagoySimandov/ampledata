//go:generate oapi-codegen --config=oapi-codegen.yaml openapi.yaml
package api

import (
	"net/http"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/gorilla/mux"
)

func SetupRoutes(server *Server, jwtVerifier *auth.JWTVerifier, userService user.Service, staticDir string) *mux.Router {
	mainRouter := mux.NewRouter()
	mainRouter.Use(CORSMiddleware().Handler)
	mainRouter.Use(LoggingMiddleware)
	mainRouter.Use(RecoveryMiddleware)

	strictHandler := NewStrictHandler(server, nil)

	webhookWrapper := &ServerInterfaceWrapper{
		Handler: strictHandler,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	}
	mainRouter.HandleFunc("/api/v1/webhooks/stripe", webhookWrapper.HandleStripeWebhook).Methods("POST")
	mainRouter.HandleFunc("/api/v1/oauth/google/callback", server.HandleOAuthCallback).Methods("GET")
	mainRouter.HandleFunc("/openapi.json", serveOpenAPISpec).Methods("GET")
	mainRouter.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", swaggerUIHandler()))

	protectedRouter := mux.NewRouter()
	protectedRouter.Use(auth.Middleware(jwtVerifier))
	protectedRouter.Use(user.UserMiddleware(userService))
	HandlerFromMuxWithBaseURL(strictHandler, protectedRouter, "/api/v1")

	protectedRouter.HandleFunc("/api/v1/oauth/google/initiate", server.HandleOAuthInitiate).Methods("GET")
	protectedRouter.HandleFunc("/api/v1/oauth/google/status", server.HandleOAuthStatus).Methods("GET")
	protectedRouter.HandleFunc("/api/v1/sources/google-sheets", server.HandleCreateGoogleSheetsSource).Methods("POST")
	protectedRouter.HandleFunc("/api/v1/google-sheets/spreadsheets", server.HandleListSpreadsheets).Methods("GET")
	protectedRouter.HandleFunc("/api/v1/google-sheets/{spreadsheetId}/sheets", server.HandleListSheetTabs).Methods("GET")

	mainRouter.PathPrefix("/api/v1").Handler(protectedRouter)
	mainRouter.PathPrefix("/").Handler(newSPAHandler(staticDir))

	return mainRouter
}
