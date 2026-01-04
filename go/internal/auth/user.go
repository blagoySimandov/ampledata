package auth

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type contextKey string

const UserContextKey contextKey = "user"
