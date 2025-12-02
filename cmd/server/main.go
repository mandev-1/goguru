package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var templates *template.Template

func init() {
	// Parse all templates (use project-relative path, not "../../...")
	templates = template.Must(template.ParseGlob("web/templates/*.html"))
}

func main() {
	// Serve static files (CSS, JS, images)
	fs := http.FileServer(http.Dir("web/static"))
	// Use URL prefix here ("/static/"), not a filesystem path
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/gallery", galleryHandler)
	http.HandleFunc("/editor", editorHandler)

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "home.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login.html", nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "register.html", nil)
}

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "gallery.html", nil)
}

func editorHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "editor.html", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}
