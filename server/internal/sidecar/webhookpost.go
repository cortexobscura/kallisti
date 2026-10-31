package sidecar

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"mime"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/time/rate"

	"kallisti.chat/server/internal/mumble"
	"kallisti.chat/server/internal/store"
)

// Posts to a webhook (see webhooks.go): POST /api/webhooks/{id}/{token},
// Discord's path and format, so tools with a "Discord webhook" setting
// work with the URL alone. JSON, or multipart with payload_json (files are
// ignored). content (or Slack's text) and embeds become the message text;
// username names the post (else the webhook's name; never its picture: a
// picture URL would make every reader load a foreign image). Nobody is
// pinged (as poll results). Answers and errors are Discord's.

const (
	maxWebhookBody = 8 << 20 // with files, which are not read
	maxPayloadJSON = 64 << 10
	maxEmbeds      = 10 // as Discord
	// Reading a post's body at most.
	webhookReadTime = 30 * time.Second
	// Discord's limit per webhook: 5 posts per 2 seconds.
	webhookRate  = rate.Limit(2.5)
	webhookBurst = 5
)

type webhookPayload struct {
	Content  string         `json:"content"`
	Text     string         `json:"text"` // Slack's
	Username string         `json:"username"`
	Embeds   []webhookEmbed `json:"embeds"`
}

type webhookEmbed struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Author      *struct {
		Name string `json:"name"`
	} `json:"author"`
	Fields []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"fields"`
	Footer *struct {
		Text string `json:"text"`
	} `json:"footer"`
}

// maskedLink is Discord's [text](url), which our Markdown does not have:
// "text (url)" (the clients link the URL).
var maskedLink = regexp.MustCompile(`\[([^\]\n]+)\]\((https?://[^)\s]+)\)`)

func unmask(s string) string { return maskedLink.ReplaceAllString(s, "$1 ($2)") }

// text is the message: content, then each embed as lines (its author, the
// title in bold with its link, the description quoted, the fields, the
// footer), cut at MaxTextLength. No pings (see quietMentions: after the
// cut, so a full content stays whole).
func (p webhookPayload) text() string {
	var lines []string
	add := func(s string) {
		if s = strings.TrimSpace(s); s != "" {
			lines = append(lines, s)
		}
	}
	add(unmask(p.Content))
	if strings.TrimSpace(p.Content) == "" {
		add(unmask(p.Text))
	}
	for _, e := range p.Embeds[:min(len(p.Embeds), maxEmbeds)] {
		if e.Author != nil {
			add(e.Author.Name)
		}
		title := strings.TrimSpace(e.Title)
		if title != "" {
			title = "**" + title + "**"
		}
		add(title + " " + e.URL)
		for _, l := range strings.Split(strings.TrimSpace(unmask(e.Description)), "\n") {
			if strings.TrimSpace(l) != "" {
				add("> " + l)
			}
		}
		for _, f := range e.Fields {
			if name := strings.TrimSpace(f.Name); name != "" {
				add("**" + name + "**: " + unmask(f.Value))
			} else {
				add(unmask(f.Value))
			}
		}
		if e.Footer != nil {
			add(e.Footer.Text)
		}
	}
	text := strings.Join(lines, "\n")
	if utf8.RuneCountInString(text) > MaxTextLength {
		text = shorten(text, MaxTextLength-1)
	}
	return quietMentions(text)
}

// mentionStart is where a mention can start (as in mentionPattern and the
// clients): an "@" elsewhere (in a URL, an address) pings nobody.
var mentionStart = regexp.MustCompile(`(^|[\s\v\p{Z}\x{FEFF}(\[{,:])@`)

// quietMentions is quiet for the "@"s that could mention: URLs stay whole.
func quietMentions(s string) string { return mentionStart.ReplaceAllString(s, "${1}@\u200b") }

// webhookLimits: a rate limiter per webhook.
type webhookLimits struct {
	mu sync.Mutex
	by map[int64]*rate.Limiter
}

// wait is how long the webhook must wait (0: it may post now).
func (l *webhookLimits) wait(id int64, now time.Time) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.by == nil {
		l.by = map[int64]*rate.Limiter{}
	}
	lim, ok := l.by[id]
	if !ok {
		lim = rate.NewLimiter(webhookRate, webhookBurst)
		l.by[id] = lim
	}
	r := lim.ReserveN(now, 1)
	d := r.DelayFrom(now)
	if d > 0 {
		r.CancelAt(now)
	}
	return d
}

func (l *webhookLimits) forget(id int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.by, id)
}

func webhookJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// webhookError answers as Discord does: {"message","code"}.
func webhookError(w http.ResponseWriter, status, code int, message string) {
	webhookJSON(w, status, map[string]any{"message": message, "code": code})
}

var errPayload = errors.New("cannot parse the body")

