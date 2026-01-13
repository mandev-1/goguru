package main

import (
	"camagru/internal/config"
	"camagru/internal/database"
	"camagru/internal/server"
	"database/sql"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load environment variables from .env file
	config.LoadEnv(".env")

	// Initialize database
	db, err := sql.Open("sqlite3", "./data/camagru.db")
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	// Initialize database schema
	if err := database.InitDB(db); err != nil {
		os.Exit(1)
	}

	// Create server instance
	srv := &server.Server{
		DB: db,
	}

	// Setup routes
	mux := http.NewServeMux()
	srv.SetupRoutes(mux)

	// Add middleware
	handler := addMiddleware(mux)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		os.Exit(1)
	}
}

func addMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Cookie")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
