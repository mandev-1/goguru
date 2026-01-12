package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Verified     bool
	VerifyToken  string
	CreatedAt    time.Time
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(username, email, passwordHash, verifyToken string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (username, email, password_hash, verified, verify_token, created_at)
		 VALUES (?, ?, ?, 0, ?, ?)`,
		username, email, passwordHash, verifyToken, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepository) FindByUsername(username string) (*User, error) {
	var u User
	var verified int
	err := r.db.QueryRow(
		"SELECT id, username, email, password_hash, verified, verify_token, created_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &verified, &u.VerifyToken, &u.CreatedAt)

	if err != nil {
		return nil, err
	}
	u.Verified = verified == 1
	return &u, nil
}

func (r *UserRepository) FindByEmail(email string) (*User, error) {
	var u User
	var verified int
	err := r.db.QueryRow(
		"SELECT id, username, email, password_hash, verified, verify_token, created_at FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &verified, &u.VerifyToken, &u.CreatedAt)

	if err != nil {
		return nil, err
	}
	u.Verified = verified == 1
	return &u, nil
}

func (r *UserRepository) FindByID(id int) (*User, error) {
	var u User
	var verified int
	err := r.db.QueryRow(
		"SELECT id, username, email, password_hash, verified, verify_token, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &verified, &u.VerifyToken, &u.CreatedAt)

	if err != nil {
		return nil, err
	}
	u.Verified = verified == 1
	return &u, nil
}

func (r *UserRepository) Exists(field, value string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(1) FROM users WHERE "+field+" = ?", value).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) Verify(token string) error {
	_, err := r.db.Exec("UPDATE users SET verified = 1, verify_token = NULL WHERE verify_token = ?", token)
	return err
}

func (r *UserRepository) UpdatePassword(userID int, newHash string) error {
	_, err := r.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", newHash, userID)
	return err
}

func (r *UserRepository) SetVerifyToken(userID int, token string) error {
	_, err := r.db.Exec("UPDATE users SET verify_token = ? WHERE id = ?", token, userID)
	return err
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
