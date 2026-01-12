package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// Using static HTML pages; templates disabled
var db *sql.DB

func init() {
	// Static pages mode; no template parsing

	// Minimal SQLite database (file-based). Pure Go driver (modernc.org/sqlite) avoids CGO.
	var err error
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		db, err = sql.Open("sqlite", "file:camagru.db?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)")
	} else {
		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)
		db, err = sql.Open("postgres", psqlInfo)
	}

	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err = initSchema(db); err != nil {
		log.Fatalf("DB init schema error: %v", err)
	}
}

// =============== MAIN == ROUTER, APIs ==============================
func main() {
	// Serve static files (CSS, JS, images)
	fs := http.FileServer(http.Dir("web/static"))
	// Use URL prefix here ("/static/"), not a filesystem path
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/resend-verification", resendVerificationHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/gallery", authRequired(galleryHandler))
	http.HandleFunc("/editor", authRequired(editorHandler))
	http.HandleFunc("/verify", verifyHandler)
	// Auth/account
	http.HandleFunc("/logout", authRequired(logoutHandler))
	http.HandleFunc("/forgot-password", forgotPasswordHandler)
	http.HandleFunc("/reset-password", resetPasswordHandler)
	http.HandleFunc("/user", authRequired(userHandler))
	// Editor APIs
	http.HandleFunc("/api/assets", authRequired(assetsHandler))
	http.HandleFunc("/api/assets/upload", authRequired(uploadAssetHandler))
	http.HandleFunc("/api/compose", authRequired(composeHandler))
	// Gallery APIs
	http.HandleFunc("/api/gallery", galleryListHandler)
	http.HandleFunc("/api/gallery/like", authRequired(likeImageHandler))
	http.HandleFunc("/api/gallery/comment", authRequired(commentImageHandler))
	http.HandleFunc("/api/current-user", authRequired(currentUserAPIToBeDeletedHandler))

	http.HandleFunc("/api/gallery/mock-upload", authRequired(mockUploadHandler))

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// HELPERS (HANDLERS) FOR DA ROUTER =======================
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	// Serve static home page
	http.ServeFile(w, r, "web/static/pages/home.html")
}

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	// Serve static gallery page
	http.ServeFile(w, r, "web/static/pages/gallery.html")
}

func authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := currentUserID(r); err != nil {
			// If it's an API call, return JSON, otherwise serve the unauthorized page.
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Unauthorized"})
			} else {
				http.ServeFile(w, r, "web/static/pages/unauthorized.html")
			}
			return
		}
		next(w, r)
	}
}

func editorHandler(w http.ResponseWriter, r *http.Request) {
	// Serve static editor page
	http.ServeFile(w, r, "web/static/pages/editor.html")
}

