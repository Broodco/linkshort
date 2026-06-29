package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/broodco/linkshort/internal/store"
)

type RedirectHandler struct {
	store *store.Store
}

func NewRedirectHandler(s *store.Store) *RedirectHandler {
	return &RedirectHandler{store: s}
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	link, err := h.store.GetLinkBySlug(slug)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		http.NotFound(w, r)
		return
	}

	go h.store.RecordClick(link.ID, r.Referer(), r.UserAgent())

	http.Redirect(w, r, link.TargetURL, http.StatusMovedPermanently)
}
