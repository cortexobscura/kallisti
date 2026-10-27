package mumbled

import (
	"testing"

	"kallisti.chat/server/internal/homes"
)

// Which home servers' accounts log in: admins set it; kept over a restart.
func TestHomeServers(t *testing.T) {
	db := testDB(t)
	setAdminPassword(db, "admin")
	s, addr := startServerAndListen(t, "127.0.0.1", Config{DB: db}, nil)
	su := superUser(t, addr)
	alice, _ := registered(t, addr, "alice")
	if !s.AllowsHome("any.example") {
		t.Fatal("not all by default")
	}
	if _, err := s.HomeServers(alice.session); err == nil {
		t.Fatal("a user read the lists")
	}
	if _, err := s.SetHomeServers(alice.session, homes.Rules{Mode: "block", Hosts: []string{"evil.example"}}); err == nil {
		t.Fatal("a user changed the lists")
	}
	got, err := s.SetHomeServers(su.session, homes.Rules{Mode: "block", Hosts: []string{" Evil.Example "}})
	if err != nil || got.Hosts[0] != "evil.example" {
		t.Fatalf("set: %+v %v", got, err)
	}
	if s.AllowsHome("evil.example") || !s.AllowsHome("good.example") {
		t.Fatal("the block list does not count")
	}
	if _, err := s.SetHomeServers(su.session, homes.Rules{Mode: "nope"}); err == nil {
		t.Fatal("a bad mode was taken")
	}
	restarted, err := New(Config{DB: db})
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	if restarted.AllowsHome("evil.example") {
		t.Fatal("the lists were not kept")
	}
}