// HELPERS (HANDLERS) FOR DA SIGN-UP and SIGN-IN =======================
func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "web/static/pages/login.html")
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		var id int
		var hash string
		var verified int
		err := db.QueryRow("SELECT id, password_hash, verified FROM users WHERE username = ?", username).Scan(&id, &hash, &verified)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Invalid username/password combination"})
			return
		}
		if verified == 0 {
			writeJSON(w, http.StatusForbidden, map[string]any{"success": false, "message": "Not verified yet. Want me to resend verification?"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Invalid username/password combination"})
			return
		}
		// Create session token
		token, err := randomToken(32)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}
		// Persist session
		_, err = db.Exec("INSERT INTO sessions (token, user_id, created_at) VALUES (?, ?, ?)", token, id, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
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
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Logged in"})
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "web/static/pages/register.html")
		return
	case http.MethodPost:
		if err := r.ParseMultipartForm(2 << 20); err != nil { // 2MB
			writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
		password := r.FormValue("password")
		confirm := r.FormValue("confirmPassword")

		if err := validateRegistration(username, email, password, confirm); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
			return
		}

		// Uniqueness checks
		if exists, _ := userExists("username", username); exists {
			writeJSON(w, http.StatusConflict, map[string]any{"success": false, "message": "Username already taken"})
			return
		}
		if exists, _ := userExists("email", email); exists {
			writeJSON(w, http.StatusConflict, map[string]any{"success": false, "message": "Email already registered"})
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}

		token, err := randomToken(32)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}

		// Store user
		_, err = db.Exec(`INSERT INTO users (username, email, password_hash, verified, verify_token, created_at)
						  VALUES (?, ?, ?, 0, ?, ?)`,
			username, email, string(hash), token, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Could not create user"})
			return
		}

		// Send verification email via MailHog SMTP (localhost:1025)
		verifyURL := fmt.Sprintf("%s://%s/verify?token=%s", schemeFromRequest(r), r.Host, token)
		if err := sendVerificationEmail(email, verifyURL); err != nil {
			log.Printf("Email send error: %v", err)
		}

		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Registration successful. Check your email to verify."})
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// HELPERS (HANDLERS) FOR DA USER VERIFICATION (EMAIL THINGS) =======================
// resendVerificationHandler triggers re-sending a verify email for a user.
func resendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	identifier := strings.TrimSpace(r.FormValue("username"))
	if identifier == "" {
		identifier = strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	}
	if identifier == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Username or email required"})
		return
	}

	var id int
	var email string
	var verified int
	var token string
	err := db.QueryRow("SELECT id, email, verified, verify_token FROM users WHERE username = ? OR email = ?", identifier, identifier).Scan(&id, &email, &verified, &token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "User not found"})
		return
	}
	if verified == 1 {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Already verified"})
		return
	}
	if token == "" {
		// regenerate token if missing
		token, err = randomToken(32)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}
		if _, err = db.Exec("UPDATE users SET verify_token = ? WHERE id = ?", token, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}
	}
	verifyURL := fmt.Sprintf("%s://%s/verify?token=%s", "http", r.Host, token)
	if err := sendVerificationEmail(email, verifyURL); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Failed to send email"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Verification email resent"})
}

