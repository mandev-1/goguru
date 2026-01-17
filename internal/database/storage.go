package database

import (
	"camagru/internal/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Storage struct {
	dataDir string
	mu      sync.RWMutex
}

func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &Storage{dataDir: dataDir}, nil
}

type userRecord struct {
	ID                   int        `json:"id"`
	Username             string     `json:"username"`
	Email                string     `json:"email"`
	PasswordHash         string     `json:"password_hash"`
	Verified             bool       `json:"verified"`
	VerificationToken    string     `json:"verification_token"`
	ResetToken           string     `json:"reset_token"`
	ResetExpires         *time.Time `json:"reset_expires"`
	SessionToken         string     `json:"session_token"`
	CommentNotifications bool       `json:"comment_notifications"`
	CreatedAt            time.Time  `json:"created_at"`
}

func (s *Storage) getUsers() (map[int]*userRecord, error) {
	path := filepath.Join(s.dataDir, "users.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[int]*userRecord), nil
	}
	if err != nil {
		return nil, err
	}
	var users map[int]*userRecord
	if len(data) == 0 {
		return make(map[int]*userRecord), nil
	}
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Storage) saveUsers(users map[int]*userRecord) error {
	path := filepath.Join(s.dataDir, "users.json")
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

type imageRecord struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Storage) getImages() (map[int]*imageRecord, error) {
	path := filepath.Join(s.dataDir, "images.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[int]*imageRecord), nil
	}
	if err != nil {
		return nil, err
	}
	var images map[int]*imageRecord
	if len(data) == 0 {
		return make(map[int]*imageRecord), nil
	}
	if err := json.Unmarshal(data, &images); err != nil {
		return nil, err
	}
	return images, nil
}

