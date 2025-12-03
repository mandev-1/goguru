package auth

import (
	"database/sql"
	"fmt"
	"camagru/internal/email"
	"camagru/internal/models"
	"camagru/internal/utils"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

// ResendVerificationHandler triggers re-sending a verify email for a user.
func ResendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	identifier := strings.TrimSpace(r.FormValue("username"))
	if identifier == "" {
		identifier = strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	}
	if identifier == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Username or email required"})
		return
	}

	var id int
	var emailAddress string
	var verified int
	var token string
	err := db.QueryRow("SELECT id, email, verified, verify_token FROM users WHERE username = ? OR email = ?", identifier, identifier).Scan(&id, &emailAddress, &verified, &token)
	if err != nil {
		utils.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "User not found"})
		return
	}
	if verified == 1 {
		utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Already verified"})
		return
	}
	if token == "" {
		// regenerate token if missing
		token, err = models.RandomToken(32)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}
		if _, err = db.Exec("UPDATE users SET verify_token = ? WHERE id = ?", token, id); err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}
	}
	verifyURL := fmt.Sprintf("%s://%s/verify?token=%s", "http", r.Host, token)
	if err := email.SendVerificationEmail(emailAddress, verifyURL); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Failed to send email"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Verification email resent"})
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	res, err := db.Exec("UPDATE users SET verified = 1, verify_token = NULL WHERE verify_token = ?", token)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}
	// On success, redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/static/pages/password.html")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	emailAddress := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	var uid int
	if err := db.QueryRow("SELECT id FROM users WHERE email = ?", emailAddress).Scan(&uid); err != nil {
		// don't leak existence
		utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "If the email exists, a reset was sent"})
		return
	}
	tok, _ := models.RandomToken(32)
	_, _ = db.Exec("INSERT INTO password_resets (user_id, token, created_at) VALUES (?, ?, ?)", uid, tok, time.Now().UTC().Format(time.RFC3339))
	resetURL := fmt.Sprintf("%s://%s/reset-password?token=%s", utils.SchemeFromRequest(r), r.Host, tok)
	_ = email.SendVerificationEmail(emailAddress, resetURL) // reuse sender
	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Reset email sent"})
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/static/pages/password.html")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	token := strings.TrimSpace(r.FormValue("token"))
	newPass := r.FormValue("password")
	if len(newPass) < 8 {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Password must be at least 8 characters"})
		return
	}
	var uid int
	if err := db.QueryRow("SELECT user_id FROM password_resets WHERE token = ?", token).Scan(&uid); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid token"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	_, _ = db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hash), uid)
	_, _ = db.Exec("DELETE FROM password_resets WHERE token = ?", token)
	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Password updated"})
}
