package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/broodco/linkshort/internal/store"
)

type AdminHandler struct {
	store *store.Store
}

func NewAdminHandler(s *store.Store) *AdminHandler {
	return &AdminHandler{store: s}
}

func (h *AdminHandler) HandleCreateLink(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Slug      string  `json:"slug"`
		TargetURL string  `json:"target_url"`
		Title     string  `json:"title"`
		ExpiresAt *string `json:"expires_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if body.Slug == "" || body.TargetURL == "" {
		http.Error(w, "slug and target_url are required", http.StatusBadRequest)
		return
	}

	var expiresAt *time.Time
	if body.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *body.ExpiresAt)
		if err != nil {
			http.Error(w, "invalid expires_at format, use RFC3339", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	link, err := h.store.CreateLink(body.Slug, body.TargetURL, body.Title, expiresAt)
	if err != nil {
		http.Error(w, "failed to create link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(link)
}

func (h *AdminHandler) HandleListLinks(w http.ResponseWriter, _ *http.Request) {
	links, err := h.store.ListLinks()
	if err != nil {
		http.Error(w, "failed to list links", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(links)
}

func (h *AdminHandler) HandleDeleteLink(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := h.store.DeleteLink(slug); err != nil {
		http.Error(w, "failed to delete link", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
