package server

import (
	"net/http"
)

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	// Serve uploads from data/uploads (but keep URL as /static/uploads/)
	// This must come before the general /static/ handler
	uploadsFS := http.FileServer(http.Dir("./data/uploads"))
	mux.Handle("/static/uploads/", http.StripPrefix("/static/uploads/", uploadsFS))
	
	// Static files (everything else under /static/)
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// HTML pages
	mux.HandleFunc("/", s.HandleHome)
	mux.HandleFunc("/login", s.HandleLoginPage)
	mux.HandleFunc("/register", s.HandleRegisterPage)
	mux.HandleFunc("/gallery", s.HandleGalleryPage)
	mux.HandleFunc("/editor", s.RequireAuth(s.HandleEditorPage))
	mux.HandleFunc("/user", s.RequireAuth(s.HandleUserPage))
	mux.HandleFunc("/password", s.HandlePasswordPage)
	mux.HandleFunc("/forgot-password", s.HandleForgotPasswordPage)
	mux.HandleFunc("/unauthorized", s.HandleUnauthorizedPage)

	// API endpoints
	mux.HandleFunc("/api/current-user", s.HandleCurrentUser)
	mux.HandleFunc("/api/assets", s.HandleAssets)
	mux.HandleFunc("/api/compose", s.RequireAuth(s.HandleCompose))
	mux.HandleFunc("/api/gallery", s.HandleGallery)
	mux.HandleFunc("/api/gallery/like", s.RequireAuth(s.HandleLike))
	mux.HandleFunc("/api/gallery/comment", s.RequireAuth(s.HandleComment))
	mux.HandleFunc("/api/gallery/delete", s.RequireAuth(s.HandleDeleteImage))
	mux.HandleFunc("/api/user/images", s.RequireAuth(s.HandleUserImages))
	mux.HandleFunc("/api/user/update", s.RequireAuth(s.HandleUpdateUser))
	mux.HandleFunc("/api/user/preferences", s.RequireAuth(s.HandleUserPreferences))

	// Auth endpoints (POST only)
	mux.HandleFunc("/logout", s.HandleLogout)
	mux.HandleFunc("/verify", s.HandleVerify)
	mux.HandleFunc("/reset-password", s.HandleResetPassword)
	mux.HandleFunc("/resend-verification", s.HandleResendVerification)
}

