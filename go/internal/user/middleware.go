package user

import (
	"context"
	"log"
	"net/http"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/models"
)

type dbContextKey string

const (
	dbUserContextKey dbContextKey = "db_user"
)

func GetDBUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(dbUserContextKey).(*models.User)
	return user, ok
}

func UserMiddleware(userService Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			workosUser, ok := auth.GetUserFromRequest(r)
			if !ok {
				http.Error(w, "Unauthorized: User not found in context", http.StatusUnauthorized)
				return
			}

			dbUser, err := userService.GetOrCreate(
				r.Context(),
				workosUser.ID,
				workosUser.Email,
				workosUser.FirstName,
				workosUser.LastName,
			)
			if err != nil {
				// TODO: better log
				log.Printf("Failed to get or create user: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), dbUserContextKey, dbUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
