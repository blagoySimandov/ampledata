package auth

import (
	"errors"
	"fmt"
	"sync"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	workosJWKSURLTemplate = "https://api.workos.com/sso/jwks/%s"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrMissingClaims = errors.New("missing required claims")
)

type JWTVerifier struct {
	jwks   *keyfunc.JWKS
	mu     sync.RWMutex
}

func NewJWTVerifier(clientID string) (*JWTVerifier, error) {
	jwksURL := fmt.Sprintf(workosJWKSURLTemplate, clientID)

	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	return &JWTVerifier{
		jwks: jwks,
	}, nil
}

func (v *JWTVerifier) VerifyToken(tokenString string) (*User, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	token, err := jwt.Parse(tokenString, v.jwks.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrMissingClaims
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("%w: missing sub claim", ErrMissingClaims)
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("%w: missing email claim", ErrMissingClaims)
	}

	return &User{
		ID:    userID,
		Email: email,
	}, nil
}

func (v *JWTVerifier) Close() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.jwks != nil {
		v.jwks.EndBackground()
	}
}
