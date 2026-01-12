package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/camagru/internal/models"
	"github.com/yourusername/camagru/internal/services"
	"github.com/yourusername/camagru/internal/utils"
)

type AuthHandler struct {
	userRepo    *models.UserRepository
	sessionRepo *models.SessionRepository
	emailSvc    *services.EmailService
}

func NewAuthHandler(userRepo *models.UserRepository, sessionRepo *models.SessionRepository, emailSvc *services.EmailService) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		emailSvc:    emailSvc,
	}
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/static/pages/login.html")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid form")
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	user, err := h.userRepo.FindByUsername(username)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid username/password combination")
		return
	}

	if !user.Verified {
		utils.WriteError(w, http.StatusForbidden, "Not verified yet. Want me to resend verification?")
		return
	}

	if !user.CheckPassword(password) {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid username/password combination")
		return
	}

	token, err := utils.RandomToken(32)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	if err := h.sessionRepo.Create(token, user.ID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	utils.WriteSuccess(w, "Logged in")
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/static/pages/register.html")
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid form")
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	confirm := r.FormValue("confirmPassword")

	if err := services.ValidateRegistration(username, email, password, confirm); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if exists, _ := h.userRepo.Exists("username", username); exists {
		utils.WriteError(w, http.StatusConflict, "Username already taken")
		return
	}

	if exists, _ := h.userRepo.Exists("email", email); exists {
		utils.WriteError(w, http.StatusConflict, "Email already registered")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	token, err := utils.RandomToken(32)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	if _, err := h.userRepo.Create(username, email, string(hash), token); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	verifyURL := fmt.Sprintf("%s://%s/verify?token=%s", utils.SchemeFromRequest(r), r.Host, token)
	_ = h.emailSvc.SendVerificationEmail(email, verifyURL)

	utils.WriteSuccess(w, "Registration successful. Check your email to verify.")
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Verify(token); err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid form")
		return
	}

	identifier := strings.TrimSpace(r.FormValue("username"))
	if identifier == "" {
		identifier = strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	}

	var user *models.User
	var err error

	if strings.Contains(identifier, "@") {
		user, err = h.userRepo.FindByEmail(identifier)
	} else {
		user, err = h.userRepo.FindByUsername(identifier)
	}

	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	if user.Verified {
		utils.WriteSuccess(w, "Already verified")
		return
	}

	token := user.VerifyToken
	if token == "" {
		token, err = utils.RandomToken(32)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Server error")
			return
		}
		if err := h.userRepo.SetVerifyToken(user.ID, token); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Server error")
			return
		}
	}

	verifyURL := fmt.Sprintf("http://%s/verify?token=%s", r.Host, token)
	if err := h.emailSvc.SendVerificationEmail(user.Email, verifyURL); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to send email")
		return
	}

	utils.WriteSuccess(w, "Verification email resent")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		_ = h.sessionRepo.Delete(c.Value)
		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
			MaxAge:  -1,
		})
	}
	utils.WriteSuccess(w, "Logged out")
}