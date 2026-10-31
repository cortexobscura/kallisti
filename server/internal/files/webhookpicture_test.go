package files

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kallisti.chat/server/internal/store"
)

func TestWebhookPictures(t *testing.T) {
	_, s, h := setupHandler(t, 1<<20)
	s.EnsureChannels([]string{"general"})
	w, _ := s.AddWebhook("hash", store.Webhook{Channel: 1, Name: "CI", Creator: "mod", CreatedAt: time.Now()}, 10)
	changed := int64(0)
	h.RedeemWebhookPicture = func(token string) (int64, bool) { return w.ID, token == "good" }
	h.OnWebhookPicture = func(id int64) { changed = id }
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	picture := fmt.Sprintf("/api/webhook-picture/%d.png", w.ID)
	get := func() (*http.Response, []byte) {
		resp, err := http.Get(srv.URL + picture)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		return resp, data
	}
	if resp, _ := get(); resp.StatusCode != http.StatusNotFound || resp.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("none yet: %d %v", resp.StatusCode, resp.Header)
	}
	if resp, _ := postEmoji(t, srv, "/api/webhook-picture", "bad", pngBytes(t)); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("a bad token: %d", resp.StatusCode)
	}
	if resp, _ := postEmoji(t, srv, "/api/webhook-picture", "good", []byte("<svg/>")); resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("not a picture: %d", resp.StatusCode)
	}
	if resp, _ := postEmoji(t, srv, "/api/webhook-picture", "good", pngBytes(t)); resp.StatusCode != http.StatusOK || changed != w.ID {
		t.Fatalf("upload: %d (changed %d)", resp.StatusCode, changed)
	}
	resp, data := get()
	img, err := png.Decode(bytes.NewReader(data))
	if resp.StatusCode != 200 || err != nil || img.Bounds().Dx() != WebhookPictureSize || resp.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("served: %d %v %v", resp.StatusCode, err, resp.Header)
	}
	// No date to revalidate by (a picture replaced within the second).
	req, _ := http.NewRequest("GET", srv.URL+picture, nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
	if resp, err := http.DefaultClient.Do(req); err != nil || resp.StatusCode != 200 || resp.Header.Get("Last-Modified") != "" {
		t.Errorf("revalidated by date: %v %v", resp.StatusCode, resp.Header)
	}
	for _, bad := range []string{"/api/webhook-picture/x.png", "/api/webhook-picture/99.png"} {
		if resp, err := http.Get(srv.URL + bad); err != nil || resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: %v", bad, resp.StatusCode)
		}
	}
}
