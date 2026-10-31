package store

import "testing"

func TestChannelList(t *testing.T) {
	s := open(t)
	if p, err := s.ChannelList(4); err != nil || p != "" {
		t.Fatalf("none yet: %q %v", p, err)
	}
	if err := s.SetChannelList(4, `{"hidden":{"text":[2]}}`); err != nil {
		t.Fatal(err)
	}
	s.SetChannelList(4, `{"hidden":{"text":[3]}}`) // replaced
	if p, _ := s.ChannelList(4); p != `{"hidden":{"text":[3]}}` {
		t.Fatalf("got %q", p)
	}
	if p, _ := s.ChannelList(5); p != "" {
		t.Fatalf("other user: %q", p)
	}
}
