package files

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"kallisti.chat/server/internal/sidecar"
	"kallisti.chat/server/internal/store"
)

// uploadSticker stores a sticker (POST /api/stickers): the token (from
// the sidecar, for those who may manage emoji) gives its name and fields;
// the body is the image.
func (h *Handler) uploadSticker(w http.ResponseWriter, r *http.Request) {
	st, ok := h.RedeemSticker(r.Header.Get("X-Upload-Token"))
	if !ok {
		jsonError(w, "invalid or expired upload token", http.StatusUnauthorized)
		return
	}
	if st.Type, st.Data, ok = readInlineImage(w, r, MaxStickerBytes, "a sticker"); !ok {
		return
	}
	st, err := h.Store.AddSticker(st, h.Now())
	switch {
	case errors.Is(err, store.ErrNameTaken):
		jsonError(w, "a sticker with this name exists", http.StatusConflict)
		return
	case errors.Is(err, store.ErrTooManyStickers):
		jsonError(w, "this server has the most stickers it can have", http.StatusConflict)
		return
	case err != nil:
		jsonError(w, "cannot store the sticker", http.StatusInternalServerError)
		return
	}
	if h.OnSticker != nil {
		h.OnSticker()
	}
	writeJSON(w, map[string]any{"id": st.ID, "name": st.Name, "url": sidecar.StickerURL(st.ID)})
}

// serveSticker serves a sticker's image (GET /api/stickers/{id}); none
// once deleted.
func (h *Handler) serveSticker(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/stickers/"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	st, err := h.Store.StickerByID(id)
	if err != nil || st.Deleted {
		http.NotFound(w, r)
		return
	}
	// Private like personal emoji: a deleted sticker should go from caches.
	serveInline(w, r, st.Type, st.Data, st.CreatedAt, false)
}
