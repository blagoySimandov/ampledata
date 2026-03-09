//go:generate oapi-codegen --config=oapi-codegen.yaml openapi.yaml
package api

import (
	"net/http"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/user"
	"github.com/gorilla/mux"
)

func SetupRoutes(server *Server, jwtVerifier *auth.JWTVerifier, userService user.Service) *mux.Router {
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

	publicRouter := mux.NewRouter()
	publicRouter.HandleFunc("/openapi.json", serveOpenAPISpec).Methods("GET")
	publicRouter.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", swaggerUIHandler()))

	// (auth required)
	protectedRouter := mux.NewRouter()
	protectedRouter.Use(auth.Middleware(jwtVerifier))
	protectedRouter.Use(user.UserMiddleware(userService))
	HandlerFromMuxWithBaseURL(strictHandler, protectedRouter, "/api/v1")

	mainRouter.PathPrefix("/api/v1").Handler(protectedRouter)
	mainRouter.PathPrefix("/").Handler(publicRouter)

	return mainRouter
}
