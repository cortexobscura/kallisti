package files

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kallisti.chat/server/internal/avatarimg"
	"kallisti.chat/server/internal/store"
)

// WebhookPictureSize is a webhook's picture (pixels a side; shown at 40).
const WebhookPictureSize = 128

// uploadWebhookPicture sets a webhook's picture (POST
// /api/webhook-picture; the token, from the sidecar, names the webhook):
// its middle square.
func (h *Handler) uploadWebhookPicture(w http.ResponseWriter, r *http.Request) {
	id, ok := h.RedeemWebhookPicture(r.Header.Get("X-Upload-Token"))
	if !ok {
		jsonError(w, "invalid or expired upload token", http.StatusUnauthorized)
		return
	}
	_, data, ok := readInlineImage(w, r, MaxLogoBytes, "a picture")
	if !ok {
		return
	}
	icons, err := avatarimg.Icons(data, []int{WebhookPictureSize})
	if errors.Is(err, avatarimg.ErrTooLarge) {
		jsonError(w, "the image has too many pixels", http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		jsonError(w, "the image is broken", http.StatusUnsupportedMediaType)
		return
	}
	if _, err := h.Store.SetWebhookPicture(id, icons[WebhookPictureSize], h.Now()); errors.Is(err, store.ErrNotFound) {
		jsonError(w, "no such webhook", http.StatusNotFound)
		return
	} else if err != nil {
		slog.Error("store a webhook picture", "err", err)
		jsonError(w, "cannot store the picture", http.StatusInternalServerError)
		return
	}
	if h.OnWebhookPicture != nil {
		h.OnWebhookPicture(id)
	}
	writeJSON(w, map[string]string{})
}

// serveWebhookPicture serves one (GET /api/webhook-picture/<id>.png; the
// clients add ?v=<version>, see the sidecar's WebhookPictureChanged),
// checked again each time.
func (h *Handler) serveWebhookPicture(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/webhook-picture/"), ".png"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data, _, err := h.Store.WebhookPicture(id)
	if errors.Is(err, store.ErrNotFound) {
		w.Header().Set("Cache-Control", "no-cache") // it may come
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "cannot read the picture", http.StatusInternalServerError)
		return
	}
	// No date: one replaced within the second would revalidate as the same.
	serveImage(w, r, "image/png", data, time.Time{}, "no-cache")
}
