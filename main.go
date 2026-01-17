package main

import (
	"camagru/internal/config"
	"camagru/internal/database"
	"camagru/internal/server"
	"net/http"
	"os"
)

func main() {
	config.LoadEnv(".env")
	storage, err := database.NewStorage("./data")
	if err != nil {
		os.Exit(1)
	}
	if err := storage.InitDB(); err != nil {
		os.Exit(1)
	}
	srv := &server.Server{
		DB: storage,
	}
	mux := http.NewServeMux()
	srv.SetupRoutes(mux)
	handler := addMiddleware(mux)
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