// assetsHandler returns list of available overlay assets
func assetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.Query("SELECT id, name, path FROM assets ORDER BY id DESC")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	defer rows.Close()
	type A struct {
		ID         int
		Name, Path string
	}
	var list []A
	for rows.Next() {
		var a A
		if err := rows.Scan(&a.ID, &a.Name, &a.Path); err == nil {
			list = append(list, a)
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, "[")
	for i, a := range list {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"id":%d,"name":%q,"path":%q}`, a.ID, a.Name, a.Path)
	}
	fmt.Fprint(w, "]")
}

// uploadAssetHandler accepts PNG with alpha and stores under web/static/assets
func uploadAssetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "File required"})
		return
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Only PNG supported"})
		return
	}
	if name == "" {
		name = hdr.Filename
	}
	_ = os.MkdirAll("web/static/assets", 0o755)
	fname := time.Now().UTC().Format("20060102T150405") + "_" + sanitizeFilename(name) + ".png"
	rel := "web/static/assets/" + fname
	f, err := os.Create(rel)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	_, err = db.Exec("INSERT INTO assets (name, path, created_at) VALUES (?, ?, ?)", name, "/static/assets/"+fname, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "DB error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Asset uploaded"})
}

// composeHandler composes a user-provided image (PNG) with a selected asset server-side and saves
func composeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Minimal auth: require session cookie and resolve user_id
	c, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var userID int
	if err := db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", c.Value).Scan(&userID); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	assetID := strings.TrimSpace(r.FormValue("asset_id"))
	if assetID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "asset_id required"})
		return
	}
	var assetPath string
	if err := db.QueryRow("SELECT path FROM assets WHERE id = ?", assetID).Scan(&assetPath); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "Asset not found"})
		return
	}
	af, err := os.Open("web/static" + strings.TrimPrefix(assetPath, "/static"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	defer af.Close()
	overlay, err := png.Decode(af)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Asset decode error"})
		return
	}

	uf, _, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Image required"})
		return
	}
	defer uf.Close()
	base, err := png.Decode(uf)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Image must be PNG"})
		return
	}

	out := image.NewNRGBA(base.Bounds())
	draw.Draw(out, base.Bounds(), base, image.Point{}, draw.Src)
	dx := (base.Bounds().Dx() - overlay.Bounds().Dx()) / 2
	dy := (base.Bounds().Dy() - overlay.Bounds().Dy()) / 2
	pos := image.Rect(dx, dy, dx+overlay.Bounds().Dx(), dy+overlay.Bounds().Dy())
	draw.Draw(out, pos, overlay, image.Point{}, draw.Over)

	_ = os.MkdirAll("web/static/uploads", 0o755)
	fname := fmt.Sprintf("%d_%s.png", userID, time.Now().UTC().Format("20060102T150405"))
	rel := "web/static/uploads/" + fname
	f, err := os.Create(rel)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Save error"})
		return
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Save error"})
		return
	}

	res, err := db.Exec("INSERT INTO images (user_id, path, created_at) VALUES (?, ?, ?)", userID, "/static/uploads/"+fname, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "DB error"})
		return
	}
	imgID, _ := res.LastInsertId()
	_, _ = db.Exec("INSERT OR IGNORE INTO gallery_posts (id, image_id, user_id, created_at) VALUES (?, ?, ?, ?)", imgID, imgID, userID, time.Now().UTC().Format(time.RFC3339))

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Composed", "path": "/static/uploads/" + fname})
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			return r
		}
		return '-'
	}, s)
	s = strings.Trim(s, "-_")
	if s == "" {
		s = "asset"
	}
	return s
}

// --- Minimal helpers ---

func initSchema(d *sql.DB) error {
	_, err := d.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		verified INTEGER NOT NULL DEFAULT 0,
		verify_token TEXT,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)
	if err != nil {
		return err
	}
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS assets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		path TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)
	if err != nil {
		return err
	}
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS gallery_posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image_id INTEGER NOT NULL UNIQUE,
		user_id INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY(image_id) REFERENCES images(id) ON DELETE CASCADE,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)
	if err != nil {
		return err
	}
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		image_id INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		UNIQUE(user_id, image_id),
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(image_id) REFERENCES images(id) ON DELETE CASCADE
	)`)
	if err != nil {
		return err
	}
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		image_id INTEGER NOT NULL,
		body TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(image_id) REFERENCES images(id) ON DELETE CASCADE
	)`)
	if err != nil {
		return err
	}
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS user_prefs (
		user_id INTEGER PRIMARY KEY,
		notify_comments INTEGER NOT NULL DEFAULT 1,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)
	if err != nil {
		return err
	}
	_, _ = d.Exec(`INSERT OR IGNORE INTO gallery_posts (id, image_id, user_id, created_at)
		SELECT id, id, user_id, created_at FROM images`)
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS password_resets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)
	return err
}

func userExists(field, value string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(1) FROM users WHERE "+field+" = ?", value).Scan(&count)
	return count > 0, err
}

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func validateRegistration(username, email, password, confirm string) error {
	if len(username) < 3 || len(username) > 20 {
		return errors.New("Username must be 3-20 characters")
	}
	for _, c := range username {
		if !(c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return errors.New("Username may contain letters, numbers, underscore")
		}
	}
	if !emailRe.MatchString(email) {
		return errors.New("Invalid email address")
	}
	if len(password) < 8 {
		return errors.New("Password must be at least 8 characters")
	}
	if password != confirm {
		return errors.New("Passwords do not match")
	}
	return nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error encoding json: %v", err)
	}
}

func schemeFromRequest(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	// honor X-Forwarded-Proto if behind proxy
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		return p
	}
	return "http"
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
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

// --- Auth helpers and handlers ---

func currentUserID(r *http.Request) (int, error) {
	c, err := r.Cookie("session")
	if err != nil {
		return 0, errors.New("no session")
	}
	var uid int
	if err := db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", c.Value).Scan(&uid); err != nil {
		return 0, errors.New("invalid session")
	}
	return uid, nil
}

// currentUsernameFromRequest removed: static pages do not personalize

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if c, err := r.Cookie("session"); err == nil {
		_, _ = db.Exec("DELETE FROM sessions WHERE token = ?", c.Value)
		// Expire cookie
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Logged out"})
}

func forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/static/pages/password.html")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	var uid int
	if err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&uid); err != nil {
		// don't leak existence
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "If the email exists, a reset was sent"})
		return
	}
	tok, _ := randomToken(32)
	_, _ = db.Exec("INSERT INTO password_resets (user_id, token, created_at) VALUES (?, ?, ?)", uid, tok, time.Now().UTC().Format(time.RFC3339))
	resetURL := fmt.Sprintf("%s://%s/reset-password?token=%s", schemeFromRequest(r), r.Host, tok)
	_ = sendVerificationEmail(email, resetURL) // reuse sender
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Reset email sent"})
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/static/pages/password.html")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	token := strings.TrimSpace(r.FormValue("token"))
	newPass := r.FormValue("password")
	if len(newPass) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Password must be at least 8 characters"})
		return
	}
	var uid int
	if err := db.QueryRow("SELECT user_id FROM password_resets WHERE token = ?", token).Scan(&uid); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid token"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	_, _ = db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hash), uid)
	_, _ = db.Exec("DELETE FROM password_resets WHERE token = ?", token)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Password updated"})
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/static/pages/user.html")
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func currentUserAPIToBeDeletedHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := currentUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Unauthorized"})
		return
	}

	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", uid).Scan(&username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "User not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "username": username})
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// --- Gallery listing and mock upload ---

func mockUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uid, err := currentUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "Unauthorized"})
		return
	}

	srcPath := filepath.Join("img", "b.jpeg")
	data, err := os.ReadFile(srcPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Mock image missing"})
		return
	}

	_ = os.MkdirAll("web/static/uploads", 0o755)
	fname := fmt.Sprintf("%d_mock_%d.jpeg", uid, time.Now().Unix())
	dstRel := filepath.Join("web/static/uploads", fname)
	if err := os.WriteFile(dstRel, data, 0o644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Save failed"})
		return
	}
	webPath := "/static/uploads/" + fname
	res, err := db.Exec("INSERT INTO images (user_id, path, created_at) VALUES (?, ?, ?)", uid, webPath, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "DB error"})
		return
	}
	imgID, _ := res.LastInsertId()
	_, _ = db.Exec("INSERT OR IGNORE INTO gallery_posts (id, image_id, user_id, created_at) VALUES (?, ?, ?, ?)", imgID, imgID, uid, time.Now().UTC().Format(time.RFC3339))

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "path": webPath})
}

// --- Gallery interactions ---

func galleryListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	page := 1
	if p := strings.TrimSpace(r.URL.Query().Get("page")); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	const pageSize = 9
	offset := (page - 1) * pageSize

	uid, _ := currentUserID(r)

	rows, err := db.Query(`SELECT gallery_posts.id, images.path, gallery_posts.created_at, users.username
		FROM gallery_posts
		JOIN images ON images.id = gallery_posts.image_id
		JOIN users ON users.id = gallery_posts.user_id
		ORDER BY gallery_posts.created_at DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	defer rows.Close()

	type comment struct {
		Body      string `json:"body"`
		Author    string `json:"author"`
		CreatedAt string `json:"createdAt"`
	}
	type item struct {
		ID        int       `json:"id"`
		Path      string    `json:"path"`
		Author    string    `json:"author"`
		Likes     int       `json:"likes"`
		Liked     bool      `json:"liked"`
		CreatedAt string    `json:"createdAt"`
		Comments  []comment `json:"comments"`
	}

	items := []item{}
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.ID, &it.Path, &it.CreatedAt, &it.Author); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
			return
		}
		_ = db.QueryRow("SELECT COUNT(1) FROM likes WHERE image_id = ?", it.ID).Scan(&it.Likes)
		if uid > 0 {
			var liked int
			_ = db.QueryRow("SELECT COUNT(1) FROM likes WHERE image_id = ? AND user_id = ?", it.ID, uid).Scan(&liked)
			it.Liked = liked > 0
		}
		crow, err := db.Query(`SELECT comments.body, comments.created_at, users.username
			FROM comments JOIN users ON users.id = comments.user_id
			WHERE comments.image_id = ? ORDER BY comments.created_at DESC LIMIT 5`, it.ID)
		if err == nil {
			for crow.Next() {
				var c comment
				if err := crow.Scan(&c.Body, &c.CreatedAt, &c.Author); err == nil {
					it.Comments = append(it.Comments, c)
				}
			}
			crow.Close()
		}
		items = append(items, it)
	}

	var total int
	_ = db.QueryRow("SELECT COUNT(1) FROM gallery_posts").Scan(&total)
	hasMore := page*pageSize < total

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"items":   items,
		"page":    page,
		"hasMore": hasMore,
	})
}

func likeImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uid, err := currentUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	imgID := strings.TrimSpace(r.FormValue("image_id"))
	_, err = db.Exec("INSERT OR IGNORE INTO likes (user_id, image_id, created_at) VALUES (?, ?, ?)", uid, imgID, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Liked"})
}

func commentImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uid, err := currentUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	imgID := strings.TrimSpace(r.FormValue("image_id"))
	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Comment required"})
		return
	}
	_, err = db.Exec("INSERT INTO comments (user_id, image_id, body, created_at) VALUES (?, ?, ?, ?)", uid, imgID, body, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	// Notify image author if pref enabled
	var authorID int
	var email string
	_ = db.QueryRow("SELECT users.id, users.email FROM images JOIN users ON users.id = images.user_id WHERE images.id = ?", imgID).Scan(&authorID, &email)
	var notify int
	_ = db.QueryRow("SELECT notify_comments FROM user_prefs WHERE user_id = ?", authorID).Scan(&notify)
	if notify == 1 && email != "" {
		_ = sendVerificationEmail(email, "New comment on your image")
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Commented"})
}

// VERIFICATION EMAIL ======================================================================
// sendVerificationEmail sends an HTML verification email via local MailHog.
func sendVerificationEmail(to, verifyURL string) error {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "mailhog"
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "1025"
	}
	addr := net.JoinHostPort(host, port)

	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "camagru@localhost"
	}

	subj := "Camagru: Verify your email"
	body := buildVerificationEmailHTML(verifyURL)
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subj,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

