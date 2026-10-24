package sidecar

import (
	"fmt"
	"testing"

	"kallisti.chat/server/internal/mumble"
	"kallisti.chat/server/internal/store"
)

func addSticker(t *testing.T, h *Hub, name, pack string) store.Sticker {
	t.Helper()
	s, err := h.store.AddSticker(store.Sticker{Name: name, Pack: pack, Description: "a " + name, Emoji: "👋", Type: "image/png", Data: []byte("\x89PNG"), Author: "admin"}, clock)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestReadyAndChangesSendTheStickers(t *testing.T) {
	h := newHub(t, "general")
	s := addSticker(t, h, "wave", "hello")
	h.StickersChanged()
	c := h.Connect()
	defer c.Close()
	c.SetIdentity(mumble.Identity{Session: 4, Name: "alice"})
	m := next(t, c)
	list, _ := m["stickers"].([]any)
	if len(list) != 1 {
		t.Fatalf("ready %v", m["stickers"])
	}
	got := list[0].(map[string]any)
	if got["id"] != float64(s.ID) || got["name"] != "wave" || got["pack"] != "hello" || got["description"] != "a wave" || got["emoji"] != "👋" || got["url"] != fmt.Sprintf("/api/stickers/%d", s.ID) {
		t.Fatalf("sticker %v", got)
	}
	addSticker(t, h, "cat", "")
	h.StickersChanged()
	if m := next(t, c); m["type"] != "stickers" || len(m["stickers"].([]any)) != 2 {
		t.Fatalf("change %v", m)
	}
}

func TestReadyHasAnEmptyStickerList(t *testing.T) {
	h := newHub(t, "general")
	c := h.Connect()
	defer c.Close()
	c.SetIdentity(mumble.Identity{Session: 4, Name: "alice"})
	if m := next(t, c); m["stickers"] == nil || len(m["stickers"].([]any)) != 0 {
		t.Fatalf("got %v", m["stickers"])
	}
}

func TestStickerTokensAreForManagers(t *testing.T) {
	h := newHub(t, "general")
	h.SetVoice(&fakeVoice{}, "")
	admin := readyAs(t, h, 1, "admin")
	bob := readyAs(t, h, 2, "bob")
	addSticker(t, h, "taken", "")
	h.StickersChanged()
	next(t, admin)
	next(t, bob)
	bob.Handle([]byte(`{"type":"sticker-token","name":"wave"}`))
	if m := next(t, bob); m["type"] != "error" || m["about"] != "manage" {
		t.Fatalf("bob got %v", m)
	}
	long := fmt.Sprintf("%0101d", 0)
	for _, bad := range []string{`"name":"Wave"`, `"name":"taken"`, `"name":"wave","emoji":"abc"`, `"name":"wave","description":"` + long + `"`, `"name":"wave","pack":"` + long + `"`} {
		admin.Handle([]byte(`{"type":"sticker-token",` + bad + `}`))
		if m := next(t, admin); m["type"] != "error" || m["about"] != "manage" {
			t.Fatalf("%s: %v", bad, m)
		}
	}
	admin.Handle([]byte(`{"type":"sticker-token","name":"wave","description":"Hello","emoji":"👋","pack":"greetings"}`))
	m := next(t, admin)
	token, _ := m["token"].(string)
	if m["type"] != "upload-token" || token == "" {
		t.Fatalf("got %v", m)
	}
	// For one sticker upload, not for emoji or files.
	st, ok := h.RedeemStickerToken(token)
	if !ok || st.Author != "admin" || st.Name != "wave" || st.Description != "Hello" || st.Emoji != "👋" || st.Pack != "greetings" {
		t.Fatalf("redeem: %+v %v", st, ok)
	}
	if _, ok := h.RedeemStickerToken(token); ok {
		t.Fatal("redeemed twice")
	}
	admin.Handle([]byte(`{"type":"sticker-token","name":"wave"}`))
	token = next(t, admin)["token"].(string)
	if _, _, ok := h.RedeemEmojiToken(token); ok {
		t.Fatal("a sticker token added an emoji")
	}
	admin.Handle([]byte(`{"type":"sticker-token","name":"wave"}`))
	if _, ok := h.RedeemUploadToken(next(t, admin)["token"].(string)); ok {
		t.Fatal("a sticker token uploaded a file")
	}
	admin.Handle([]byte(`{"type":"emoji-token","name":"party"}`))
	if _, ok := h.RedeemStickerToken(next(t, admin)["token"].(string)); ok {
		t.Fatal("an emoji token added a sticker")
	}
}

func TestStickerTokensAreRefusedAtTheLimit(t *testing.T) {
	h := newHub(t, "general")
	h.SetVoice(&fakeVoice{}, "")
	admin := readyAs(t, h, 1, "admin")
	for i := range store.MaxStickers {
		addSticker(t, h, fmt.Sprintf("s%d", i), "")
	}
	h.StickersChanged()
	next(t, admin)
	admin.Handle([]byte(`{"type":"sticker-token","name":"more"}`))
	if m := next(t, admin); m["type"] != "error" {
		t.Fatalf("got %v", m)
	}
}

func TestManagersRenameAndDeleteStickers(t *testing.T) {
	h := newHub(t, "general")
	h.SetVoice(&fakeVoice{}, "")
	s := addSticker(t, h, "wave", "")
	addSticker(t, h, "cat", "")
	h.StickersChanged()
	admin := readyAs(t, h, 1, "admin")
	bob := readyAs(t, h, 2, "bob")
	req := func(c *Conn, body string) { c.Handle([]byte(fmt.Sprintf(body, s.ID))) }
	req(bob, `{"type":"delete-sticker","sticker":%d}`)
	if m := next(t, bob); m["type"] != "error" || m["about"] != "manage" {
		t.Fatalf("bob got %v", m)
	}
	req(admin, `{"type":"rename-sticker","sticker":%d,"name":"cat"}`)
	if m := next(t, admin); m["type"] != "error" {
		t.Fatalf("rename to a taken name: %v", m)
	}
	req(admin, `{"type":"rename-sticker","sticker":%d,"name":"hi","description":"Hi","emoji":"🙂","pack":"p"}`)
	for _, c := range []*Conn{admin, bob} {
		m := next(t, c)
		if m["type"] != "stickers" || fmt.Sprint(m["stickers"]) == "" {
			t.Fatalf("after rename: %v", m)
		}
	}
	if got, _ := h.store.StickerByID(s.ID); got.Name != "hi" || got.Pack != "p" || got.Emoji != "🙂" {
		t.Fatalf("renamed: %+v", got)
	}
	req(admin, `{"type":"delete-sticker","sticker":%d}`)
	if m := next(t, bob); m["type"] != "stickers" || len(m["stickers"].([]any)) != 1 {
		t.Fatalf("after delete: %v", m)
	}
	next(t, admin)
	req(admin, `{"type":"delete-sticker","sticker":%d}`)
	if m := next(t, admin); m["type"] != "error" {
		t.Fatalf("deleted twice: %v", m)
	}
}

// A sticker message: the sticker's id and name, and its name as text for
// what only shows text (Mumble, previews, quotes). No text with it, no
// unknown or deleted sticker, no edit.
func TestSendASticker(t *testing.T) {
	h := newHub(t, "general")
	alice, bob := ready(t, h, "alice"), ready(t, h, "bob")
	s := addSticker(t, h, "wave", "")
	send := func(c *Conn, body string) { c.Handle([]byte(fmt.Sprintf(body, s.ID))) }
	send(alice, `{"type":"send","channel":1,"sticker":%d}`)
	m := next(t, alice)
	msg, _ := m["message"].(map[string]any)
	st, _ := msg["sticker"].(map[string]any)
	if m["type"] != "message" || msg["text"] != "[sticker: wave]" || st["id"] != float64(s.ID) || st["name"] != "wave" {
		t.Fatalf("got %v", m)
	}
	if got := next(t, bob)["message"].(map[string]any)["sticker"]; got == nil {
		t.Fatal("bob got no sticker")
	}
	if got := h.mumbleText(mustMessage(t, h, int64(msg["id"].(float64)))); got != "[sticker: wave]" {
		t.Fatalf("Mumble: %q", got)
	}
	send(alice, `{"type":"send","channel":1,"sticker":%d,"text":"hi"}`)
	if m := next(t, alice); m["type"] != "error" {
		t.Fatalf("with text: %v", m)
	}
	alice.Handle([]byte(`{"type":"send","channel":1,"sticker":9999}`))
	if m := next(t, alice); m["type"] != "error" {
		t.Fatalf("unknown sticker: %v", m)
	}
	alice.Handle([]byte(fmt.Sprintf(`{"type":"edit-message","message":%v,"text":"changed"}`, msg["id"])))
	if m := next(t, alice); m["type"] != "error" {
		t.Fatalf("edit: %v", m)
	}
	h.store.DeleteSticker(s.ID)
	send(alice, `{"type":"send","channel":1,"sticker":%d}`)
	if m := next(t, alice); m["type"] != "error" {
		t.Fatalf("deleted sticker: %v", m)
	}
}

func mustMessage(t *testing.T, h *Hub, id int64) store.Message {
	t.Helper()
	m, err := h.store.MessageByID(id)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// Fixes from a review: slow mode does not give the sticker's text back
// for the composer; a retry of a stored sticker message is confirmed also
// once the sticker is deleted.
func TestStickerSendEdges(t *testing.T) {
	h := newHub(t, "general")
	h.SetVoice(&fakeVoice{}, "")
	admin, bob := readyAs(t, h, 1, "admin"), readyAs(t, h, 2, "bob")
	s := addSticker(t, h, "wave", "")
	bob.Handle(fmt.Appendf(nil, `{"type":"send","channel":1,"sticker":%d,"nonce":"n1"}`, s.ID))
	first := next(t, bob)
	next(t, admin)
	h.store.DeleteSticker(s.ID)
	bob.Handle(fmt.Appendf(nil, `{"type":"send","channel":1,"sticker":%d,"nonce":"n1"}`, s.ID))
	if m := next(t, bob); m["type"] != "message" || m["nonce"] != "n1" || m["message"].(map[string]any)["id"] != first["message"].(map[string]any)["id"] {
		t.Fatalf("retry: %v", m)
	}
	s = addSticker(t, h, "cat", "")
	admin.Handle([]byte(`{"type":"set-slow-mode","channel":1,"seconds":30}`))
	next(t, admin)
	next(t, bob)
	// bob posted just now: his turn has not come.
	bob.Handle(fmt.Appendf(nil, `{"type":"send","channel":1,"sticker":%d}`, s.ID))
	if m := next(t, bob); m["about"] != "slow-mode" || m["text"] != "" {
		t.Fatalf("slow mode: %v", m)
	}
}
