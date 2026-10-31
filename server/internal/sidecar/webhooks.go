package sidecar

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"
	"unicode"
	"unicode/utf8"

	"kallisti.chat/server/internal/mumble"
	"kallisti.chat/server/internal/store"
)

// Webhooks post into a text channel from outside (see webhookpost.go):
// admins and moderators make them. The URL holds a random token (192
// bits, as invite links); the store keeps only its hash, so only the
// maker sees the URL, once.

const (
	maxWebhooks    = 10 // per channel
	maxWebhookName = 80 // characters, as Discord
)

type wireWebhook struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Creator string `json:"creator"`
	Created int64  `json:"created"` // Unix ms
}

// webhookName is a name as kept (false: empty, too long, or with control
// characters).
func webhookName(s string) (string, bool) {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return s, n > 0 && n <= maxWebhookName && !strings.ContainsFunc(s, unicode.IsControl)
}

// mayWebhook: staff, for a text channel there is; else a manage error.
func (c *Conn) mayWebhook(id mumble.Identity, channel int64) bool {
	switch {
	case c.hub.voice == nil || !c.hub.voice.MayManage(id.Session):
		c.manageError("only admins and moderators manage webhooks")
	case !c.hub.channelExists(channel):
		c.manageError("webhooks are for text channels")
	default:
		return true
	}
	return false
}

func (c *Conn) sendWebhooks(channel int64) {
	list, err := c.hub.store.Webhooks(channel)
	if err != nil {
		slog.Error("list webhooks", "err", err)
		c.manageError("cannot list the webhooks")
		return
	}
	out := make([]wireWebhook, len(list))
	for i, w := range list {
		out[i] = wireWebhook{ID: w.ID, Name: w.Name, Creator: w.Creator, Created: w.CreatedAt.UnixMilli()}
	}
	c.send(map[string]any{"type": "webhooks", "channel": channel, "webhooks": out})
}

// webhookRequest answers the webhook requests: webhooks (the list of a
// channel), create-webhook, rename-webhook, delete-webhook.
func (c *Conn) webhookRequest(id mumble.Identity, req request) {
	h := c.hub
	channel := req.Channel
	if req.Type == "rename-webhook" || req.Type == "delete-webhook" {
		var err error
		if channel, err = h.store.WebhookChannel(req.Webhook); err != nil {
			c.manageError("no such webhook")
			return
		}
	}
	if !c.mayWebhook(id, channel) {
		return
	}
	name, ok := webhookName(req.Name)
	if !ok && (req.Type == "create-webhook" || req.Type == "rename-webhook") {
		c.manageError("a webhook's name has 1 to 80 characters")
		return
	}
	switch req.Type {
	case "create-webhook":
		b := make([]byte, 24)
		rand.Read(b)
		token := base64.RawURLEncoding.EncodeToString(b)
		w, err := h.store.AddWebhook(inviteHash(token), store.Webhook{Channel: channel, Name: name, Creator: id.Name, CreatedAt: h.now()}, maxWebhooks)
		if errors.Is(err, store.ErrTooManyWebhooks) {
			c.manageError("at most 10 webhooks per channel")
			return
		} else if err != nil {
			slog.Error("add webhook", "err", err)
			c.manageError("cannot make the webhook")
			return
		}
		c.send(map[string]any{"type": "webhook-created", "channel": channel, "id": w.ID, "token": token})
	case "rename-webhook":
		if _, err := h.store.RenameWebhook(req.Webhook, name); err != nil {
			c.manageError("no such webhook")
			return
		}
	case "delete-webhook":
		if _, err := h.store.DeleteWebhook(req.Webhook); err != nil {
			c.manageError("no such webhook")
			return
		}
	}
	c.sendWebhooks(channel)
}
