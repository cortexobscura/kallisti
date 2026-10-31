package sidecar

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/time/rate"

	"kallisti.chat/server/internal/mumble"
)

// webhookHub: a hub with the voice server; staff (session 1) and carol.
func webhookHub(t *testing.T) (h *Hub, staff, carol *Conn) {
	t.Helper()
	h = newHub(t, "general", "ci")
	h.SetVoice(&fakeVoice{hidden: map[uint32]bool{}, members: map[uint32]bool{}}, "")
	staff = joinAs(t, h, mumble.Identity{Session: 1, Name: "mod", Registered: true, UserID: 10})
	carol = joinAs(t, h, mumble.Identity{Session: 3, Name: "carol", Registered: true, UserID: 30})
	staff.limit = rate.NewLimiter(rate.Inf, 0)
	carol.limit = rate.NewLimiter(rate.Inf, 0)
	return h, staff, carol
}

// createWebhook makes one in the channel and returns its id and token.
func createWebhook(t *testing.T, staff *Conn, channel int64, name string) (int64, string) {
	t.Helper()
	staff.Handle([]byte(fmt.Sprintf(`{"type":"create-webhook","channel":%d,"name":%q}`, channel, name)))
	m := next(t, staff)
	token, _ := m["token"].(string)
	if m["type"] != "webhook-created" || m["channel"] != float64(channel) || len(token) < 32 {
		t.Fatalf("got %v", m)
	}
	webhooksOf(t, staff, channel)
	return int64(m["id"].(float64)), token
}

func webhooksOf(t *testing.T, c *Conn, channel int64) []any {
	t.Helper()
	m := next(t, c)
	if m["type"] != "webhooks" || m["channel"] != float64(channel) {
		t.Fatalf("got %v", m)
	}
	return m["webhooks"].([]any)
}

func TestManageWebhooks(t *testing.T) {
	h, staff, carol := webhookHub(t)
	ci := h.channels[1].ID
	id, _ := createWebhook(t, staff, ci, " Gitea ")
	staff.Handle([]byte(fmt.Sprintf(`{"type":"webhooks","channel":%d}`, ci)))
	list := webhooksOf(t, staff, ci)
	w := list[0].(map[string]any)
	if len(list) != 1 || w["id"] != float64(id) || w["name"] != "Gitea" || w["creator"] != "mod" || w["created"] == nil || w["token"] != nil {
		t.Fatalf("list: %v", list)
	}
	staff.Handle([]byte(fmt.Sprintf(`{"type":"rename-webhook","webhook":%d,"name":"Forgejo"}`, id)))
	if w := webhooksOf(t, staff, ci)[0].(map[string]any); w["name"] != "Forgejo" {
		t.Fatalf("renamed: %v", w)
	}
	// Others may not; bad names, channels and ids are refused.
	for _, req := range []string{
		fmt.Sprintf(`{"type":"webhooks","channel":%d}`, ci),
		fmt.Sprintf(`{"type":"create-webhook","channel":%d,"name":"x"}`, ci),
		fmt.Sprintf(`{"type":"delete-webhook","webhook":%d}`, id),
	} {
		carol.Handle([]byte(req))
		if m := next(t, carol); m["type"] != "error" {
			t.Fatalf("carol %s: %v", req, m)
		}
	}
	for _, req := range []string{
		fmt.Sprintf(`{"type":"create-webhook","channel":%d,"name":"  "}`, ci),
		fmt.Sprintf(`{"type":"create-webhook","channel":%d,"name":%q}`, ci, strings.Repeat("a", maxWebhookName+1)),
		`{"type":"create-webhook","channel":999,"name":"x"}`,
		`{"type":"rename-webhook","webhook":999,"name":"x"}`,
		fmt.Sprintf(`{"type":"rename-webhook","webhook":%d,"name":""}`, id),
		`{"type":"delete-webhook","webhook":999}`,
	} {
		staff.Handle([]byte(req))
		if m := next(t, staff); m["type"] != "error" {
			t.Fatalf("%s: %v", req, m)
		}
	}
	// At most maxWebhooks per channel.
	for i := 1; i < maxWebhooks; i++ {
		createWebhook(t, staff, ci, fmt.Sprint("hook", i))
	}
	staff.Handle([]byte(fmt.Sprintf(`{"type":"create-webhook","channel":%d,"name":"one more"}`, ci)))
	if m := next(t, staff); m["type"] != "error" {
		t.Fatalf("over the limit: %v", m)
	}
	staff.Handle([]byte(fmt.Sprintf(`{"type":"delete-webhook","webhook":%d}`, id)))
	if list := webhooksOf(t, staff, ci); len(list) != maxWebhooks-1 {
		t.Fatalf("after delete: %d", len(list))
	}
}
