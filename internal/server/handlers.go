package server

import (
	"camagru/internal/auth"
	"camagru/internal/models"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Page handlers
func (s *Server) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	// Check if user is logged in
	user, err := s.GetCurrentUser(r)
	if err != nil || user == nil || !user.Verified {
		// Not logged in - redirect to gallery
		http.Redirect(w, r, "/gallery", http.StatusFound)
		return
	}
	
	// Logged in - show home page
	http.ServeFile(w, r, "./web/static/pages/home.html")
}

func (s *Server) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "./web/static/pages/login.html")
		return
	}
	// POST requests handled by handleLogin
	if r.Method == "POST" {
		s.HandleLogin(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "./web/static/pages/register.html")
		return
	}
	// POST requests handled by handleRegister
	if r.Method == "POST" {
		s.HandleRegister(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) HandleGalleryPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/pages/gallery.html")
}

func (s *Server) HandleEditorPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/pages/editor.html")
}

func (s *Server) HandleUserPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/pages/user.html")
}

func (s *Server) HandlePasswordPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/pages/password.html")
}

func (s *Server) HandleUnauthorizedPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/pages/unauthorized.html")
}

// API handlers
func (s *Server) HandleCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := s.GetCurrentUser(r)
	if err != nil {
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: false,
		})
		return
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
			"id":       user.ID,
		},
	})
}

func (s *Server) HandleAssets(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query("SELECT id, name, path FROM assets ORDER BY id")
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to load assets",
		})
		return
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var asset models.Asset
		if err := rows.Scan(&asset.ID, &asset.Name, &asset.Path); err != nil {
			continue
		}
		assets = append(assets, asset)
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    assets,
	})
}

