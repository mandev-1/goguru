package models

import "time"

type User struct {
	ID                   int
	Username             string
	Email                string
	PasswordHash         string
	Verified             bool
	VerificationToken    string
	ResetToken           string
	ResetExpires         *time.Time
	CommentNotifications bool
	CreatedAt            time.Time
}

type Image struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
	Author    string    `json:"author"`
	Likes     int       `json:"likes"`
	Liked     bool      `json:"liked"`
	Comments  []Comment `json:"comments"`
}

type Comment struct {
	ID        int       `json:"id"`
	ImageID   int       `json:"image_id"`
	UserID    int       `json:"user_id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type Asset struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

