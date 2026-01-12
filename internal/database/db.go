package database

import (
	"database/sql"
	"fmt"

	"github.com/yourusername/camagru/internal/config"
	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func New(cfg *config.Config) (*DB, error) {
	var db *sql.DB
	var err error

	if cfg.Database.Host == "" {
		// Use SQLite
		db, err = sql.Open("sqlite", "file:camagru.db?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)")
	} else {
		// Use PostgreSQL
		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)
		db, err = sql.Open("postgres", psqlInfo)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	wrapper := &DB{db}
	if err := wrapper.InitSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return wrapper, nil
}