func (s *Server) HandleGallery(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 12
	offset := (page - 1) * limit

	// Get images with pagination
	rows, err := s.DB.Query(`
		SELECT i.id, i.user_id, i.path, i.created_at, u.username,
			(SELECT COUNT(*) FROM likes WHERE image_id = i.id) as like_count
		FROM images i
		JOIN users u ON i.user_id = u.id
		ORDER BY i.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to load gallery",
		})
		return
	}
	defer rows.Close()

	var images []models.Image
	var imageIDs []int
	for rows.Next() {
		var img models.Image
		if err := rows.Scan(&img.ID, &img.UserID, &img.Path, &img.CreatedAt, &img.Author, &img.Likes); err != nil {
			continue
		}
		images = append(images, img)
		imageIDs = append(imageIDs, img.ID)
	}

	// Check if user liked these images
	user, _ := s.GetCurrentUser(r)
	if user != nil && len(imageIDs) > 0 {
		placeholders := strings.Repeat("?,", len(imageIDs))
		placeholders = placeholders[:len(placeholders)-1]
		query := fmt.Sprintf("SELECT image_id FROM likes WHERE user_id = ? AND image_id IN (%s)", placeholders)
		args := []interface{}{user.ID}
		for _, id := range imageIDs {
			args = append(args, id)
		}
		likeRows, err := s.DB.Query(query, args...)
		if err == nil {
			likedMap := make(map[int]bool)
			for likeRows.Next() {
				var imgID int
				likeRows.Scan(&imgID)
				likedMap[imgID] = true
			}
			likeRows.Close()
			for i := range images {
				images[i].Liked = likedMap[images[i].ID]
			}
		}
	}

	// Get comments for each image
	for i := range images {
		commentRows, err := s.DB.Query(`
			SELECT c.id, c.body, c.created_at, u.username
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.image_id = ?
			ORDER BY c.created_at DESC
			LIMIT 10
		`, images[i].ID)
		if err == nil {
			var comments []models.Comment
			for commentRows.Next() {
				var c models.Comment
				commentRows.Scan(&c.ID, &c.Body, &c.CreatedAt, &c.Author)
				comments = append(comments, c)
			}
			commentRows.Close()
			images[i].Comments = comments
		}
	}

	// Get total count and calculate total pages
	var total int
	s.DB.QueryRow("SELECT COUNT(*) FROM images").Scan(&total)
	hasMore := offset+limit < total
	totalPages := (total + limit - 1) / limit // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"items":      images,
			"hasMore":    hasMore,
			"totalPages": totalPages,
			"currentPage": page,
			"total":      total,
		},
	})
}

func (s *Server) HandleLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := s.GetCurrentUser(r)
	if err != nil {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "redirect:/register",
		})
		return
	}

	imageID, _ := strconv.Atoi(r.FormValue("image_id"))
	if imageID == 0 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid image ID",
		})
		return
	}

	// Check if already liked
	var exists int
	err = s.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND image_id = ?", user.ID, imageID).Scan(&exists)
	if err == nil && exists > 0 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Already liked",
		})
		return
	}

	// Insert like
	_, err = s.DB.Exec("INSERT INTO likes (user_id, image_id) VALUES (?, ?)", user.ID, imageID)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to like image",
		})
		return
	}

	// Get image owner to send notification
	var imageOwnerID int
	var imageOwnerEmail string
	var imageOwnerNotifications bool
	err = s.DB.QueryRow(`
		SELECT u.id, u.email, u.comment_notifications
		FROM images i
		JOIN users u ON i.user_id = u.id
		WHERE i.id = ?
	`, imageID).Scan(&imageOwnerID, &imageOwnerEmail, &imageOwnerNotifications)

	if err == nil && imageOwnerID != user.ID && imageOwnerNotifications {
		// Send email notification
		go s.SendLikeNotification(imageOwnerEmail, user.Username)
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Image liked",
	})
}

func (s *Server) HandleComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := s.GetCurrentUser(r)
	if err != nil {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "redirect:/register",
		})
		return
	}

	imageID, _ := strconv.Atoi(r.FormValue("image_id"))
	body := strings.TrimSpace(r.FormValue("body"))
	if imageID == 0 || body == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// Insert comment
	_, err = s.DB.Exec("INSERT INTO comments (image_id, user_id, body) VALUES (?, ?, ?)", imageID, user.ID, body)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to add comment",
		})
		return
	}

	// Get image owner to send notification
	var imageOwnerID int
	var imageOwnerEmail string
	var imageOwnerNotifications bool
	err = s.DB.QueryRow(`
		SELECT u.id, u.email, u.comment_notifications
		FROM images i
		JOIN users u ON i.user_id = u.id
		WHERE i.id = ?
	`, imageID).Scan(&imageOwnerID, &imageOwnerEmail, &imageOwnerNotifications)

	if err == nil && imageOwnerID != user.ID && imageOwnerNotifications {
		// Send email notification
		go s.SendCommentNotification(imageOwnerEmail, user.Username, body)
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Comment added",
	})
}

func (s *Server) HandleDeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := s.GetCurrentUser(r)
	if err != nil {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	imageID, _ := strconv.Atoi(r.FormValue("image_id"))
	if imageID == 0 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid image ID",
		})
		return
	}

	// Verify ownership
	var ownerID int
	err = s.DB.QueryRow("SELECT user_id FROM images WHERE id = ?", imageID).Scan(&ownerID)
	if err != nil || ownerID != user.ID {
		s.SendJSON(w, http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Not authorized to delete this image",
		})
		return
	}

	// Get image path
	var path string
	s.DB.QueryRow("SELECT path FROM images WHERE id = ?", imageID).Scan(&path)

	// Delete from database (cascade will handle likes/comments)
	_, err = s.DB.Exec("DELETE FROM images WHERE id = ?", imageID)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete image",
		})
		return
	}

	// Delete file
	if path != "" && strings.HasPrefix(path, "/static/uploads/") {
		// Convert /static/uploads/filename.jpg to ./data/uploads/filename.jpg
		filename := strings.TrimPrefix(path, "/static/uploads/")
		filePath := "./data/uploads/" + filename
		os.Remove(filePath)
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Image deleted",
	})
}

func (s *Server) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := s.GetCurrentUser(r)
	if err != nil {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	updates := []string{}
	args := []interface{}{}

	if username != "" && username != user.Username {
		// Check if username exists
		var exists int
		s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND id != ?", username, user.ID).Scan(&exists)
		if exists > 0 {
			s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Username already taken",
			})
			return
		}
		updates = append(updates, "username = ?")
		args = append(args, username)
	}

	if email != "" && email != user.Email {
		// Check if email exists
		var exists int
		s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?", email, user.ID).Scan(&exists)
		if exists > 0 {
			s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Email already taken",
			})
			return
		}
		updates = append(updates, "email = ?")
		args = append(args, email)
	}

	if password != "" {
		hash, err := auth.HashPassword(password)
		if err != nil {
			s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to hash password",
			})
			return
		}
		updates = append(updates, "password_hash = ?")
		args = append(args, hash)
	}

	if len(updates) == 0 {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "No changes provided",
		})
		return
	}

	args = append(args, user.ID)
	query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	_, err = s.DB.Exec(query, args...)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update profile",
		})
		return
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile updated",
	})
}

func (s *Server) HandleUserPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user, err := s.GetCurrentUser(r)
		if err != nil {
			s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authentication required",
			})
			return
		}

		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"comment_notifications": user.CommentNotifications,
			},
		})
		return
	}

	if r.Method == "POST" {
		user, err := s.GetCurrentUser(r)
		if err != nil {
			s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authentication required",
			})
			return
		}

		notifications := r.FormValue("comment_notifications") == "true"
		_, err = s.DB.Exec("UPDATE users SET comment_notifications = ? WHERE id = ?", notifications, user.ID)
		if err != nil {
			s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to update preferences",
			})
			return
		}

		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Preferences updated",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) HandleUserImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := s.GetCurrentUser(r)
	if err != nil {
		s.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	// Get user's images, most recent first
	rows, err := s.DB.Query(`
		SELECT id, path, created_at
		FROM images
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 20
	`, user.ID)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to load images",
		})
		return
	}
	defer rows.Close()

	var images []map[string]interface{}
	for rows.Next() {
		var img models.Image
		if err := rows.Scan(&img.ID, &img.Path, &img.CreatedAt); err != nil {
			continue
		}
		images = append(images, map[string]interface{}{
			"id":         img.ID,
			"path":       img.Path,
			"created_at": img.CreatedAt,
		})
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    images,
	})
}

