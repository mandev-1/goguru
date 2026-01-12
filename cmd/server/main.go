package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yourusername/camagru/internal/config"
	"github.com/yourusername/camagru/internal/database"
	"github.com/yourusername/camagru/internal/handlers"
	"github.com/yourusername/camagru/internal/middleware"
	"github.com/yourusername/camagru/internal/models"
	"github.com/yourusername/camagru/internal/router"
	"github.com/yourusername/camagru/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := models.NewUserRepository(db.DB)
	sessionRepo := models.NewSessionRepository(db.DB)
	imageRepo := models.NewImageRepository(db.DB)
	assetRepo := models.NewAssetRepository(db.DB)
	commentRepo := models.NewCommentRepository(db.DB)

	// Initialize services
	emailSvc := services.NewEmailService(&cfg.SMTP)
	imageSvc := services.NewImageService()

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, emailSvc)
	galleryHandler := handlers.NewGalleryHandler(imageRepo, commentRepo, userRepo)
	editorHandler := handlers.NewEditorHandler(assetRepo, imageRepo, imageSvc)
	userHandler := handlers.NewUserHandler(userRepo, sessionRepo, emailSvc)

	// Setup router
	rt := router.New(authHandler, galleryHandler, editorHandler, userHandler, authMiddleware)
	handler := rt.Setup()

	// Start server
	addr := ":" + cfg.Server.Port
	fmt.Printf("Server starting on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
