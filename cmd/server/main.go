package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/broodco/linkshort/internal/assets"
	"github.com/broodco/linkshort/internal/handler"
	"github.com/broodco/linkshort/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/links.db"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	adminPass := os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		adminPass = "changeme"
	}

	s, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	tmpl, err := template.ParseFS(assets.WebFS, "web/templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	staticFS, err := fs.Sub(assets.WebFS, "web/static")
	if err != nil {
		log.Fatalf("failed to create static fs: %v", err)
	}

	adminHandler := handler.NewAdminHandler(s, tmpl, baseURL)
	redirectHandler := handler.NewRedirectHandler(s, tmpl)
	qrHandler := handler.NewQRHandler(s, baseURL)

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"linkshort"}`))
	})

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Admin UI
	mux.HandleFunc("GET /admin", basicAuth(adminPass, adminHandler.HandlePage))
	mux.HandleFunc("POST /admin/links", basicAuth(adminPass, adminHandler.HandleHTMXCreate))
	mux.HandleFunc("DELETE /admin/links/{slug}", basicAuth(adminPass, adminHandler.HandleHTMXDelete))

	// API JSON
	mux.HandleFunc("POST /api/links", basicAuth(adminPass, adminHandler.HandleCreateLink))
	mux.HandleFunc("GET /api/links", basicAuth(adminPass, adminHandler.HandleListLinks))
	mux.HandleFunc("DELETE /api/links/{slug}", basicAuth(adminPass, adminHandler.HandleDeleteLink))

	// Index
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_ = tmpl.ExecuteTemplate(w, "index.html", nil)
	})

	// Redirection
	mux.Handle("GET /{slug}", redirectHandler)

	// QR Codes
	mux.Handle("GET /qr/{slug}", qrHandler)

	log.Printf("starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func basicAuth(password string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, pass, ok := r.BasicAuth()
		if !ok || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="linkshort"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
