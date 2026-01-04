package api

import (
	"encoding/json"
	"net/http"

	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

const (
	callbackErrorMessage      = "Authentication failed"
	authenticateErrorMessage  = "Failed to authenticate"
	userInfoErrorMessage      = "Failed to get user info"
	refreshTokenErrorMessage  = "Failed to refresh token"
	invalidRequestMessage     = "Invalid request"
)

type AuthHandler struct {
	client       *usermanagement.Client
	clientID     string
	redirectURI  string
}

func NewAuthHandler(apiKey, clientID, redirectURI string) *AuthHandler {
	client := usermanagement.NewClient(apiKey)
	return &AuthHandler{
		client:      client,
		clientID:    clientID,
		redirectURI: redirectURI,
	}
}

type AuthenticateRequest struct {
	Code string `json:"code"`
}

type AuthenticateResponse struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         interface{} `json:"user"`
}

func (h *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req AuthenticateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, invalidRequestMessage, http.StatusBadRequest)
		return
	}

	authResp, err := h.client.AuthenticateWithCode(r.Context(), usermanagement.AuthenticateWithCodeOpts{
		Code:       req.Code,
		ClientID:   h.clientID,
	})
	if err != nil {
		http.Error(w, authenticateErrorMessage, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthenticateResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		User:         authResp.User,
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, invalidRequestMessage, http.StatusBadRequest)
		return
	}

	authResp, err := h.client.RefreshSession(r.Context(), usermanagement.RefreshSessionOpts{
		RefreshToken: req.RefreshToken,
		ClientID:     h.clientID,
	})
	if err != nil {
		http.Error(w, refreshTokenErrorMessage, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RefreshTokenResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
	})
}
