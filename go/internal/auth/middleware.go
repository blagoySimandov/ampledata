package auth

import (
	"context"
	"net/http"
	"strings"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	unauthorizedMessage = "Unauthorized"
	invalidTokenMessage = "Invalid token"
)

type Middleware struct {
	verifier *JWTVerifier
}

func NewMiddleware(verifier *JWTVerifier) *Middleware {
	return &Middleware{
		verifier: verifier,
	}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(authorizationHeader)
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			http.Error(w, unauthorizedMessage, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		user, err := m.verifier.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, invalidTokenMessage, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserContextKey).(*User)
	return user, ok
}
