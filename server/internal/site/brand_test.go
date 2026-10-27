package site

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const indexHTML = `<head>
    <link rel="manifest" href="/manifest.json" />
    <link rel="icon" href="/icons/favicon-32.png" sizes="32x32" />
    <link rel="apple-touch-icon" href="/icons/apple-touch-icon.png" />
    <title>Kallisti</title>
</head>`

const manifestJSON = `{"name": "Kallisti", "short_name": "Kallisti", "display": "standalone",
  "icons": [{ "src": "/icons/icon-192.png", "sizes": "192x192", "type": "image/png" }]}`

func brandedDir(t *testing.T) string {
	dir := webDir(t)
	os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0o644)
	os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifestJSON), 0o644)
	return dir
}

func TestTheServersNameAndLogoInThePage(t *testing.T) {
	brand := Brand{Name: "Lain's <place>", Logo: "abc123"}
	h := New(bridge, brandedDir(t), func() Brand { return brand })
	for _, path := range []string{"/", "/channels/3"} {
		w := get(t, h, path, false)
		body := w.Body.String()
		for _, want := range []string{
			"<title>Lain&#39;s &lt;place&gt;</title>",
			`<link rel="icon" href="/api/logo/32.png?v=abc123" sizes="32x32" />`,
			`<link rel="apple-touch-icon" href="/api/logo/180.png?v=abc123" />`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("%s: no %q in %s", path, want, body)
			}
		}
		if w.Header().Get("Cache-Control") != "no-cache" || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
			t.Fatalf("%s: headers %v", path, w.Header())
		}
	}

	w := get(t, h, "/manifest.json", false)
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	icons, _ := json.Marshal(m["icons"])
	if m["name"] != brand.Name || m["short_name"] != brand.Name || m["display"] != "standalone" ||
		string(icons) != `[{"sizes":"192x192","src":"/api/logo/192.png?v=abc123","type":"image/png"},{"sizes":"512x512","src":"/api/logo/512.png?v=abc123","type":"image/png"}]` {
		t.Fatalf("manifest %s", w.Body)
	}
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("manifest headers %v", w.Header())
	}

	// No logo: Kallisti's icons stay.
	brand = Brand{Name: "Our place"}
	if body := get(t, h, "/", false).Body.String(); !strings.Contains(body, "/icons/favicon-32.png") || !strings.Contains(body, "<title>Our place</title>") {
		t.Fatalf("no logo: %s", body)
	}
	if body := get(t, h, "/manifest.json", false).Body.String(); !strings.Contains(body, "/icons/icon-192.png") {
		t.Fatalf("no logo: manifest %s", body)
	}
}

// The web app's own index.html and manifest: branding needs their exact
// title and icon links.
func TestTheWebAppsPageCanBeBranded(t *testing.T) {
	web := filepath.Join("..", "..", "..", "apps", "web")
	page, err := os.ReadFile(filepath.Join(web, "index.html"))
	if err != nil {
		t.Skip("no web app here:", err)
	}
	b := Brand{Name: "Our place", Logo: "abc"}
	got := brandIndex(string(page), b)
	for _, want := range []string{"<title>Our place</title>", `href="/api/logo/32.png?v=abc"`, `href="/api/logo/180.png?v=abc"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("index.html: no %q after branding", want)
		}
	}
	manifest, err := os.ReadFile(filepath.Join(web, "public", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if out, err := brandManifest(manifest, b); err != nil || !strings.Contains(string(out), `"name":"Our place"`) {
		t.Fatalf("manifest: %s %v", out, err)
	}
}
