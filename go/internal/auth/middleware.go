package auth

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

func (a *Auth) validateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, a.jwks.Keyfunc)
}

func (a *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(ACCESS_TOKEN_COOKIE_NAME)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "No access token provided")
			return
		}

		token, err := a.validateToken(cookie.Value)
		if err != nil {
			refreshCookie, refreshErr := r.Cookie(REFRESH_TOKEN_COOKIE_NAME)
			if refreshErr == nil {
				authResp, refreshTokenErr := a.workosClient.AuthenticateWithRefreshToken(
					context.Background(),
					usermanagement.AuthenticateWithRefreshTokenOpts{
						ClientID:     a.clientID,
						RefreshToken: refreshCookie.Value,
					},
				)

				if refreshTokenErr == nil {
					http.SetCookie(w, &http.Cookie{
						Name:     ACCESS_TOKEN_COOKIE_NAME,
						Value:    authResp.AccessToken,
						Path:     "/",
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteLaxMode,
					})

					if authResp.RefreshToken != "" {
						http.SetCookie(w, &http.Cookie{
							Name:     REFRESH_TOKEN_COOKIE_NAME,
							Value:    authResp.RefreshToken,
							Path:     "/",
							HttpOnly: true,
							Secure:   true,
							SameSite: http.SameSiteLaxMode,
						})
					}

					token, err = a.validateToken(authResp.AccessToken)
					if err != nil {
						writeJSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token validation failed after refresh")
						return
					}
				} else {
					writeJSONError(w, http.StatusUnauthorized, "REFRESH_FAILED", "Failed to refresh token")
					return
				}
			} else {
				writeJSONError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Access token expired and no refresh token available")
				return
			}
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token claims")
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims["sub"])
		ctx = context.WithValue(ctx, ContextKeySessionID, claims["sid"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
