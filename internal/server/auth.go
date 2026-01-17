package server

import (
	"camagru/internal/auth"
	"camagru/internal/models"
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

	user, err := s.DB.GetUserByUsernameOrEmail(username)
	if err != nil {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid username/password combination, sorry! (🇨🇦)",
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
	sessionToken, err := auth.GenerateToken()
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create session",
		})
		return
	}

	if err := s.DB.UpdateUserSessionToken(user.ID, sessionToken); err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create session",
		})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 7,
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
	usernameRegex := `^[a-zA-Z0-9_]+$`
	matched, _ := regexp.MatchString(usernameRegex, username)
	if !matched {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Username can only contain letters, numbers, and underscores",
		})
		return
	}
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
	usernameExists, emailExists, err := s.DB.UserExists(username, email)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	if usernameExists {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Username already taken",
		})
		return
	}

	if emailExists {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Email already registered",
		})
		return
	}
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process password",
		})
		return
	}
	verificationToken, err := auth.GenerateToken()
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}
	_, err = s.DB.CreateUser(username, email, passwordHash, verificationToken)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create account",
		})
		return
	}
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
		s.DB.ClearUserSessionToken(cookie.Value)
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

	err := s.DB.VerifyUser(token)
	if err != nil {
		http.Error(w, "Invalid verification token", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login?verified=1", http.StatusFound)
}

func (s *Server) HandleForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "./web/static/pages/forgot-password.html")
		return
	}
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

	user, err := s.DB.GetUserByUsernameOrEmail(email)
	if err != nil {
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "If the email exists, a password reset link has been sent.",
		})
		return
	}
	username := user.Username
	resetToken, err := auth.GenerateToken()
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate reset token",
		})
		return
	}
	expires := time.Now().Add(30 * time.Minute)
	err = s.DB.SetPasswordResetToken(email, resetToken, expires)

	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}
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
	user, err := s.DB.GetUserByResetToken(token)
	if err != nil {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid or expired reset token",
		})
		return
	}
	userID := user.ID
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process password",
		})
		return
	}
	err = s.DB.UpdateUserPassword(userID, passwordHash)

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

	user, err := s.DB.GetUserByUsernameOrEmail(username)
	if err != nil {
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "If the account exists, a verification email has been sent.",
		})
		return
	}

	if user.VerificationToken == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Account already verified",
		})
		return
	}

	email := user.Email
	usernameVal := user.Username
	token := user.VerificationToken

	verificationURL := fmt.Sprintf("http://localhost:8080/verify?token=%s", token)
	go s.SendVerificationEmail(email, usernameVal, verificationURL)

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Verification email sent. Please check your inbox.",
	})
}
