package models

import (
	"database/sql"
	"time"
)

type Session struct {
	Token     string
	UserID    int
	CreatedAt time.Time
}

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(token string, userID int) error {
	_, err := r.db.Exec(
		"INSERT INTO sessions (token, user_id, created_at) VALUES (?, ?, ?)",
		token, userID, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *SessionRepository) FindUserID(token string) (int, error) {
	var userID int
	err := r.db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", token).Scan(&userID)
	return userID, err
}

func (r *SessionRepository) Delete(token string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}
