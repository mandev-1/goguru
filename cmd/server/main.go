package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var templates *template.Template
var dev bool

func init() {
	// Determine environment once at startup
	dev = os.Getenv("ENV") == "development"
	// In production, parse and cache templates once
	if !dev {
		templates = template.Must(template.ParseGlob("web/templates/*.html"))
	}
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
	fmt.Printf("Server starting on http://localhost%s (dev=%v)\n", port, dev)
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
	// In development, re-parse templates on each request so changes appear on refresh
	var t *template.Template
	var err error
	if dev {
		t, err = template.ParseGlob("web/templates/*.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Template parse error: %v", err)
			return
		}
	} else {
		t = templates
	}

	if err = t.ExecuteTemplate(w, tmpl, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Template exec error: %v", err)
	}
}
