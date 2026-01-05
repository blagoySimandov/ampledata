package auth

import (
	"log"
	"net/http"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/workos/workos-go/v6/pkg/sso"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

const (
	ORG_ID = "org_test_idp" // Will figure out later when and why i need this.
)

func LoginHandler() http.Handler {
	cfg := config.GetConfig()
	return sso.Login(sso.GetAuthorizationURLOpts{
		Organization: ORG_ID,
		RedirectURI:  cfg.WorkOSRedirectURL,
	})
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
