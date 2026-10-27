package files

import (
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"kallisti.chat/server/internal/avatarimg"
	"kallisti.chat/server/internal/store"
)

// LogoSizes are the logo's icons (pixels a side): the tab icon, Apple's
// home screen icon, and the installed app's.
var LogoSizes = []int{32, 180, 192, 512}

// MaxLogoBytes is the largest logo upload.
const MaxLogoBytes = 2 << 20

// uploadLogo replaces the server's logo (POST /api/logo; the token, from
// the sidecar, is for admins): its middle square in each of LogoSizes.
func (h *Handler) uploadLogo(w http.ResponseWriter, r *http.Request) {
	author, ok := h.RedeemLogo(r.Header.Get("X-Upload-Token"))
	if !ok {
		jsonError(w, "invalid or expired upload token", http.StatusUnauthorized)
		return
	}
	_, data, ok := readInlineImage(w, r, MaxLogoBytes, "a logo")
	if !ok {
		return
	}
	icons, err := avatarimg.Icons(data, LogoSizes)
	if errors.Is(err, avatarimg.ErrTooLarge) {
		jsonError(w, "the image has too many pixels", http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		jsonError(w, "the image is broken", http.StatusUnsupportedMediaType)
		return
	}
	version, err := h.Store.SetLogo(icons)
	if err != nil {
		slog.Error("store the logo", "err", err)
		jsonError(w, "cannot store the logo", http.StatusInternalServerError)
		return
	}
	slog.Info("logo changed", "by", author)
	if h.OnLogo != nil {
		h.OnLogo()
	}
	writeJSON(w, map[string]string{"version": version})
}

// serveLogo serves an icon (GET /api/logo/<size>.png?v=<version>): kept
// for ever at the current version's address, else only until it changes.
func (h *Handler) serveLogo(w http.ResponseWriter, r *http.Request) {
	size, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/logo/"), ".png"))
	if err != nil || !slices.Contains(LogoSizes, size) {
		http.NotFound(w, r)
		return
	}
	data, version, err := h.Store.Logo(size)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "cannot read the logo", http.StatusInternalServerError)
		return
	}
	cache := "no-cache"
	if r.URL.Query().Get("v") == version {
		cache = "public, max-age=31536000, immutable"
	}
	serveImage(w, r, "image/png", data, time.Time{}, cache)
}
