package middleware

import (
	"database/sql"
	"goguru/internal/utils"
	"net/http"
	"strings"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := utils.CurrentUserID(r, db); err != nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				utils.WriteJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Unauthorized"})
			} else {
				http.ServeFile(w, r, "web/static/pages/unauthorized.html")
			}
			return
		}
		next(w, r)
	}
}
