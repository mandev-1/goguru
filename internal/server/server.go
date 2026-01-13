package server

import (
	"camagru/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type Server struct {
	DB *sql.DB
}

func (s *Server) SendJSON(w http.ResponseWriter, status int, resp models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) GetCurrentUser(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.DB.QueryRow(`
		SELECT id, username, email, verified, comment_notifications, created_at
		FROM users
		WHERE session_token = ?
	`, cookie.Value).Scan(
		&user.ID, &user.Username, &user.Email, &user.Verified,
		&user.CommentNotifications, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Server) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.GetCurrentUser(r)
		if err != nil || user == nil || !user.Verified {
			if r.URL.Path == "/editor" {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/") {
				s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Message: "Authentication required",
				})
				return
			}
			http.Redirect(w, r, "/unauthorized", http.StatusFound)
			return
		}
		next(w, r)
	}
}

