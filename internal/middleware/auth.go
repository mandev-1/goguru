package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/yourusername/camagru/internal/models"
	"github.com/yourusername/camagru/internal/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

type AuthMiddleware struct {
	sessionRepo *models.SessionRepository
}

func NewAuthMiddleware(sessionRepo *models.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{sessionRepo: sessionRepo}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := m.currentUserID(r)
		if err != nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			} else {
				http.ServeFile(w, r, "web/static/pages/unauthorized.html")
			}
			return
		}

		// Add userID to context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) currentUserID(r *http.Request) (int, error) {
	c, err := r.Cookie("session")
	if err != nil {
		return 0, errors.New("no session")
	}

	userID, err := m.sessionRepo.FindUserID(c.Value)
	if err != nil {
		return 0, errors.New("invalid session")
	}

	return userID, nil
}

// Helper function to get user ID from context
func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}
