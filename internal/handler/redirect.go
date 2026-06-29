package handler

import (
	"errors"
	"html/template"
	"net/http"
	"time"

	"github.com/broodco/linkshort/internal/store"
)

type RedirectHandler struct {
	store *store.Store
	tmpl  *template.Template
}

func NewRedirectHandler(s *store.Store, tmpl *template.Template) *RedirectHandler {
	return &RedirectHandler{store: s, tmpl: tmpl}
}

func (h *RedirectHandler) NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	_ = h.tmpl.ExecuteTemplate(w, "404.gohtml", nil)
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		h.NotFound(w)
		return
	}

	link, err := h.store.GetLinkBySlug(slug)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.NotFound(w)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		h.NotFound(w)
		return
	}

	go h.store.RecordClick(link.ID, r.Referer(), r.UserAgent())

	http.Redirect(w, r, link.TargetURL, http.StatusMovedPermanently)
}
