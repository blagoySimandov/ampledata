package googleoauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/uptrace/bun"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthTokenDB struct {
	bun.BaseModel `bun:"table:user_oauth_tokens"`

	UserID       string    `bun:"user_id,pk"`
	Provider     string    `bun:"provider,pk"`
	AccessToken  string    `bun:"access_token"`
	RefreshToken string    `bun:"refresh_token"`
	ExpiresAt    time.Time `bun:"expires_at"`
}

type OAuthStateDB struct {
	bun.BaseModel `bun:"table:oauth_states"`

	State     string    `bun:"state,pk"`
	UserID    string    `bun:"user_id"`
	ExpiresAt time.Time `bun:"expires_at"`
}

type Service struct {
	config *oauth2.Config
	db     *bun.DB
}

func NewService(clientID, clientSecret, redirectURL string, db *bun.DB) *Service {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/drive.metadata.readonly",
		},
		Endpoint: google.Endpoint,
	}
	return &Service{config: config, db: db}
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) AuthURL(ctx context.Context, userID string) (string, error) {
	state, err := generateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	stateDB := &OAuthStateDB{State: state, UserID: userID, ExpiresAt: time.Now().Add(10 * time.Minute)}
	if _, err := s.db.NewInsert().Model(stateDB).Exec(ctx); err != nil {
		return "", fmt.Errorf("failed to save oauth state: %w", err)
	}
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

func (s *Service) HandleCallback(ctx context.Context, code, state string) (string, error) {
	userID, err := s.consumeState(ctx, state)
	if err != nil {
		return "", err
	}
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	if err := s.saveToken(ctx, userID, token); err != nil {
		return "", err
	}
	return userID, nil
}

func (s *Service) consumeState(ctx context.Context, state string) (string, error) {
	var stateDB OAuthStateDB
	err := s.db.NewSelect().Model(&stateDB).Where("state = ? AND expires_at > ?", state, time.Now()).Scan(ctx)
	if err != nil {
		return "", fmt.Errorf("invalid or expired oauth state")
	}
	s.db.NewDelete().Model(&stateDB).WherePK().Exec(ctx)
	return stateDB.UserID, nil
}

func (s *Service) saveToken(ctx context.Context, userID string, token *oauth2.Token) error {
	tokenDB := &OAuthTokenDB{
		UserID:       userID,
		Provider:     "google",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
	}
	_, err := s.db.NewInsert().Model(tokenDB).
		On("CONFLICT (user_id, provider) DO UPDATE SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, expires_at = EXCLUDED.expires_at").
		Exec(ctx)
	return err
}

func (s *Service) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	var tokenDB OAuthTokenDB
	err := s.db.NewSelect().Model(&tokenDB).Where("user_id = ? AND provider = 'google'", userID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("google oauth not connected")
	}
	return &oauth2.Token{
		AccessToken:  tokenDB.AccessToken,
		RefreshToken: tokenDB.RefreshToken,
		Expiry:       tokenDB.ExpiresAt,
	}, nil
}

func (s *Service) GetOAuthClient(ctx context.Context, userID string) (*http.Client, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, err
	}
	tokenSource := s.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh google token: %w", err)
	}
	if newToken.AccessToken != token.AccessToken {
		s.saveToken(ctx, userID, newToken)
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(newToken)), nil
}

func (s *Service) IsConnected(ctx context.Context, userID string) bool {
	_, err := s.GetToken(ctx, userID)
	return err == nil
}
