package store

import (
	"errors"
	"testing"
	"time"
)

func TestWebhooks(t *testing.T) {
	s := open(t)
	s.EnsureChannels([]string{"general", "ci"})
	at := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	w, err := s.AddWebhook("hash1", Webhook{Channel: 2, Name: "Gitea", Creator: "alice", CreatedAt: at}, 2)
	if err != nil || w.ID == 0 {
		t.Fatalf("add: %+v %v", w, err)
	}
	if _, err := s.AddWebhook("hash2", Webhook{Channel: 2, Name: "CI", Creator: "alice", CreatedAt: at}, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddWebhook("hash3", Webhook{Channel: 2, Name: "More", Creator: "alice", CreatedAt: at}, 2); !errors.Is(err, ErrTooManyWebhooks) {
		t.Fatalf("over the limit: %v", err)
	}
	list, err := s.Webhooks(2)
	if err != nil || len(list) != 2 || list[0] != w {
		t.Fatalf("list: %+v %v", list, err)
	}
	if got, err := s.WebhookByToken(w.ID, "hash1"); err != nil || got != w {
		t.Fatalf("by token: %+v %v", got, err)
	}
	for _, c := range []struct {
		id   int64
		hash string
	}{{w.ID, "hash2"}, {99, "hash1"}} {
		if _, err := s.WebhookByToken(c.id, c.hash); !errors.Is(err, ErrNotFound) {
			t.Errorf("%d %s: %v", c.id, c.hash, err)
		}
	}
	if ch, err := s.RenameWebhook(w.ID, "Forgejo"); err != nil || ch != 2 {
		t.Fatalf("rename: %d %v", ch, err)
	}
	if got, _ := s.WebhookByToken(w.ID, "hash1"); got.Name != "Forgejo" {
		t.Errorf("renamed: %+v", got)
	}
	if ch, err := s.DeleteWebhook(w.ID); err != nil || ch != 2 {
		t.Fatalf("delete: %d %v", ch, err)
	}
	if _, err := s.DeleteWebhook(w.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted twice: %v", err)
	}
	// Ids are not used again (a stale delete must not hit a new one).
	last, _ := s.Webhooks(2)
	s.DeleteWebhook(last[0].ID)
	if again, _ := s.AddWebhook("hash5", Webhook{Channel: 2, Name: "New", Creator: "alice", CreatedAt: at}, 2); again.ID <= last[0].ID {
		t.Errorf("id %d used again", again.ID)
	}
	// A hidden (deleted) channel's webhooks go with it.
	if err := s.HideChannel(2); err != nil {
		t.Fatal(err)
	}
	if list, _ := s.Webhooks(2); len(list) != 0 {
		t.Errorf("after the channel: %+v", list)
	}
	// Also a configured channel no longer configured.
	s.AddWebhook("hash4", Webhook{Channel: 1, Name: "x", Creator: "alice", CreatedAt: at}, 2)
	if err := s.EnsureChannels([]string{"other"}); err != nil {
		t.Fatal(err)
	}
	if list, _ := s.Webhooks(1); len(list) != 0 {
		t.Errorf("after the configuration: %+v", list)
	}
}

func TestWebhookMessages(t *testing.T) {
	s := open(t)
	s.EnsureChannels([]string{"general"})
	w, err := s.AddWebhook("hash", Webhook{Channel: 1, Name: "CI", Creator: "alice", CreatedAt: time.Now()}, 10)
	if err != nil {
		t.Fatal(err)
	}
	m, err := s.AddMessage(NewMessage{ChannelID: 1, Author: "Build bot", Text: "green", At: time.Now(), Webhook: w.ID})
	if err != nil || m.Webhook != w.ID {
		t.Fatalf("got %+v %v", m, err)
	}
	if hist, _ := s.History(1, 0, 10); hist[0].Webhook != w.ID {
		t.Errorf("history: %+v", hist[0])
	}
	// A reply quotes it as a webhook's (its name may be a user's).
	r, err := s.AddMessage(NewMessage{ChannelID: 1, Author: "bob", Text: "thanks", At: time.Now(), ReplyTo: m.ID})
	if err != nil || r.Reply == nil || !r.Reply.Webhook {
		t.Fatalf("reply: %+v %v", r.Reply, err)
	}
	if got, _ := s.MessageByID(r.ID); got.Reply == nil || !got.Reply.Webhook {
		t.Errorf("stored reply: %+v", got.Reply)
	}
}
