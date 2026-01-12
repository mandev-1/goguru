package router

import (
	"net/http"

	"github.com/yourusername/camagru/internal/handlers"
	"github.com/yourusername/camagru/internal/middleware"
)

type Router struct {
	authHandler    *handlers.AuthHandler
	galleryHandler *handlers.GalleryHandler
	editorHandler  *handlers.EditorHandler
	userHandler    *handlers.UserHandler
	authMiddleware *middleware.AuthMiddleware
}

func New(
	authHandler *handlers.AuthHandler,
	galleryHandler *handlers.GalleryHandler,
	editorHandler *handlers.EditorHandler,
	userHandler *handlers.UserHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		authHandler:    authHandler,
		galleryHandler: galleryHandler,
		editorHandler:  editorHandler,
		userHandler:    userHandler,
		authMiddleware: authMiddleware,
	}
}

func (rt *Router) Setup() http.Handler {
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public pages
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rt.authHandler.LoginPage(w, r)
		} else if r.Method == http.MethodPost {
			rt.authHandler.Login(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rt.authHandler.RegisterPage(w, r)
		} else if r.Method == http.MethodPost {
			rt.authHandler.Register(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/verify", rt.authHandler.Verify)
	mux.HandleFunc("/resend-verification", rt.authHandler.ResendVerification)
	mux.HandleFunc("/forgot-password", rt.userHandler.ForgotPasswordPage)
	mux.HandleFunc("/reset-password", rt.userHandler.ResetPasswordPage)

	// Protected pages
	mux.HandleFunc("/gallery", rt.authMiddleware.RequireAuth(rt.galleryHandler.GalleryPage))
	mux.HandleFunc("/editor", rt.authMiddleware.RequireAuth(rt.editorHandler.EditorPage))
	mux.HandleFunc("/user", rt.authMiddleware.RequireAuth(rt.userHandler.UserPage))
	mux.HandleFunc("/logout", rt.authMiddleware.RequireAuth(rt.authHandler.Logout))

	// Gallery API (some public, some protected)
	mux.HandleFunc("/api/gallery", rt.galleryHandler.List)
	mux.HandleFunc("/api/gallery/like", rt.authMiddleware.RequireAuth(rt.galleryHandler.Like))
	mux.HandleFunc("/api/gallery/comment", rt.authMiddleware.RequireAuth(rt.galleryHandler.Comment))
	mux.HandleFunc("/api/gallery/mock-upload", rt.authMiddleware.RequireAuth(rt.galleryHandler.MockUpload))

	// Editor API (protected)
	mux.HandleFunc("/api/assets", rt.authMiddleware.RequireAuth(rt.editorHandler.ListAssets))
	mux.HandleFunc("/api/assets/upload", rt.authMiddleware.RequireAuth(rt.editorHandler.UploadAsset))
	mux.HandleFunc("/api/compose", rt.authMiddleware.RequireAuth(rt.editorHandler.Compose))

	// User API (protected)
	mux.HandleFunc("/api/current-user", rt.authMiddleware.RequireAuth(rt.userHandler.CurrentUser))

	return mux
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/static/pages/home.html")
}
