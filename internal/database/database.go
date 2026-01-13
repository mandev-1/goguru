package database

import "database/sql"

func InitDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		verified INTEGER DEFAULT 0,
		verification_token TEXT,
		reset_token TEXT,
		reset_expires DATETIME,
		session_token TEXT,
		comment_notifications INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		image_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, image_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		body TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS assets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_images_user_id ON images(user_id);
	CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at);
	CREATE INDEX IF NOT EXISTS idx_likes_image_id ON likes(image_id);
	CREATE INDEX IF NOT EXISTS idx_comments_image_id ON comments(image_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Insert default assets if they don't exist
	assets := []struct {
		name string
		path string
	}{
		{"Cat", "/static/assets/cat.png"},
		{"Cat 2", "/static/assets/cat2.png"},
		{"Caughing Cat", "/static/assets/caughing_cat.png"},
		{"Halo", "/static/assets/halo.png"},
		{"Necklace", "/static/assets/necklace.png"},
	}

	for _, asset := range assets {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM assets WHERE path = ?", asset.path).Scan(&count)
		if err != nil {
			continue
		}
		if count == 0 {
			db.Exec("INSERT INTO assets (name, path) VALUES (?, ?)", asset.name, asset.path)
		}
	}

	return nil
}

