package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
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

type JWTVerifier struct {
	clientID    string
	keySet      jwk.Set
	mu          sync.RWMutex
	lastFetch   time.Time
	debugBypass bool
}

func NewJWTVerifier(clientID string, debugBypass bool) (*JWTVerifier, error) {
	v := &JWTVerifier{
		clientID:    clientID,
		debugBypass: debugBypass,
	}

	if debugBypass {
		log.Printf("DEBUG_AUTH_BYPASS enabled - authentication disabled for development!")
		return v, nil
	}

	if err := v.refreshKeySet(); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return v, nil
}

func (v *JWTVerifier) refreshKeySet() error {
	jwksURL := fmt.Sprintf("https://api.workos.com/sso/jwks/%s", v.clientID)

	keySet, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return err
	}

	v.mu.Lock()
	v.keySet = keySet
	v.lastFetch = time.Now()
	v.mu.Unlock()

	return nil
}

func (v *JWTVerifier) getKeySet() jwk.Set {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if time.Since(v.lastFetch) > 1*time.Hour {
		go v.refreshKeySet()
	}

	return v.keySet
}

func (v *JWTVerifier) VerifyToken(tokenString string) (jwt.Token, error) {
	keySet := v.getKeySet()

	token, err := jwt.ParseString(tokenString, jwt.WithKeySet(keySet), jwt.WithValidate(true))
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return token, nil
}

func Middleware(verifier *JWTVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Debug bypass: inject a fake user and skip all auth checks
			if verifier.debugBypass {
				debugUser := &WorkOSUser{
					ID:        "debug-user-id",
					Email:     "debug@localhost",
					FirstName: "Debug",
					LastName:  "User",
				}
				ctx := context.WithValue(r.Context(), userContextKey, debugUser)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Normal auth flow
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

			token, err := verifier.VerifyToken(accessToken)
			if err != nil {
				log.Printf("Failed to verify token: %v", err)
				http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
				return
			}

			claims := token.PrivateClaims()

			user := &WorkOSUser{
				ID:        getStringClaim(claims, "sid"),
				Email:     getStringClaim(claims, "email"),
				FirstName: getStringClaim(claims, "first_name"),
				LastName:  getStringClaim(claims, "last_name"),
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)

			log.Printf("User authenticated: %s (%s)", user.Email, user.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getStringClaim(claims map[string]interface{}, key string) string {
	if val, ok := claims[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
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
