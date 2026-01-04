package auth

type contextKey string

const (
	ACCESS_TOKEN_COOKIE_NAME  = "accessToken"
	REFRESH_TOKEN_COOKIE_NAME = "refreshToken"

	ContextKeyUserID    contextKey = "user_id"
	ContextKeySessionID contextKey = "session_id"

	DefaultRedirectURI         = "http://localhost:8080/callback"
	DefaultRedirectAfterLogin  = "/"
)
