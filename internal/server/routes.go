package server

import (
	"net/http"
)

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	uploadsFS := http.FileServer(http.Dir("./data/uploads"))
	mux.Handle("/static/uploads/", http.StripPrefix("/static/uploads/", uploadsFS))
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", s.HandleHome)
	mux.HandleFunc("/login", s.HandleLoginPage)
	mux.HandleFunc("/register", s.HandleRegisterPage)
	mux.HandleFunc("/gallery", s.HandleGalleryPage)
	mux.HandleFunc("/editor", s.RequireAuth(s.HandleEditorPage))
	mux.HandleFunc("/user", s.RequireAuth(s.HandleUserPage))
	mux.HandleFunc("/password", s.HandlePasswordPage)
	mux.HandleFunc("/forgot-password", s.HandleForgotPasswordPage)
	mux.HandleFunc("/unauthorized", s.HandleUnauthorizedPage)
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
	mux.HandleFunc("/logout", s.HandleLogout)
	mux.HandleFunc("/verify", s.HandleVerify)
	mux.HandleFunc("/reset-password", s.HandleResetPassword)
	mux.HandleFunc("/resend-verification", s.HandleResendVerification)
}