func buildVerificationEmailHTML(verifyURL string) string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset='iso-8859-1' content='text/html' http-equiv='Content-Type'>
<link rel="stylesheet" media="screen" href="https://use.typekit.net/bzd7hlb.css" />
<title>
Information
</title>
</head>
<body bgcolor='#fff' style='background-color: #FFFFFF'>
<div style="background-color: #FFFFFF; font-family: 'Noto Sans', sans-serif; color: #333;">
<table bgcolor='#FFFFFF' border='0' cellpadding='0' cellspacing='0' style='margin: auto; background-color: #FFFFFF' width='600'>
<tbody>
<tr>
<td>
<table bgcolor='#FFFFFF' border='0' cellpadding='0' cellspacing='0' style='margin: auto; background-color: #FFFFFF' width='600'>
<tbody>
<tr>
<td width='30'></td>
<td style='text-align: left;' width='80'></td>
<td style='text-align: center; height: 100px; width: auto;' width='380'>
<img alt='logo42' style='height: 100px; width: auto;'>
 </td>
<td width='80'></td>
<td width='30'></td>
</tr>
</tbody>
</table>
<table bgcolor='#FFFFFF' border='0' cellpadding='0' cellspacing='0' height='20' style='margin: auto; background-color: #FFFFFF' width='600'>
<tbody>
<tr>
<td width='20'></td>
<td style="font-family: 'Noto Sans', sans-serif; font-size: 28px; font-weight: normal; color: #777; text-align: center; padding: 30px 0px 10px;" width='560'>Registration to Camagru</td>
<td width='20'></td>
</tr>
<tr>
<td width='20'></td>
<td style="font-family: 'Noto Sans', sans-serif; font-size: 18px; font-weight: normal; color: #999; text-align: center; padding: 0px 0px 50px;" width='560'></td>
<td width='20'></td>
</tr>
</tbody>
</table>
<table bgcolor='#FFFFFF' border='0' cellpadding='0' cellspacing='0' height='' style='margin: auto; background-color: #FFFFFF' width='600'>
<tbody>
<tr>
<td width='30'></td>
<td style='border-top: 1px solid #eee; padding: 30px 0' width='540'>
<table border='0' cellspacing='0' height='' style='margin: auto; background-color: #FFFFFF' width='540'>
<tbody>
<tr>
<td height='15' width='20'></td>
<td height='15' width='500'></td>
<td height='15' width='20'></td>
</tr>
<tr>
<td width='15'></td>
<td style="text-align: justify; font-family: 'Noto Sans', sans-serif; font-size: 14px; color: #333" width='500'>
<p>You just registered to the use Martin's <em>Camagru</em>.</p>
<p>To complete your registration, please verify your email by clicking the link below:</p>
<p style='text-align:center; margin: 20px 0;'>
  <a href='` + verifyURL + `' style='display:inline-block; background:#00BABC; color:#fff; padding:12px 18px; text-decoration:none; border-radius:4px;'>Verify Email</a>
