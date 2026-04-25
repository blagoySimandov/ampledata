package api

import (
	"log"
	"net/http"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
)

func (s *Server) HandleOAuthInitiate(w http.ResponseWriter, r *http.Request) {
	u, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	url, err := s.oauthService.AuthURL(r.Context(), u.ID)
	if err != nil {
		log.Printf("Failed to generate OAuth URL: %v", err)
		http.Error(w, "Failed to initiate OAuth", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"url": url})
}

func (s *Server) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "Missing code or state", http.StatusBadRequest)
		return
	}
	if _, err := s.oauthService.HandleCallback(r.Context(), code, state); err != nil {
		log.Printf("OAuth callback failed: %v", err)
		http.Redirect(w, r, "/app?google_error=true", http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/app?google_connected=true", http.StatusTemporaryRedirect)
}

func (s *Server) HandleOAuthStatus(w http.ResponseWriter, r *http.Request) {
	u, ok := auth.GetUserFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	connected := s.oauthService.IsConnected(r.Context(), u.ID)
	WriteJSON(w, http.StatusOK, map[string]bool{"connected": connected})
}
