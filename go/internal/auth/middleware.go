package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

type contextKey string

const (
	userContextKey contextKey = "workos_user"
)

type WorkOSUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func Middleware(client *usermanagement.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("Missing Authorization header")
				http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("Invalid Authorization header format")
				http.Error(w, "Unauthorized: Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			accessToken := parts[1]

			auth, err := client.AuthenticateWithAccessToken(r.Context(), usermanagement.AuthenticateWithAccessTokenOpts{
				AccessToken: accessToken,
			})
			if err != nil {
				log.Printf("Failed to authenticate token: %v", err)
				http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
				return
			}

			user := &WorkOSUser{
				ID:        auth.User.ID,
				Email:     auth.User.Email,
				FirstName: auth.User.FirstName,
				LastName:  auth.User.LastName,
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)

			log.Printf("User authenticated: %s (%s)", user.Email, user.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (*WorkOSUser, bool) {
	user, ok := ctx.Value(userContextKey).(*WorkOSUser)
	return user, ok
}

func GetUserFromRequest(r *http.Request) (*WorkOSUser, bool) {
	return GetUserFromContext(r.Context())
}

func RequireUser(w http.ResponseWriter, r *http.Request) (*WorkOSUser, bool) {
	user, ok := GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized: User not found in context", http.StatusUnauthorized)
		return nil, false
	}
	return user, true
}
