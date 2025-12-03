package image

import (
	"database/sql"
	"fmt"
	"camagru/internal/email"
	"camagru/internal/utils"
	"image"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"strings"
	"time"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

// assetsHandler returns list of available overlay assets
func AssetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.Query("SELECT id, name, path FROM assets ORDER BY id DESC")
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
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
func UploadAssetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	file, hdr, err := r.FormFile("file")
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "File required"})
		return
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Only PNG supported"})
		return
	}
	if name == "" {
		name = hdr.Filename
	}
	_ = os.MkdirAll("web/static/assets", 0o755)
	fname := time.Now().UTC().Format("20060102T150405") + "_" + utils.SanitizeFilename(name) + ".png"
	rel := "web/static/assets/" + fname
	f, err := os.Create(rel)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	_, err = db.Exec("INSERT INTO assets (name, path, created_at) VALUES (?, ?, ?)", name, "/static/assets/"+fname, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "DB error"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Asset uploaded"})
}

// composeHandler composes a user-provided image (PNG) with a selected asset server-side and saves
func ComposeHandler(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	assetID := strings.TrimSpace(r.FormValue("asset_id"))
	if assetID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "asset_id required"})
		return
	}
	var assetPath string
	if err := db.QueryRow("SELECT path FROM assets WHERE id = ?", assetID).Scan(&assetPath); err != nil {
		utils.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "Asset not found"})
		return
	}
	af, err := os.Open("web/static" + strings.TrimPrefix(assetPath, "/static"))
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	defer af.Close()
	overlay, err := png.Decode(af)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Asset decode error"})
		return
	}

	uf, _, err := r.FormFile("image")
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Image required"})
		return
	}
	defer uf.Close()
	base, err := png.Decode(uf)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Image must be PNG"})
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
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Save error"})
		return
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Save error"})
		return
	}

	_, err = db.Exec("INSERT INTO images (user_id, path, created_at) VALUES (?, ?, ?)", userID, "/static/uploads/"+fname, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "DB error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Composed", "path": "/static/uploads/" + fname})
}

func LikeImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uid, err := utils.CurrentUserID(r, db)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	imgID := strings.TrimSpace(r.FormValue("image_id"))
	_, err = db.Exec("INSERT OR IGNORE INTO likes (user_id, image_id, created_at) VALUES (?, ?, ?)", uid, imgID, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Liked"})
}

func CommentImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uid, err := utils.CurrentUserID(r, db)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Invalid form"})
		return
	}
	imgID := strings.TrimSpace(r.FormValue("image_id"))
	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "Comment required"})
		return
	}
	_, err = db.Exec("INSERT INTO comments (user_id, image_id, body, created_at) VALUES (?, ?, ?, ?)", uid, imgID, body, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Server error"})
		return
	}
	// Notify image author if pref enabled
	var authorID int
	var emailAddress string
	_ = db.QueryRow("SELECT users.id, users.email FROM images JOIN users ON users.id = images.user_id WHERE images.id = ?", imgID).Scan(&authorID, &emailAddress)
	var notify int
	_ = db.QueryRow("SELECT notify_comments FROM user_prefs WHERE user_id = ?", authorID).Scan(&notify)
	if notify == 1 && emailAddress != "" {
		_ = email.SendVerificationEmail(emailAddress, "New comment on your image")
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Commented"})
}
