package files

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kallisti.chat/server/internal/store"
)

// setupStickers: the token "good" adds the sticker "wave" (by alice).
func setupStickers(t *testing.T) (*httptest.Server, *store.Store, *int) {
	t.Helper()
	_, s, h := setupHandler(t, 1<<20)
	changes := 0
	h.RedeemSticker = func(token string) (store.Sticker, bool) {
		return store.Sticker{Name: "wave", Description: "Hi", Emoji: "👋", Pack: "hello", Author: "alice"}, token == "good"
	}
	h.OnSticker = func() { changes++ }
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, s, &changes
}

func TestUploadAndServeASticker(t *testing.T) {
	srv, s, changes := setupStickers(t)
	img := pngBytes(t)
	if resp, _ := postEmoji(t, srv, "/api/stickers", "bad", img); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad token: %d", resp.StatusCode)
	}
	resp, v := postEmoji(t, srv, "/api/stickers", "good", img)
	if resp.StatusCode != http.StatusOK || v["name"] != "wave" || *changes != 1 {
		t.Fatalf("status %d %v (changes %d)", resp.StatusCode, v, *changes)
	}
	list, _ := s.Stickers()
	if len(list) != 1 || list[0].Pack != "hello" || list[0].Author != "alice" || list[0].Type != "image/png" {
		t.Fatalf("stored %+v", list)
	}
	url := v["url"].(string)
	get, err := http.Get(srv.URL + url)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(get.Body)
	get.Body.Close()
	if !bytes.Equal(body, img) || get.Header.Get("Content-Type") != "image/png" || !strings.Contains(get.Header.Get("Cache-Control"), "immutable") || get.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("served %d bytes, %v", len(body), get.Header)
	}
	s.DeleteSticker(list[0].ID)
	if get, _ := http.Get(srv.URL + url); get.StatusCode != http.StatusNotFound {
		t.Fatalf("deleted: %d", get.StatusCode)
	}
}

// Stickers may be larger than emoji (512 KB), with the same type checks.
func TestStickerUploadsAreChecked(t *testing.T) {
	srv, _, _ := setupStickers(t)
	big := append(pngBytes(t), make([]byte, 300<<10)...) // over the emoji limit, under the sticker one
	if resp, v := postEmoji(t, srv, "/api/stickers", "good", big); resp.StatusCode != http.StatusOK {
		t.Fatalf("300 KB: %d %v", resp.StatusCode, v)
	}
	if resp, _ := postEmoji(t, srv, "/api/stickers", "good", make([]byte, MaxStickerBytes+1)); resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("too large: %d", resp.StatusCode)
	}
	if resp, _ := postEmoji(t, srv, "/api/stickers", "good", []byte("<svg></svg>")); resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("svg: %d", resp.StatusCode)
	}
}
