package sidecar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kallisti.chat/server/internal/store"
)

func TestWebhookText(t *testing.T) {
	for _, c := range []struct {
		name, body, want string
	}{
		{"content", `{"content":"build **green**"}`, "build **green**"},
		{"slack", `{"text":"deployed"}`, "deployed"},
		{"no pings", `{"content":"@everyone @carol (@bob)"}`, "@\u200beveryone @\u200bcarol (@\u200bbob)"},
		{"URLs stay whole", `{"content":"https://x.org/@scope/pkg@1.0 https://x.org/s?q=foo,@alice @bob"}`, "https://x.org/@scope/pkg@1.0 https://x.org/s?q=foo,@alice @\u200bbob"},
		{"embed", `{"content":"Push","embeds":[{"author":{"name":"lain"},"title":"[kallisti] 2 new commits","url":"https://git.example/c",
			"description":"[abc123](https://git.example/abc) fix: grids\n[def456](https://git.example/def) docs","fields":[{"name":"Branch","value":"main"}],"footer":{"text":"Gitea"}}]}`,
			"Push\nlain\n**[kallisti] 2 new commits** https://git.example/c\n> abc123 (https://git.example/abc) fix: grids\n> def456 (https://git.example/def) docs\n**Branch**: main\nGitea"},
		{"embed with a link only", `{"embeds":[{"url":"https://x.example"}]}`, "https://x.example"},
		{"masked link in content", `{"content":"see [the log](https://ci.example/1)"}`, "see the log (https://ci.example/1)"},
		{"field without a name", `{"embeds":[{"fields":[{"name":" ","value":"v"}]}]}`, "v"},
		{"at most 10 embeds", `{"embeds":[` + strings.Repeat(`{"title":"t"},`, 11) + `{"title":"t"}]}`, strings.TrimSuffix(strings.Repeat("**t**\n", 10), "\n")},
		{"empty", `{"content":"  ","embeds":[{}]}`, ""},
	} {
		var p webhookPayload
		if err := json.Unmarshal([]byte(c.body), &p); err != nil {
			t.Fatal(err)
		}
		if got := p.text(); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

// postWebhook posts to a hub's webhook endpoint.
func postWebhook(h *Hub, path, ctype, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", path, strings.NewReader(body))
	r.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	h.ServeWebhook(w, r)
	return w
}

func TestWebhookPosts(t *testing.T) {
	h, staff, carol := webhookHub(t)
	ci := h.channels[1].ID
	id, token := createWebhook(t, staff, ci, "Gitea")
	url := fmt.Sprintf("/api/webhooks/%d/%s", id, token)

	res := postWebhook(h, url, "application/json", `{"content":"build green","username":"CI bot"}`)
	if res.Code != http.StatusNoContent {
		t.Fatalf("post: %d %s", res.Code, res.Body)
	}
	var msgID int64
	for _, c := range []*Conn{staff, carol} {
		m := next(t, c)
		msg := m["message"].(map[string]any)
		if m["type"] != "message" || msg["author"] != "CI bot" || msg["text"] != "build green" || msg["webhook"] != float64(id) || msg["verified"] != false {
			t.Fatalf("got %v", m)
		}
		msgID = int64(msg["id"].(float64))
	}
	// Mumble users see the name, marked (the text comes "from the server");
	// their history shows it once.
	stored := mustMessage(t, h, msgID)
	if got := h.mumbleLive(stored); got != "CI bot (webhook): build green" {
		t.Errorf("mumble: %q", got)
	}
	if got := h.historyLines([]store.Message{stored}); !strings.Contains(got, "<b>CI bot (webhook)</b>") || strings.Count(got, "CI bot") != 1 {
		t.Errorf("history: %q", got)
	}
	reply := store.Message{Author: "carol", Text: "thanks", Reply: &store.Quote{Author: "CI bot", Text: "build green", Webhook: true}}
	if got := h.mumbleText(reply); !strings.HasPrefix(got, "> CI bot (webhook): ") {
		t.Errorf("reply: %q", got)
	}

	// ?wait=true: the message; no username: the webhook's name; multipart
	// with payload_json (files ignored).
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	mw.WriteField("payload_json", `{"content":"with a file"}`)
	f, _ := mw.CreateFormFile("files[0]", "log.txt")
	f.Write([]byte("lots of log"))
	mw.Close()
	res = postWebhook(h, url+"?wait=true", mw.FormDataContentType(), b.String())
	var got map[string]any
	json.Unmarshal(res.Body.Bytes(), &got)
	if res.Code != http.StatusOK || got["content"] != "with a file" || got["author"].(map[string]any)["username"] != "Gitea" || got["channel_id"] != fmt.Sprint(ci) {
		t.Fatalf("wait: %d %v", res.Code, got)
	}
	next(t, carol)
	next(t, staff)

	// Discord takes form fields too; a full 2000 characters stay whole.
	b.Reset()
	mw = multipart.NewWriter(&b)
	full := strings.Repeat("a", MaxTextLength-2) + "@b"
	mw.WriteField("content", full)
	mw.Close()
	if res := postWebhook(h, url, mw.FormDataContentType(), b.String()); res.Code != 204 {
		t.Fatalf("form: %d %s", res.Code, res.Body)
	}
	if m := next(t, carol)["message"].(map[string]any); m["text"] != full {
		t.Fatalf("form text: %.40q…", m["text"])
	}
	next(t, staff)

	// Errors, as Discord's.
	for _, c := range []struct {
		path, ctype, body string
		code, discord     int
	}{
		{fmt.Sprintf("/api/webhooks/%d/wrong", id), "application/json", `{"content":"x"}`, 401, 50027},
		{"/api/webhooks/999/" + token, "application/json", `{"content":"x"}`, 404, 10015},
		{"/api/webhooks/x", "application/json", `{"content":"x"}`, 404, 10015},
		{url, "application/json", `{"content":""}`, 400, 50006},
		{url, "application/json", `{"content":` + fmt.Sprintf("%q", strings.Repeat("a", MaxTextLength+1)) + `}`, 400, 50035},
		{url, "application/json", `{"content":"x","username":"` + strings.Repeat("a", maxWebhookName+1) + `"}`, 400, 50035},
		{url, "application/json", `not json`, 400, 50109},
		{url, "application/json", `{"content":"x"}not json`, 400, 50109},
		{url, "application/json", `{"content":"x"}` + strings.Repeat(" ", maxPayloadJSON) + `{}`, 400, 50109},
		{url, "text/plain", `hello`, 400, 50109},
	} {
		h.webhookLimits.forget(id) // a fixed clock: each case on its own
		res := postWebhook(h, c.path, c.ctype, c.body)
		var e struct {
			Code int `json:"code"`
		}
		json.Unmarshal(res.Body.Bytes(), &e)
		if res.Code != c.code || e.Code != c.discord {
			t.Errorf("%s %s: %d %s", c.path, c.body[:min(len(c.body), 30)], res.Code, res.Body)
		}
	}
	empty(t, carol)

	// The webhook's channel gone: its URL too.
	staff.Handle([]byte(fmt.Sprintf(`{"type":"delete-webhook","webhook":%d}`, id)))
	webhooksOf(t, staff, ci)
	if res := postWebhook(h, url, "application/json", `{"content":"x"}`); res.Code != 404 {
		t.Errorf("deleted: %d", res.Code)
	}
}

// Discord's limit: 5 posts per 2 seconds per webhook; then 429 with
// Retry-After.
func TestWebhookRateLimit(t *testing.T) {
	h, staff, _ := webhookHub(t)
	ci := h.channels[1].ID
	id, token := createWebhook(t, staff, ci, "Spam")
	url := fmt.Sprintf("/api/webhooks/%d/%s", id, token)
	for i := range 5 {
		if res := postWebhook(h, url, "application/json", `{"content":"x"}`); res.Code != 204 {
			t.Fatalf("post %d: %d", i, res.Code)
		}
	}
	// Before the body is read: also a bad one counts as over the limit.
	res := postWebhook(h, url, "application/json", `not json`)
	if res.Code != 429 || res.Header().Get("Retry-After") == "" {
		t.Fatalf("over the limit: %d %v", res.Code, res.Header())
	}
	var e struct {
		RetryAfter float64 `json:"retry_after"`
	}
	json.Unmarshal(res.Body.Bytes(), &e)
	if e.RetryAfter <= 0 || e.RetryAfter > 2 {
		t.Errorf("retry_after %v", e.RetryAfter)
	}
}