</p>
<p>If the button doesn't work, copy and paste this URL into your browser:</p>
<p style='word-break:break-all;'>` + verifyURL + `</p>
</td>
<td width='15'></td>
</tr>
<tr>
<td height='15' width='20'></td>
<td height='15' width='500'></td>
<td height='15' width='20'></td>
</tr>
</tbody>
</table>
</td>
<td width='30'></td>
</tr>
</tbody>
</table>

<table bgcolor='#FFFFFF' border='0' cellpadding='0' cellspacing='0' height='' style='margin: auto; background-color: #FFFFFF; padding: 30px 0;' width='600'>
<tbody>
<tr>
<td width='30'></td>
<td style='border-top: 1px solid #eee; text-align: center;' width='540'>
<table bgcolor='#FFFFFF' border='0' cellspacing='0' height='' style='margin: auto; background-color: #FFFFFF;' width='540'>
<tbody>
<tr style='background-color: #FFFFFF;'>
<td height='15' width='20'></td>
<td height='15' width='500'></td>
<td height='15' width='20'></td>
</tr>
<tr style='background-color: #FFFFFF;'>
<td bgcolor='#FFFFFF' colspan='3' style="text-align: center; color: #BBB; font-family: 'futura-pt', sans-serif; font-size: 12px; text-transform: uppercase;">
This email was sent by
<a href='http://www.42.fr' style='color: #777; text-decoration: none;'>
<span style='color: #00BABC;'>mman</span>
</a>
<br>
<span></span>
<br>
<span> </span>
<br>
<br>
</td>
</tr>
<tr style='background-color: #FFFFFF;'>
<td height='15' width='20'></td>
<td height='15' width='500'></td>
<td height='15' width='20'></td>
</tr>
<tr style='text-align: center; background-color: #FFFFFF;'>
<td bgcolor='#FFFFFF' width='20'></td>
<td bgcolor='#FFFFFF' width='500'>
</td>
<td bgcolor='#FFFFFF' width='20'></td>
</tr>
<tr style='background-color: #FFFFFF;'>
<td height='15' width='20'></td>
<td height='15' width='500'></td>
<td height='15' width='20'></td>
</tr>
<tr style='background-color: #FFFFFF;'>
<td height='15' width='20'></td>
<td height='15' width='500'></td>
<td height='15' width='20'></td>
</tr>
<tr style='background-color: #FFFFFF;'>
<td height='15' width='20'></td>
<td height='15' width='500'></td>
<td height='15' width='20'></td>
</tr>
</tbody>
</table>
</td>
<td width='30'></td>
</tr>
</tbody>
</table>
</div>
</body>
</html>`
}
