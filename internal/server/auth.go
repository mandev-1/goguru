package server

import (
	"camagru/internal/auth"
	"camagru/internal/models"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	if username == "" || password == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Username and password required",
		})
		return
	}

	var user models.User
	err := s.DB.QueryRow(`
		SELECT id, username, email, password_hash, verified
		FROM users
		WHERE username = ? OR email = ?
	`, username, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Verified,
	)

	if err == sql.ErrNoRows {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid username/password combination, sorry! (🇨🇦)",
		})
		return
	}

	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid username/password combination, sorry! (🇨🇦)",
		})
		return
	}

	if !user.Verified {
		s.SendJSON(w, http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Account not verified. Please check your email.",
		})
		return
	}

	// Create session token
	sessionToken, err := auth.GenerateToken()
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create session",
		})
		return
	}

	_, err = s.DB.Exec("UPDATE users SET session_token = ? WHERE id = ?", sessionToken, user.ID)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create session",
		})
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 7, // 7 days
	})

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
	})
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	// Validation - Username
	if username == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Username is required",
		})
		return
	}

	if len(username) < 3 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Username must be at least 3 characters",
		})
		return
	}

	if len(username) > 20 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Username must be at most 20 characters",
		})
		return
	}

	// Username can only contain letters, numbers, and underscores
	usernameRegex := `^[a-zA-Z0-9_]+$`
	matched, _ := regexp.MatchString(usernameRegex, username)
	if !matched {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Username can only contain letters, numbers, and underscores",
		})
		return
	}

	// Validation - Email
	if email == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Email is required",
		})
		return
	}

	if !auth.IsValidEmail(email) {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid email address",
		})
		return
	}

	// Validation - Password
	if password == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Password is required",
		})
		return
	}

	if len(password) < 8 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Password must be at least 8 characters",
		})
		return
	}

	if len(password) > 128 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Password must be at most 128 characters",
		})
		return
	}

	// Validation - Confirm Password
	if confirmPassword == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Please confirm your password",
		})
		return
	}

	if password != confirmPassword {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Passwords do not match",
		})
		return
	}

	// Check if username exists
	var exists int
	s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
	if exists > 0 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Username already taken",
		})
		return
	}

	// Check if email exists
	s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
	if exists > 0 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Email already registered",
		})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process password",
		})
		return
	}

	// Generate verification token
	verificationToken, err := auth.GenerateToken()
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	// Insert user
	_, err = s.DB.Exec(`
		INSERT INTO users (username, email, password_hash, verification_token)
		VALUES (?, ?, ?, ?)
	`, username, email, passwordHash, verificationToken)

	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create account",
		})
		return
	}

	// Send verification email
	verificationURL := fmt.Sprintf("http://localhost:8080/verify?token=%s", verificationToken)
	go s.SendVerificationEmail(email, username, verificationURL)

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Registration successful. Please check your email to verify your account.",
	})
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		s.DB.Exec("UPDATE users SET session_token = NULL WHERE session_token = ?", cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/gallery", http.StatusFound)
}

func (s *Server) HandleVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Invalid verification link", http.StatusBadRequest)
		return
	}

	var userID int
	err := s.DB.QueryRow("SELECT id FROM users WHERE verification_token = ?", token).Scan(&userID)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid verification token", http.StatusBadRequest)
		return
	}

	_, err = s.DB.Exec("UPDATE users SET verified = 1, verification_token = NULL WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to verify account", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login?verified=1", http.StatusFound)
}

func (s *Server) HandleForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "./web/static/pages/forgot-password.html")
		return
	}
	// POST requests handled by HandleForgotPassword
	if r.Method == "POST" {
		s.HandleForgotPassword(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if !auth.IsValidEmail(email) {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid email address",
		})
		return
	}

	var userID int
	var username string
	err := s.DB.QueryRow("SELECT id, username FROM users WHERE email = ?", email).Scan(&userID, &username)
	if err == sql.ErrNoRows {
		// Don't reveal if email exists
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "If the email exists, a password reset link has been sent.",
		})
		return
	}

	// Generate reset token
	resetToken, err := auth.GenerateToken()
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate reset token",
		})
		return
	}

	// Set expiration (1 hour)
	expires := time.Now().Add(1 * time.Hour)
	_, err = s.DB.Exec(`
		UPDATE users
		SET reset_token = ?, reset_expires = ?
		WHERE id = ?
	`, resetToken, expires, userID)

	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	// Send reset email
	resetURL := fmt.Sprintf("http://localhost:8080/password?token=%s", resetToken)
	go s.SendPasswordResetEmail(email, username, resetURL)

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "If the email exists, a password reset link has been sent.",
	})
}

func (s *Server) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	if token == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Reset token required",
		})
		return
	}

	if len(password) < 8 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Password must be at least 8 characters",
		})
		return
	}

	if password != confirmPassword {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Passwords do not match",
		})
		return
	}

	// Verify token
	var userID int
	var expires time.Time
	err := s.DB.QueryRow(`
		SELECT id, reset_expires
		FROM users
		WHERE reset_token = ?
	`, token).Scan(&userID, &expires)

	if err == sql.ErrNoRows || time.Now().After(expires) {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid or expired reset token",
		})
		return
	}

	// Hash new password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process password",
		})
		return
	}

	// Update password and clear reset token
	_, err = s.DB.Exec(`
		UPDATE users
		SET password_hash = ?, reset_token = NULL, reset_expires = NULL
		WHERE id = ?
	`, passwordHash, userID)

	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to reset password",
		})
		return
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password reset successful. You can now login.",
	})
}

func (s *Server) HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))

	var userID int
	var email, usernameVal, token string
	err := s.DB.QueryRow(`
		SELECT id, email, username, verification_token
		FROM users
		WHERE username = ? OR email = ?
	`, username, username).Scan(&userID, &email, &usernameVal, &token)

	if err == sql.ErrNoRows {
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "If the account exists, a verification email has been sent.",
		})
		return
	}

	if token == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Account already verified",
		})
		return
	}

	verificationURL := fmt.Sprintf("http://localhost:8080/verify?token=%s", token)
	go s.SendVerificationEmail(email, usernameVal, verificationURL)

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Verification email sent. Please check your inbox.",
	})
}

