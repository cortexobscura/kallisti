package files

import (
	"bytes"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"kallisti.chat/server/internal/store"
)

// setupLogo: the token "good" uploads a logo.
func setupLogo(t *testing.T) (*httptest.Server, *store.Store, *int) {
	t.Helper()
	_, s, h := setupHandler(t, 1<<20)
	changes := 0
	h.RedeemLogo = func(token string) (string, bool) { return "admin", token == "good" }
	h.OnLogo = func() { changes++ }
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, s, &changes
}

func TestUploadAndServeTheLogo(t *testing.T) {
	srv, s, changes := setupLogo(t)
	if resp, _ := postEmoji(t, srv, "/api/logo", "bad", pngBytes(t)); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("a bad token: %d", resp.StatusCode)
	}
	if resp, _ := postEmoji(t, srv, "/api/logo", "good", []byte("<svg/>")); resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("not a picture: %d", resp.StatusCode)
	}
	resp, v := postEmoji(t, srv, "/api/logo", "good", pngBytes(t))
	version, _ := s.LogoVersion()
	if resp.StatusCode != http.StatusOK || v["version"] != version || version == "" || *changes != 1 {
		t.Fatalf("status %d %v (version %q, changes %d)", resp.StatusCode, v, version, *changes)
	}
	get := func(path string) *http.Response {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { resp.Body.Close() })
		return resp
	}
	for _, size := range []string{"32", "180", "192", "512"} {
		want, _ := strconv.Atoi(size)
		resp := get("/api/logo/" + size + ".png?v=" + version)
		data, _ := io.ReadAll(resp.Body)
		img, err := png.Decode(bytes.NewReader(data))
		if resp.StatusCode != 200 || err != nil || img.Bounds().Dx() != want {
			t.Fatalf("%s: %d %v", size, resp.StatusCode, err)
		}
		if cc := resp.Header.Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
			t.Fatalf("%s: cache %q", size, cc)
		}
	}
	// An old or no version gets the logo now, kept only until it changes
	// (else a browser would keep the new one under the old address).
	for _, path := range []string{"/api/logo/192.png", "/api/logo/192.png?v=old"} {
		if resp := get(path); resp.StatusCode != 200 || resp.Header.Get("Cache-Control") != "no-cache" {
			t.Fatalf("%s: %d, cache %q", path, resp.StatusCode, resp.Header.Get("Cache-Control"))
		}
	}
	if resp := get("/api/logo/64.png"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("a size not made: %d", resp.StatusCode)
	}
	s.DeleteLogo()
	if resp := get("/api/logo/32.png"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("no logo: %d", resp.StatusCode)
	}
}
