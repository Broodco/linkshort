package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/broodco/linkshort/internal/store"
)

type AdminHandler struct {
	store   *store.Store
	tmpl    *template.Template
	baseURL string
}

func NewAdminHandler(s *store.Store, tmpl *template.Template, baseURL string) *AdminHandler {
	return &AdminHandler{store: s, tmpl: tmpl, baseURL: baseURL}
}

func (h *AdminHandler) HandlePage(w http.ResponseWriter, _ *http.Request) {
	data, err := h.buildPageData()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = h.tmpl.ExecuteTemplate(w, "admin.html", data)
}

func (h *AdminHandler) buildPageData() (*AdminPageData, error) {
	links, err := h.store.ListLinks()
	if err != nil {
		return nil, err
	}

	linkCount, _ := h.store.CountLinks()
	clickCount, _ := h.store.CountClicks()

	var views []LinkView
	for _, l := range links {
		clicks, _ := h.store.ClicksPerLink(l.ID)
		view := LinkView{
			ID:        l.ID,
			Slug:      l.Slug,
			TargetURL: l.TargetURL,
			Title:     l.Title,
			ExpiresAt: l.ExpiresAt,
			CreatedAt: l.CreatedAt,
			IsActive:  l.IsActive,
			Clicks:    clicks,
		}
		if l.IsActive && l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt) {
			view.IsExpired = true
		}
		views = append(views, view)
	}

	return &AdminPageData{
		BaseURL:    h.baseURL,
		Links:      views,
		LinkCount:  linkCount,
		ClickCount: clickCount,
	}, nil
}

func (h *AdminHandler) HandleHTMXCreate(w http.ResponseWriter, r *http.Request) {
	slug := r.FormValue("slug")
	targetURL := r.FormValue("target_url")
	title := r.FormValue("title")
	expiresAtStr := r.FormValue("expires_at")

	var expiresAt *time.Time
	if expiresAtStr != "" {
		t, err := time.Parse("2006-01-02", expiresAtStr)
		if err == nil {
			expiresAt = &t
		}
	}

	_, err := h.store.CreateLink(slug, targetURL, title, expiresAt)
	if err != nil {
		data, _ := h.buildPageData()
		data.Error = "Ce slug existe déjà ou une erreur est survenue."
		_ = h.tmpl.ExecuteTemplate(w, "links-rows", data)
		return
	}

	data, err := h.buildPageData()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = h.tmpl.ExecuteTemplate(w, "links-rows", data)
}

func (h *AdminHandler) HandleHTMXDelete(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	_ = h.store.DeleteLink(slug)
	w.WriteHeader(http.StatusOK)
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
		if err == nil {
			expiresAt = &t
		}
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

func (h *AdminHandler) HandleDeleteLink(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
