package router

import (
	"database/sql"
	"camagru/internal/handlers"
	"camagru/internal/handlers/auth"
	"camagru/internal/handlers/image"
	"camagru/internal/middleware"
	"net/http"
)

func SetupRoutes(db *sql.DB) {
	// Pass the database connection to the packages that need it.
	handlers.SetDB(db)
	auth.SetDB(db)
	image.SetDB(db)
	middleware.SetDB(db)

	// Serve static files (CSS, JS, images)
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Static pages
	http.HandleFunc("/", handlers.HomeHandler)

	// Auth
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/logout", middleware.AuthRequired(handlers.LogoutHandler))

	// Authenticated pages
	http.HandleFunc("/gallery", middleware.AuthRequired(handlers.GalleryHandler))
	http.HandleFunc("/editor", middleware.AuthRequired(handlers.EditorHandler))
	http.HandleFunc("/user", middleware.AuthRequired(handlers.UserHandler))

	// API
	http.HandleFunc("/api/current-user", middleware.AuthRequired(handlers.CurrentUserAPIHandler))

	// Auth APIs
	http.HandleFunc("/resend-verification", auth.ResendVerificationHandler)
	http.HandleFunc("/verify", auth.VerifyHandler)
	http.HandleFunc("/forgot-password", auth.ForgotPasswordHandler)
	http.HandleFunc("/reset-password", auth.ResetPasswordHandler)

	// Image APIs
	http.HandleFunc("/api/assets", middleware.AuthRequired(image.AssetsHandler))
	http.HandleFunc("/api/assets/upload", middleware.AuthRequired(image.UploadAssetHandler))
	http.HandleFunc("/api/compose", middleware.AuthRequired(image.ComposeHandler))
	http.HandleFunc("/api/gallery/like", middleware.AuthRequired(image.LikeImageHandler))
	http.HandleFunc("/api/gallery/comment", middleware.AuthRequired(image.CommentImageHandler))
}
