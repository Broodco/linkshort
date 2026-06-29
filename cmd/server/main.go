package main

import (
	"log"
	"net/http"
	"os"

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

	// Initialise le store
	s, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	// Handlers
	redirectHandler := handler.NewRedirectHandler(s)
	adminHandler := handler.NewAdminHandler(s)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"linkshort"}`))
	})

	// Redirection
	mux.Handle("GET /{slug}", redirectHandler)

	// Admin
	adminPass := os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		adminPass = "changeme"
	}

	mux.HandleFunc("POST /api/links", basicAuth(adminPass, adminHandler.HandleCreateLink))
	mux.HandleFunc("GET /api/links", basicAuth(adminPass, adminHandler.HandleListLinks))
	mux.HandleFunc("DELETE /api/links/{slug}", basicAuth(adminPass, adminHandler.HandleDeleteLink))
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
