package bridge

import (
	"encoding/json"
	"strings"
	"testing"

	"kallisti.chat/server/internal/sidecar"
	"kallisti.chat/server/internal/store"
)

// The largest send the server takes fits a text frame: the longest
// caption and descriptions, in 4-byte characters (as browsers send them,
// not escaped).
func TestTheLargestSendFitsAFrame(t *testing.T) {
	wide := func(n int) string { return strings.Repeat("😀", n) }
	type file struct {
		Key         string `json:"key"`
		Description string `json:"description"`
	}
	files := make([]file, store.MaxUploads)
	for i := range files {
		files[i] = file{Key: strings.Repeat("k", 32), Description: wide(store.MaxDescription)}
	}
	req := map[string]any{"type": "send", "channel": 1 << 40, "text": wide(sidecar.MaxTextLength), "uploads": files,
		"reply": 1 << 40, "noPing": true, "nonce": strings.Repeat("n", 64)}
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(req); err != nil {
		t.Fatal(err)
	}
	if b.Len() > maxTextFrame {
		t.Fatalf("%d bytes, the limit is %d", b.Len(), maxTextFrame)
	}
}
