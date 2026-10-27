package sidecar

import (
	"errors"
	"testing"

	"kallisti.chat/server/internal/homes"
)

// Session 1 is an admin.
func (f *fakeVoice) HomeServers(session uint32) (homes.Rules, error) {
	if session != 1 {
		return homes.Rules{}, errors.New("only admins see this setting")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.homes, nil
}

func (f *fakeVoice) SetHomeServers(session uint32, r homes.Rules) (homes.Rules, error) {
	if session != 1 {
		return homes.Rules{}, errors.New("only admins change this setting")
	}
	r, err := homes.Normalize(r)
	if err == nil {
		f.mu.Lock()
		f.homes = r
		f.mu.Unlock()
	}
	return r, err
}

func TestHomeServerLists(t *testing.T) {
	h := newHub(t, "general")
	h.SetVoice(&fakeVoice{}, "")
	admin := readyAs(t, h, 1, "admin")
	bob := readyAs(t, h, 2, "bob")
	bob.Handle([]byte(`{"type":"set-home-servers","homes":{"mode":"block","hosts":["evil.example"]}}`))
	if m := next(t, bob); m["type"] != "error" || m["about"] != "manage" {
		t.Fatalf("bob: %v", m)
	}
	admin.Handle([]byte(`{"type":"home-servers"}`))
	if m := next(t, admin); m["type"] != "home-servers" || m["homes"].(map[string]any)["hosts"] == nil {
		t.Fatalf("read: %v", m)
	}
	admin.Handle([]byte(`{"type":"set-home-servers","homes":{"mode":"block","hosts":[" Evil.Example "]}}`))
	m := next(t, admin)
	if got := m["homes"].(map[string]any); m["type"] != "home-servers" || got["mode"] != "block" || got["hosts"].([]any)[0] != "evil.example" {
		t.Fatalf("set: %v", m)
	}
}
