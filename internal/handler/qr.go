package handler

import (
	"net/http"

	"github.com/broodco/linkshort/internal/store"
	qrcode "github.com/skip2/go-qrcode"
)

type QRHandler struct {
	store   *store.Store
	baseURL string
}

func NewQRHandler(s *store.Store, baseURL string) *QRHandler {
	return &QRHandler{store: s, baseURL: baseURL}
}

func (h *QRHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	_, err := h.store.GetLinkBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	url := h.baseURL + "/" + slug

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=86400")
	_, _ = w.Write(png)
}