func (s *Storage) saveImages(images map[int]*imageRecord) error {
	path := filepath.Join(s.dataDir, "images.json")
	data, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

type likeRecord struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ImageID   int       `json:"image_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Storage) getLikes() (map[int]*likeRecord, error) {
	path := filepath.Join(s.dataDir, "likes.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[int]*likeRecord), nil
	}
	if err != nil {
		return nil, err
	}
	var likes map[int]*likeRecord
	if len(data) == 0 {
		return make(map[int]*likeRecord), nil
	}
	if err := json.Unmarshal(data, &likes); err != nil {
		return nil, err
	}
	return likes, nil
}

func (s *Storage) saveLikes(likes map[int]*likeRecord) error {
	path := filepath.Join(s.dataDir, "likes.json")
	data, err := json.MarshalIndent(likes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

type commentRecord struct {
	ID        int       `json:"id"`
	ImageID   int       `json:"image_id"`
	UserID    int       `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Storage) getComments() (map[int]*commentRecord, error) {
	path := filepath.Join(s.dataDir, "comments.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[int]*commentRecord), nil
	}
	if err != nil {
		return nil, err
	}
	var comments map[int]*commentRecord
	if len(data) == 0 {
		return make(map[int]*commentRecord), nil
	}
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *Storage) saveComments(comments map[int]*commentRecord) error {
	path := filepath.Join(s.dataDir, "comments.json")
	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

type assetRecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (s *Storage) getAssets() (map[int]*assetRecord, error) {
	path := filepath.Join(s.dataDir, "assets.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[int]*assetRecord), nil
	}
	if err != nil {
		return nil, err
	}
	var assets map[int]*assetRecord
	if len(data) == 0 {
		return make(map[int]*assetRecord), nil
	}
	if err := json.Unmarshal(data, &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

func (s *Storage) saveAssets(assets map[int]*assetRecord) error {
	path := filepath.Join(s.dataDir, "assets.json")
	data, err := json.MarshalIndent(assets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

type idCounters struct {
	UserID    int `json:"user_id"`
	ImageID   int `json:"image_id"`
	LikeID    int `json:"like_id"`
	CommentID int `json:"comment_id"`
	AssetID   int `json:"asset_id"`
}

func (s *Storage) getIDCounters() (*idCounters, error) {
	path := filepath.Join(s.dataDir, "ids.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &idCounters{}, nil
	}
	if err != nil {
		return nil, err
	}
	var counters idCounters
	if len(data) == 0 {
		return &idCounters{}, nil
	}
	if err := json.Unmarshal(data, &counters); err != nil {
		return nil, err
	}
	return &counters, nil
}

func (s *Storage) saveIDCounters(counters *idCounters) error {
	path := filepath.Join(s.dataDir, "ids.json")
	data, err := json.MarshalIndent(counters, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) InitDB() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	assets, err := s.getAssets()
	if err != nil {
		return err
	}

	if len(assets) == 0 {
		counters, err := s.getIDCounters()
		if err != nil {
			return err
		}

		defaultAssets := []struct {
			name string
			path string
		}{
			{"Cat", "/static/assets/cat.png"},
			{"Cat 2", "/static/assets/cat2.png"},
			{"Caughing Cat", "/static/assets/caughing_cat.png"},
			{"Halo", "/static/assets/halo.png"},
			{"Necklace", "/static/assets/necklace.png"},
		}

		for _, asset := range defaultAssets {
			counters.AssetID++
			assets[counters.AssetID] = &assetRecord{
				ID:   counters.AssetID,
				Name: asset.name,
				Path: asset.path,
			}
		}

		if err := s.saveAssets(assets); err != nil {
			return err
		}
		if err := s.saveIDCounters(counters); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) GetUserByID(id int) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.getUsers()
	if err != nil {
		return nil, err
	}

	user, exists := users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return &models.User{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		PasswordHash:         user.PasswordHash,
		Verified:             user.Verified,
		VerificationToken:    user.VerificationToken,
		ResetToken:           user.ResetToken,
		ResetExpires:         user.ResetExpires,
		CommentNotifications: user.CommentNotifications,
		CreatedAt:            user.CreatedAt,
	}, nil
}

func (s *Storage) GetUserByUsernameOrEmail(usernameOrEmail string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.getUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == usernameOrEmail || user.Email == usernameOrEmail {
			return &models.User{
				ID:                   user.ID,
				Username:             user.Username,
				Email:                user.Email,
				PasswordHash:         user.PasswordHash,
				Verified:             user.Verified,
				VerificationToken:    user.VerificationToken,
				ResetToken:           user.ResetToken,
				ResetExpires:         user.ResetExpires,
				CommentNotifications: user.CommentNotifications,
				CreatedAt:            user.CreatedAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (s *Storage) GetUserBySessionToken(token string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.getUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.SessionToken == token {
			return &models.User{
				ID:                   user.ID,
				Username:             user.Username,
				Email:                user.Email,
				PasswordHash:         user.PasswordHash,
				Verified:             user.Verified,
				VerificationToken:    user.VerificationToken,
				ResetToken:           user.ResetToken,
				ResetExpires:         user.ResetExpires,
				CommentNotifications: user.CommentNotifications,
				CreatedAt:            user.CreatedAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (s *Storage) UserExists(username, email string) (bool, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.getUsers()
	if err != nil {
		return false, false, err
	}

	usernameExists := false
	emailExists := false

	for _, user := range users {
		if user.Username == username {
			usernameExists = true
		}
		if user.Email == email {
			emailExists = true
		}
	}

	return usernameExists, emailExists, nil
}

func (s *Storage) CreateUser(username, email, passwordHash, verificationToken string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.getUsers()
	if err != nil {
		return 0, err
	}

	counters, err := s.getIDCounters()
	if err != nil {
		return 0, err
	}

	counters.UserID++
	userID := counters.UserID

	users[userID] = &userRecord{
		ID:                   userID,
		Username:             username,
		Email:                email,
		PasswordHash:         passwordHash,
		Verified:             false,
		VerificationToken:    verificationToken,
		CommentNotifications: true,
		CreatedAt:            time.Now(),
	}

	if err := s.saveUsers(users); err != nil {
		return 0, err
	}
	if err := s.saveIDCounters(counters); err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *Storage) UpdateUserSessionToken(userID int, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.getUsers()
	if err != nil {
		return err
	}

	user, exists := users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.SessionToken = token
	if err := s.saveUsers(users); err != nil {
		return err
	}

	return nil
}

func (s *Storage) ClearUserSessionToken(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.getUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.SessionToken == token {
			user.SessionToken = ""
			if err := s.saveUsers(users); err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func (s *Storage) VerifyUser(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.getUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.VerificationToken == token {
			user.Verified = true
			user.VerificationToken = ""
			if err := s.saveUsers(users); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("invalid verification token")
}

func (s *Storage) GetUserByVerificationToken(token string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.getUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.VerificationToken == token {
			return &models.User{
				ID:                   user.ID,
				Username:             user.Username,
				Email:                user.Email,
				PasswordHash:         user.PasswordHash,
				Verified:             user.Verified,
				VerificationToken:    user.VerificationToken,
				ResetToken:           user.ResetToken,
				ResetExpires:         user.ResetExpires,
				CommentNotifications: user.CommentNotifications,
				CreatedAt:            user.CreatedAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (s *Storage) SetPasswordResetToken(email string, token string, expires time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.getUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Email == email {
			user.ResetToken = token
			user.ResetExpires = &expires
			if err := s.saveUsers(users); err != nil {
				return err
			}
			return nil
		}
	}

	return nil // Don't reveal if email exists
}

func (s *Storage) GetUserByResetToken(token string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.getUsers()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, user := range users {
		if user.ResetToken == token {
			if user.ResetExpires != nil && now.After(*user.ResetExpires) {
				return nil, fmt.Errorf("token expired")
			}
			return &models.User{
				ID:                   user.ID,
				Username:             user.Username,
				Email:                user.Email,
				PasswordHash:         user.PasswordHash,
				Verified:             user.Verified,
				VerificationToken:    user.VerificationToken,
				ResetToken:           user.ResetToken,
				ResetExpires:         user.ResetExpires,
				CommentNotifications: user.CommentNotifications,
				CreatedAt:            user.CreatedAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid reset token")
}

func (s *Storage) UpdateUserPassword(userID int, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.getUsers()
	if err != nil {
		return err
	}

	user, exists := users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.PasswordHash = passwordHash
	user.ResetToken = ""
	user.ResetExpires = nil

	if err := s.saveUsers(users); err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateUser(userID int, username, email, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.getUsers()
	if err != nil {
		return err
	}

	user, exists := users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	if username != "" {
		user.Username = username
	}
	if email != "" {
		user.Email = email
	}
	if passwordHash != "" {
		user.PasswordHash = passwordHash
	}

	if err := s.saveUsers(users); err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateUserPreferences(userID int, commentNotifications bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.getUsers()
	if err != nil {
		return err
	}

	user, exists := users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.CommentNotifications = commentNotifications

	if err := s.saveUsers(users); err != nil {
		return err
	}

	return nil
}

func (s *Storage) CreateImage(userID int, path string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	images, err := s.getImages()
	if err != nil {
		return 0, err
	}

	counters, err := s.getIDCounters()
	if err != nil {
		return 0, err
	}

	counters.ImageID++
	imageID := counters.ImageID

	images[imageID] = &imageRecord{
		ID:        imageID,
		UserID:    userID,
		Path:      path,
		CreatedAt: time.Now(),
	}

	if err := s.saveImages(images); err != nil {
		return 0, err
	}
	if err := s.saveIDCounters(counters); err != nil {
		return 0, err
	}

	return imageID, nil
}

func (s *Storage) GetImageByID(id int) (*models.Image, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	images, err := s.getImages()
	if err != nil {
		return nil, err
	}

	image, exists := images[id]
	if !exists {
		return nil, fmt.Errorf("image not found")
	}

	users, err := s.getUsers()
	if err != nil {
		return nil, err
	}

	user, exists := users[image.UserID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return &models.Image{
		ID:        image.ID,
		UserID:    image.UserID,
		Path:      image.Path,
		CreatedAt: image.CreatedAt,
		Author:    user.Username,
	}, nil
}

func (s *Storage) GetImagesPaginated(page, limit int) ([]models.Image, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	images, err := s.getImages()
	if err != nil {
		return nil, 0, err
	}

	users, err := s.getUsers()
	if err != nil {
		return nil, 0, err
	}

	likes, err := s.getLikes()
	if err != nil {
		return nil, 0, err
	}
	imageList := make([]*imageRecord, 0, len(images))
	for _, img := range images {
		imageList = append(imageList, img)
	}

	sort.Slice(imageList, func(i, j int) bool {
		return imageList[i].CreatedAt.After(imageList[j].CreatedAt)
	})

	total := len(imageList)
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	if offset > len(imageList) {
		offset = len(imageList)
	}

	end := offset + limit
	if end > len(imageList) {
		end = len(imageList)
	}

	result := make([]models.Image, 0, end-offset)
	for i := offset; i < end; i++ {
		img := imageList[i]
		user, exists := users[img.UserID]
		if !exists {
			continue
		}
		likeCount := 0
		for _, like := range likes {
			if like.ImageID == img.ID {
				likeCount++
			}
		}

		result = append(result, models.Image{
			ID:        img.ID,
			UserID:    img.UserID,
			Path:      img.Path,
			CreatedAt: img.CreatedAt,
			Author:    user.Username,
			Likes:     likeCount,
		})
	}

	return result, total, nil
}

func (s *Storage) GetUserImages(userID int, limit int) ([]models.Image, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	images, err := s.getImages()
	if err != nil {
		return nil, err
	}

	imageList := make([]*imageRecord, 0)
	for _, img := range images {
		if img.UserID == userID {
			imageList = append(imageList, img)
		}
	}

	sort.Slice(imageList, func(i, j int) bool {
		return imageList[i].CreatedAt.After(imageList[j].CreatedAt)
	})

	if limit > 0 && limit < len(imageList) {
		imageList = imageList[:limit]
	}

	result := make([]models.Image, 0, len(imageList))
	for _, img := range imageList {
		result = append(result, models.Image{
			ID:        img.ID,
			UserID:    img.UserID,
			Path:      img.Path,
			CreatedAt: img.CreatedAt,
		})
	}

	return result, nil
}

func (s *Storage) DeleteImage(imageID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	images, err := s.getImages()
	if err != nil {
		return err
	}

	_, exists := images[imageID]
	if !exists {
		return fmt.Errorf("image not found")
	}

	delete(images, imageID)
	likes, err := s.getLikes()
	if err != nil {
		return err
	}

	for id, like := range likes {
		if like.ImageID == imageID {
			delete(likes, id)
		}
	}

	comments, err := s.getComments()
	if err != nil {
		return err
	}

	for id, comment := range comments {
		if comment.ImageID == imageID {
			delete(comments, id)
		}
	}

	if err := s.saveImages(images); err != nil {
		return err
	}
	if err := s.saveLikes(likes); err != nil {
		return err
	}
	if err := s.saveComments(comments); err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetImageOwner(imageID int) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	images, err := s.getImages()
	if err != nil {
		return 0, err
	}

	img, exists := images[imageID]
	if !exists {
		return 0, fmt.Errorf("image not found")
	}

	return img.UserID, nil
}

func (s *Storage) LikeExists(userID, imageID int) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	likes, err := s.getLikes()
	if err != nil {
		return false, err
	}

	for _, like := range likes {
		if like.UserID == userID && like.ImageID == imageID {
			return true, nil
		}
	}

	return false, nil
}

func (s *Storage) CreateLike(userID, imageID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	likes, err := s.getLikes()
	if err != nil {
		return err
	}
	for _, like := range likes {
		if like.UserID == userID && like.ImageID == imageID {
			return fmt.Errorf("already liked")
		}
	}

	counters, err := s.getIDCounters()
	if err != nil {
		return err
	}

	counters.LikeID++
	likeID := counters.LikeID

	likes[likeID] = &likeRecord{
		ID:        likeID,
		UserID:    userID,
		ImageID:   imageID,
		CreatedAt: time.Now(),
	}

	if err := s.saveLikes(likes); err != nil {
		return err
	}
	if err := s.saveIDCounters(counters); err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetLikedImageIDs(userID int, imageIDs []int) (map[int]bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	likes, err := s.getLikes()
	if err != nil {
		return nil, err
	}

	result := make(map[int]bool)
	for _, like := range likes {
		if like.UserID == userID {
			for _, imgID := range imageIDs {
				if like.ImageID == imgID {
					result[imgID] = true
				}
			}
		}
	}

	return result, nil
}

func (s *Storage) CreateComment(imageID, userID int, body string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	comments, err := s.getComments()
	if err != nil {
		return 0, err
	}

	counters, err := s.getIDCounters()
	if err != nil {
		return 0, err
	}

	counters.CommentID++
	commentID := counters.CommentID

	comments[commentID] = &commentRecord{
		ID:        commentID,
		ImageID:   imageID,
		UserID:    userID,
		Body:      body,
		CreatedAt: time.Now(),
	}

	if err := s.saveComments(comments); err != nil {
		return 0, err
	}
	if err := s.saveIDCounters(counters); err != nil {
		return 0, err
	}

	return commentID, nil
}

func (s *Storage) GetCommentsByImageID(imageID int, limit int) ([]models.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	comments, err := s.getComments()
	if err != nil {
		return nil, err
	}

	users, err := s.getUsers()
	if err != nil {
		return nil, err
	}

	commentList := make([]*commentRecord, 0)
	for _, comment := range comments {
		if comment.ImageID == imageID {
			commentList = append(commentList, comment)
		}
	}

	sort.Slice(commentList, func(i, j int) bool {
		return commentList[i].CreatedAt.After(commentList[j].CreatedAt)
	})

	if limit > 0 && limit < len(commentList) {
		commentList = commentList[:limit]
	}

	result := make([]models.Comment, 0, len(commentList))
	for _, comment := range commentList {
		user, exists := users[comment.UserID]
		if !exists {
			continue
		}

		result = append(result, models.Comment{
			ID:        comment.ID,
			ImageID:   comment.ImageID,
			UserID:    comment.UserID,
			Author:    user.Username,
			Body:      comment.Body,
			CreatedAt: comment.CreatedAt,
		})
	}

	return result, nil
}

func (s *Storage) GetAssets() ([]models.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assets, err := s.getAssets()
	if err != nil {
		return nil, err
	}

	result := make([]models.Asset, 0, len(assets))
	for _, asset := range assets {
		result = append(result, models.Asset{
			ID:   asset.ID,
			Name: asset.Name,
			Path: asset.Path,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func (s *Storage) GetAssetByID(id int) (*models.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assets, err := s.getAssets()
	if err != nil {
		return nil, err
	}

	asset, exists := assets[id]
	if !exists {
		return nil, fmt.Errorf("asset not found")
	}

	return &models.Asset{
		ID:   asset.ID,
		Name: asset.Name,
		Path: asset.Path,
	}, nil
}
