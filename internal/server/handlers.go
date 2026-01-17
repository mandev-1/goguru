package server

import (
	"camagru/internal/auth"
	"camagru/internal/models"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (s *Server) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	user, err := s.GetCurrentUser(r)
	if err != nil || user == nil || !user.Verified {
		http.Redirect(w, r, "/gallery", http.StatusFound)
		return
	}
	http.ServeFile(w, r, "./web/static/pages/home.html")
}

func (s *Server) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user, err := s.GetCurrentUser(r)
		if err == nil && user != nil && user.Verified {
			http.Redirect(w, r, "/gallery", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./web/static/pages/login.html")
		return
	}
	if r.Method == "POST" {
		s.HandleLogin(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user, err := s.GetCurrentUser(r)
		if err == nil && user != nil && user.Verified {
			http.Redirect(w, r, "/gallery", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./web/static/pages/register.html")
		return
	}
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
	assets, err := s.DB.GetAssets()
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to load assets",
		})
		return
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
	images, total, err := s.DB.GetImagesPaginated(page, limit)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to load gallery",
		})
		return
	}
	user, _ := s.GetCurrentUser(r)
	if user != nil && len(images) > 0 {
		imageIDs := make([]int, 0, len(images))
		for _, img := range images {
			imageIDs = append(imageIDs, img.ID)
		}
		likedMap, err := s.DB.GetLikedImageIDs(user.ID, imageIDs)
		if err == nil {
			for i := range images {
				images[i].Liked = likedMap[images[i].ID]
			}
		}
	}
	for i := range images {
		comments, err := s.DB.GetCommentsByImageID(images[i].ID, 10)
		if err == nil {
			images[i].Comments = comments
		}
	}

	hasMore := (page-1)*limit+len(images) < total
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	s.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"items":       images,
			"hasMore":     hasMore,
			"totalPages":  totalPages,
			"currentPage": page,
			"total":       total,
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
	exists, err := s.DB.LikeExists(user.ID, imageID)
	if err == nil && exists {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Already liked",
		})
		return
	}
	err = s.DB.CreateLike(user.ID, imageID)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to like image",
		})
		return
	}
	img, err := s.DB.GetImageByID(imageID)
	if err != nil {
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Image liked",
		})
		return
	}

	imageOwner, err := s.DB.GetUserByID(img.UserID)
	if err != nil {
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Image liked",
		})
		return
	}

	imageOwnerID := imageOwner.ID
	imageOwnerEmail := imageOwner.Email
	imageOwnerNotifications := imageOwner.CommentNotifications

	if imageOwnerID != user.ID && imageOwnerNotifications {
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
	_, err = s.DB.CreateComment(imageID, user.ID, body)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to add comment",
		})
		return
	}
	img, err := s.DB.GetImageByID(imageID)
	if err != nil {
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Comment added",
		})
		return
	}

	imageOwner, err := s.DB.GetUserByID(img.UserID)
	if err != nil {
		s.SendJSON(w, http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Comment added",
		})
		return
	}

	imageOwnerID := imageOwner.ID
	imageOwnerEmail := imageOwner.Email
	imageOwnerNotifications := imageOwner.CommentNotifications

	if imageOwnerID != user.ID && imageOwnerNotifications {
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
	ownerID, err := s.DB.GetImageOwner(imageID)
	if err != nil || ownerID != user.ID {
		s.SendJSON(w, http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Not authorized to delete this image",
		})
		return
	}
	img, err := s.DB.GetImageByID(imageID)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete image",
		})
		return
	}
	path := img.Path
	err = s.DB.DeleteImage(imageID)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete image",
		})
		return
	}
	if path != "" && strings.HasPrefix(path, "/static/uploads/") {
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

	updateUsername := ""
	updateEmail := ""
	updatePassword := ""

	if username != "" && username != user.Username {
		usernameExists, _, err := s.DB.UserExists(username, "")
		if err != nil {
			s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			return
		}
		if usernameExists {
			existingUser, err := s.DB.GetUserByUsernameOrEmail(username)
			if err == nil && existingUser.ID != user.ID {
				s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
					Success: false,
					Message: "Username already taken",
				})
				return
			}
		}
		updateUsername = username
	}

	if email != "" && email != user.Email {
		_, emailExists, err := s.DB.UserExists("", email)
		if err != nil {
			s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			return
		}
		if emailExists {
			existingUser, err := s.DB.GetUserByUsernameOrEmail(email)
			if err == nil && existingUser.ID != user.ID {
				s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
					Success: false,
					Message: "Email already taken",
				})
				return
			}
		}
		updateEmail = email
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
		updatePassword = hash
	}

	if updateUsername == "" && updateEmail == "" && updatePassword == "" {
		s.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "No changes provided",
		})
		return
	}

	err = s.DB.UpdateUser(user.ID, updateUsername, updateEmail, updatePassword)
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
		err = s.DB.UpdateUserPreferences(user.ID, notifications)
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
	userImages, err := s.DB.GetUserImages(user.ID, 20)
	if err != nil {
		s.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to load images",
		})
		return
	}

	images := make([]map[string]interface{}, 0, len(userImages))
	for _, img := range userImages {
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
