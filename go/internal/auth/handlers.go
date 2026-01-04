package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

type Auth struct {
	workosClient *usermanagement.Client
	clientID     string
	redirectURI  string
	jwks         keyfunc.Keyfunc
}

func NewAuth(apiKey, clientID, redirectURI string) *Auth {
	client := usermanagement.NewClient(apiKey)

	jwksURL, err := client.GetJWKSURL(clientID)
	if err != nil {
		panic("failed to get JWKS URL: " + err.Error())
	}

	jwks, err := keyfunc.NewDefault([]string{jwksURL.String()})
	if err != nil {
		panic("failed to create JWKS keyfunc: " + err.Error())
	}

	return &Auth{
		workosClient: client,
		clientID:     clientID,
		redirectURI:  redirectURI,
		jwks:         jwks,
	}
}

func (a *Auth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	authURL, err := a.workosClient.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		ClientID:    a.clientID,
		RedirectURI: a.redirectURI,
		State:       a.getState(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL.String(), http.StatusTemporaryRedirect)
}

func (a *Auth) getState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (a *Auth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	authResp, err := a.workosClient.AuthenticateWithCode(
		context.Background(),
		usermanagement.AuthenticateWithCodeOpts{
			ClientID: a.clientID,
			Code:     code,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     ACCESS_TOKEN_COOKIE_NAME,
		Value:    authResp.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     REFRESH_TOKEN_COOKIE_NAME,
		Value:    authResp.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, DefaultRedirectAfterLogin, http.StatusTemporaryRedirect)
}

func (a *Auth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(ACCESS_TOKEN_COOKIE_NAME)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	token, _ := jwt.Parse(cookie.Value, nil)
	if token != nil {
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			sessionID, _ := claims["sid"].(string)
			if sessionID != "" {
				logoutURL, _ := a.workosClient.GetLogoutURL(usermanagement.GetLogoutURLOpts{
					SessionID: sessionID,
				})

				http.SetCookie(w, &http.Cookie{
					Name:   ACCESS_TOKEN_COOKIE_NAME,
					MaxAge: -1,
					Path:   "/",
				})

				http.SetCookie(w, &http.Cookie{
					Name:   REFRESH_TOKEN_COOKIE_NAME,
					MaxAge: -1,
					Path:   "/",
				})

				http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
				return
			}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:   ACCESS_TOKEN_COOKIE_NAME,
		MaxAge: -1,
		Path:   "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:   REFRESH_TOKEN_COOKIE_NAME,
		MaxAge: -1,
		Path:   "/",
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
