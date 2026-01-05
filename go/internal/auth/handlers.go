package auth

import (
	"log"
	"net/http"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

func LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := config.GetConfig()

		authorizationURL, err := usermanagement.GetAuthorizationURL(
			usermanagement.GetAuthorizationURLOpts{
				ClientID:    cfg.WorkOSClientID,
				Provider:    "authkit",
				RedirectURI: cfg.WorkOSRedirectURL,
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, authorizationURL.String(), http.StatusSeeOther)
	}
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()

	log.Println("Before opts")
	opts := usermanagement.AuthenticateWithCodeOpts{
		ClientID: cfg.WorkOSClientID,
		Code:     r.URL.Query().Get("code"),
	}

	log.Println("BEFORE AUTH WITH CODE")
	_, err := usermanagement.AuthenticateWithCode(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
	log.Println("Logged in")

	// userProfile := profileAndToken.User
	// accessToken := profileAndToken.AccessToken

	http.Redirect(w, r, cfg.FE_BASE_URL, http.StatusSeeOther)
}