// readPayload reads JSON, or multipart: payload_json, or the content and
// username fields (files are skipped).
func readPayload(r *http.Request) (webhookPayload, error) {
	var p webhookPayload
	decode := func(body io.Reader) (webhookPayload, error) {
		d := json.NewDecoder(io.LimitReader(body, maxPayloadJSON))
		if d.Decode(&p) != nil || d.Decode(&struct{}{}) != io.EOF { // one value only
			return p, errPayload
		}
		return p, nil
	}
	ctype, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch ctype {
	case "application/json":
		return decode(r.Body)
	case "multipart/form-data":
		mr, err := r.MultipartReader()
		if err != nil {
			return p, errPayload
		}
		fields := false
		for {
			part, err := mr.NextPart()
			if err == io.EOF && fields {
				return p, nil
			} else if err != nil {
				return p, errPayload
			}
			switch part.FormName() {
			case "payload_json":
				return decode(part)
			case "content", "username":
				b, err := io.ReadAll(io.LimitReader(part, maxPayloadJSON))
				if err != nil {
					return p, errPayload
				}
				if part.FormName() == "content" {
					p.Content = string(b)
				} else {
					p.Username = string(b)
				}
				fields = true
			}
		}
	}
	return p, errPayload
}

// ServeWebhook answers posts to webhooks.
func (h *Hub) ServeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		webhookError(w, http.StatusMethodNotAllowed, 0, "405: Method Not Allowed")
		return
	}
	// /api/webhooks/{id}/{token}, also with Discord's /slack (the same here).
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/webhooks/"), "/")
	if len(parts) == 3 && parts[2] == "slack" {
		parts = parts[:2]
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if len(parts) != 2 || err != nil {
		webhookError(w, http.StatusNotFound, 10015, "Unknown Webhook")
		return
	}
	hook, err := h.store.WebhookByToken(id, inviteHash(parts[1]))
	if errors.Is(err, store.ErrNotFound) {
		if _, err := h.store.WebhookChannel(id); err == nil {
			webhookError(w, http.StatusUnauthorized, 50027, "Invalid Webhook Token")
		} else {
			webhookError(w, http.StatusNotFound, 10015, "Unknown Webhook")
		}
		return
	} else if err != nil {
		slog.Error("webhook", "err", err)
		webhookError(w, http.StatusInternalServerError, 0, "Internal Server Error")
		return
	}
	if d := h.webhookLimits.wait(hook.ID, h.now()); d > 0 {
		secs := math.Ceil(d.Seconds()*1000) / 1000
		w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(d.Seconds()))))
		webhookJSON(w, http.StatusTooManyRequests, map[string]any{"message": "You are being rate limited.", "retry_after": secs, "global": false})
		return
	}
	// A slow body cannot hold a post past its webhook's deletion for long
	// (checked again before storing, too).
	http.NewResponseController(w).SetReadDeadline(time.Now().Add(webhookReadTime))
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBody)
	p, err := readPayload(r)
	if err != nil {
		webhookError(w, http.StatusBadRequest, 50109, "The request body contains invalid JSON.")
		return
	}
	name := strings.TrimSpace(p.Username)
	if name == "" {
		name = hook.Name
	} else if _, ok := webhookName(name); !ok {
		webhookError(w, http.StatusBadRequest, 50035, "Invalid Form Body: username")
		return
	}
	if utf8.RuneCountInString(p.Content) > MaxTextLength {
		webhookError(w, http.StatusBadRequest, 50035, fmt.Sprintf("Invalid Form Body: content must be %d or fewer in length.", MaxTextLength))
		return
	}
	text := p.text()
	if text == "" {
		webhookError(w, http.StatusBadRequest, 50006, "Cannot send an empty message")
		return
	}
	// A webhook's channel is a text channel (hidden ones lost theirs).
	if !h.channelExists(hook.Channel) {
		webhookError(w, http.StatusNotFound, 10003, "Unknown Channel")
		return
	}
	ch := chat{id: hook.Channel}
	if _, err := h.store.WebhookByToken(hook.ID, inviteHash(parts[1])); err != nil { // deleted meanwhile
		webhookError(w, http.StatusNotFound, 10015, "Unknown Webhook")
		return
	}
	m, err := h.store.AddMessage(store.NewMessage{ChannelID: hook.Channel, Author: name, Text: text, At: h.now(), Webhook: hook.ID})
	if err != nil {
		slog.Error("webhook post", "err", err)
		webhookError(w, http.StatusInternalServerError, 0, "Internal Server Error")
		return
	}
	h.deliver(mumble.Identity{Name: name}, ch, m, nil, "")
	if r.URL.Query().Get("wait") != "true" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	webhookJSON(w, http.StatusOK, map[string]any{
		"id": strconv.FormatInt(m.ID, 10), "channel_id": strconv.FormatInt(m.ChannelID, 10), "webhook_id": strconv.FormatInt(hook.ID, 10),
		"content": m.Text, "timestamp": m.CreatedAt.UTC().Format(time.RFC3339Nano),
		"author": map[string]any{"id": strconv.FormatInt(hook.ID, 10), "username": name, "bot": true},
	})
}
