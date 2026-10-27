package sidecar

import (
	"testing"

	"kallisti.chat/server/internal/mumble"
)

func (f *fakeVoice) IsAdmin(session uint32) bool { return session == 1 }

func TestLogo(t *testing.T) {
	h := newHub(t, "general")
	h.SetVoice(&fakeVoice{}, "")
	admin := readyAs(t, h, 1, "admin")
	bob := readyAs(t, h, 2, "bob")

	for _, req := range []string{`{"type":"logo-token"}`, `{"type":"remove-logo"}`} {
		bob.Handle([]byte(req))
		if m := next(t, bob); m["type"] != "error" || m["about"] != "manage" {
			t.Fatalf("%s: bob got %v", req, m)
		}
	}
	admin.Handle([]byte(`{"type":"logo-token"}`))
	m := next(t, admin)
	token, _ := m["token"].(string)
	if m["type"] != "upload-token" || token == "" {
		t.Fatalf("got %v", m)
	}
	// For one logo upload, nothing else.
	if _, ok := h.RedeemUploadToken(token); ok {
		t.Fatal("a logo token uploads files")
	}
	admin.Handle([]byte(`{"type":"logo-token"}`))
	token, _ = next(t, admin)["token"].(string)
	if author, ok := h.RedeemLogoToken(token); !ok || author != "admin" {
		t.Fatalf("redeem: %q %v", author, ok)
	}
	if _, ok := h.RedeemLogoToken(token); ok {
		t.Fatal("redeemed twice")
	}

	// A new logo: everyone hears its version, late joiners in ready.
	version, err := h.store.SetLogo(map[int][]byte{32: []byte("png")})
	if err != nil {
		t.Fatal(err)
	}
	h.LogoChanged()
	for _, c := range []*Conn{admin, bob} {
		if m := next(t, c); m["type"] != "logo" || m["version"] != version {
			t.Fatalf("got %v", m)
		}
	}
	late := h.Connect()
	t.Cleanup(late.Close)
	late.SetIdentity(mumble.Identity{Session: 4, Name: "late"})
	if m := next(t, late); m["logo"] != version {
		t.Fatalf("ready %v", m)
	}
	admin.Handle([]byte(`{"type":"remove-logo"}`))
	for _, c := range []*Conn{admin, bob} {
		if m := next(t, c); m["type"] != "logo" || m["version"] != "" {
			t.Fatalf("after remove %v", m)
		}
	}
}
