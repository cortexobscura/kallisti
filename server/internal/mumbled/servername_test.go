package mumbled

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"kallisti.chat/server/internal/mumble"
	"kallisti.chat/server/internal/mumble/pb"
)

func TestServerName(t *testing.T) {
	db := testDB(t)
	setAdminPassword(db, "admin")
	s, addr := startServerAndListen(t, "127.0.0.1", Config{DB: db, RootName: "From env"}, nil)
	su := superUser(t, addr)
	alice, _ := registered(t, addr, "alice")
	if !s.IsAdmin(su.session) || s.IsAdmin(alice.session) {
		t.Fatal("IsAdmin")
	}
	if s.ServerName() != "From env" {
		t.Fatalf("default %q", s.ServerName())
	}
	if _, err := s.SetServerName(alice.session, "Mine"); err == nil || s.ServerName() != "From env" {
		t.Fatal("a user changed the name")
	}
	for _, bad := range []string{"", "   ", strings.Repeat("x", maxChannelName+1), "two\nlines", "bell\a"} {
		if _, err := s.SetServerName(su.session, bad); err == nil {
			t.Fatalf("%q: no error", bad)
		}
	}
	// Freer than channel names: punctuation and emoji.
	for _, ok := range []string{"Lain's place", "kallisti.chat", "Chat & Chill 🎶", "Hi!"} {
		if got, err := s.SetServerName(su.session, ok); err != nil || got != ok {
			t.Fatalf("%q: %q %v", ok, got, err)
		}
		alice.expect(mumble.TypeChannelState)
	}
	if got, err := s.SetServerName(su.session, "  Our place  "); err != nil || got != "Our place" || s.ServerName() != "Our place" {
		t.Fatalf("set: %v %q %q", err, got, s.ServerName())
	}
	// Everyone sees the root channel renamed.
	var m pb.ChannelState
	proto.Unmarshal(alice.expect(mumble.TypeChannelState), &m)
	if m.GetChannelId() != 0 || m.GetName() != "Our place" {
		t.Fatalf("alice saw %v", &m)
	}
	// Kept after a restart, over the env default.
	restarted, err := New(Config{DB: db, RootName: "From env"})
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	if restarted.ServerName() != "Our place" {
		t.Fatalf("after a restart %q", restarted.ServerName())
	}
}
