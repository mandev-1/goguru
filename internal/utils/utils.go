package utils

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

// WriteJSON sends JSON with status code and disables HTML escaping for convenience.
func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

// SchemeFromRequest infers http/https taking proxies into account.
func SchemeFromRequest(r *http.Request) string {
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		return "https"
	}
	return "http"
}

var reSanitize = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

// SanitizeFilename keeps filenames filesystem friendly.
func SanitizeFilename(s string) string {
	return reSanitize.ReplaceAllString(strings.TrimSpace(s), "_")
}

// CurrentUserID resolves the logged-in user id using the session cookie.
func CurrentUserID(r *http.Request, db *sql.DB) (int, error) {
	c, err := r.Cookie("session")
	if err != nil {
		return 0, err
	}
	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", c.Value).Scan(&userID)
	return userID, err
}





































}	return userID, err	err = db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", c.Value).Scan(&userID)	var userID int	}		return 0, err	if err != nil {	c, err := r.Cookie("session")func CurrentUserID(r *http.Request, db *sql.DB) (int, error) {}	return reSanitize.ReplaceAllString(strings.TrimSpace(s), "_")func SanitizeFilename(s string) string {var reSanitize = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)}	return "http"	}		return "https"	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {func SchemeFromRequest(r *http.Request) string {}	enc.Encode(v)	enc.SetEscapeHTML(false)	enc := json.NewEncoder(w)	w.WriteHeader(code)	w.Header().Set("Content-Type", "application/json; charset=utf-8")func WriteJSON(w http.ResponseWriter, code int, v any) {)	"strings"	"regexp"	"net/http"	"encoding/json"	"database/sql"import (package utils