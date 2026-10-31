package sidecar

import (
	"encoding/json"
	"strings"
	"testing"

	"kallisti.chat/server/internal/mumble"
)

func TestChannelList(t *testing.T) {
	h, _, _, bob, carol := threadHub(t)
	bob2 := threadUser(t, h, 6, "bob", 20) // bob on another device
	drain(h.all()...)
	none := `{"hidden":{"text":[],"voice":[]},"order":{"text":[],"voice":[]}}`
	if got := marshal(t, readyOf(t, h, 7, "bob", 20)["channelList"]); got != none {
		t.Fatalf("none yet: %s", got)
	}
	do(bob, `{"type":"set-channel-list","list":{"hidden":{"text":[2]},"order":{"voice":[5,3]}}}`)
	want := `{"hidden":{"text":[2],"voice":[]},"order":{"text":[],"voice":[5,3]}}`
	for _, c := range []*Conn{bob, bob2} {
		if m := nextOf(t, c, "channel-list"); marshal(t, m["list"]) != want {
			t.Fatalf("got %v", m)
		}
	}
	empty(t, carol)
	if got := marshal(t, readyOf(t, h, 7, "bob", 20)["channelList"]); got != want {
		t.Fatalf("ready %s", got)
	}
	// Too long, none, or without an account: refused.
	long := "[" + strings.Repeat("1,", maxListIDs) + "1]" // one too many
	refused(t, bob, `{"type":"set-channel-list","list":{"hidden":{"text":%s}}}`, long)
	refused(t, bob, `{"type":"set-channel-list"}`)
	guest := joinAs(t, h, mumble.Identity{Session: 8, Name: "guest"})
	refused(t, guest, `{"type":"set-channel-list","list":{}}`)
	if _, ok := readyOf(t, h, 7, "bob", 20)["channelList"]; !ok {
		t.Fatal("kept after refusals")
	}
}

func marshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
