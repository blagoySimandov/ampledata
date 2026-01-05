package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	userContextKey contextKey = "workos_user"
)

// WorkOSUser represents the authenticated user information
type WorkOSUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware verifies the JWT token from WorkOS and adds user info to the request context
func AuthMiddleware(client *usermanagement.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("Missing Authorization header")
				http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("Invalid Authorization header format")
				http.Error(w, "Unauthorized: Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			accessToken := parts[1]

			// Authenticate the session using WorkOS
			auth, err := client.AuthenticateWithAccessToken(r.Context(), usermanagement.AuthenticateWithAccessTokenOpts{
				AccessToken: accessToken,
			})
			if err != nil {
				log.Printf("Failed to authenticate token: %v", err)
				http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Extract user information
			user := &WorkOSUser{
				ID:        auth.User.ID,
				Email:     auth.User.Email,
				FirstName: auth.User.FirstName,
				LastName:  auth.User.LastName,
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), userContextKey, user)

			// Log successful authentication
			log.Printf("User authenticated: %s (%s)", user.Email, user.ID)

			// Continue with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(ctx context.Context) (*WorkOSUser, bool) {
	user, ok := ctx.Value(userContextKey).(*WorkOSUser)
	return user, ok
}

// GetUserFromRequest retrieves the authenticated user from the request context
// This is a convenience function that wraps GetUserFromContext
func GetUserFromRequest(r *http.Request) (*WorkOSUser, bool) {
	return GetUserFromContext(r.Context())
}

// RequireUser is a helper function that retrieves the user from context
// and writes an error response if the user is not found
func RequireUser(w http.ResponseWriter, r *http.Request) (*WorkOSUser, bool) {
	user, ok := GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized: User not found in context", http.StatusUnauthorized)
		return nil, false
	}
	return user, true
}

// WriteJSON is a helper function to write JSON responses
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
