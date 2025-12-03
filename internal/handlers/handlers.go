package handlers

import (
	"database/sql"
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

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/static/pages/home.html")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "web/static/pages/login.html")
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		var id int
		var hash string
		var verified int
		err := db.QueryRow("SELECT id, password_hash, verified FROM users WHERE username = ?", username).Scan(&id, &hash, &verified)
		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Invalid username/password combination"})
			return
		}
		if verified == 0 {
			utils.WriteJSON(w, http.StatusForbidden, map[string]any{"success": false, "message": "Not verified yet. Want me to resend verification?"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Invalid username/password combination"})
			return
		}
		// Create session token
		token, err := models.RandomToken(32)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}
		// Persist session
		_, err = db.Exec("INSERT INTO sessions (token, user_id, created_at) VALUES (?, ?, ?)", token, id, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}
		// Set cookie (HttpOnly)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(24 * time.Hour),
		})
		utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Logged in"})
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "web/static/pages/register.html")
		return
	case http.MethodPost:
		if err := r.ParseMultipartForm(2 << 20); err != nil { // 2MB
			utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
		password := r.FormValue("password")
		confirm := r.FormValue("confirmPassword")

		if err := models.ValidateRegistration(username, email, password, confirm); err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
			return
		}

		// Uniqueness checks
		if exists, _ := userExists("username", username); exists {
			utils.WriteJSON(w, http.StatusConflict, map[string]any{"success": false, "message": "Username already taken"})
			return
		}
		if exists, _ := userExists("email", email); exists {
			utils.WriteJSON(w, http.StatusConflict, map[string]any{"success": false, "message": "Email already registered"})
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}

		token, err := models.RandomToken(32)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}

		// Store user
		_, err = db.Exec(`INSERT INTO users (username, email, password_hash, verified, verify_token, created_at)
						  VALUES (?, ?, ?, 0, ?, ?)`,
			username, email, string(hash), token, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Could not create user"})
			return
		}

		// Send verification email via MailHog SMTP (localhost:1025)
		verifyURL := "http://" + r.Host + "/verify?token=" + token
		// if err := email.SendVerificationEmail(email, verifyURL); err != nil {
		// 	log.Printf("Email send error: %v", err)
		// }

		utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Registration successful. Check your email to verify.", "verify_url": verifyURL})
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func GalleryHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/static/pages/gallery.html")
}

func EditorHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/static/pages/editor.html")
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/static/pages/user.html")
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if c, err := r.Cookie("session"); err == nil {
		_, _ = db.Exec("DELETE FROM sessions WHERE token = ?", c.Value)
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1})
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Logged out"})
}

func CurrentUserAPIHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := utils.CurrentUserID(r, db)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Unauthorized"})
		return
	}

	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", uid).Scan(&username)
	if err != nil {
		utils.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "User not found"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "username": username})
}

func userExists(field, value string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(1) FROM users WHERE "+field+" = ?", value).Scan(&count)
	return count > 0, err
}
