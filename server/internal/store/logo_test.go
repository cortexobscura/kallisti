package store

import (
	"errors"
	"testing"
)

func TestServerLogo(t *testing.T) {
	s := open(t)
	if v, err := s.LogoVersion(); err != nil || v != "" {
		t.Fatalf("none yet: %q %v", v, err)
	}
	if _, _, err := s.Logo(32); !errors.Is(err, ErrNotFound) {
		t.Fatalf("none yet: %v", err)
	}
	v1, err := s.SetLogo(map[int][]byte{32: []byte("small"), 512: []byte("big")})
	if err != nil || v1 == "" {
		t.Fatalf("set: %q %v", v1, err)
	}
	if got, v, err := s.Logo(512); err != nil || string(got) != "big" || v != v1 {
		t.Fatalf("512: %q %v", got, err)
	}
	if _, _, err := s.Logo(64); !errors.Is(err, ErrNotFound) {
		t.Fatalf("a size not made: %v", err)
	}
	// A new logo replaces every size, with a new version.
	v2, err := s.SetLogo(map[int][]byte{32: []byte("other")})
	if err != nil || v2 == v1 {
		t.Fatalf("again: %q %v", v2, err)
	}
	if _, _, err := s.Logo(512); !errors.Is(err, ErrNotFound) {
		t.Fatalf("the old size stayed: %v", err)
	}
	if v, _ := s.LogoVersion(); v != v2 {
		t.Fatalf("version %q, want %q", v, v2)
	}
	if err := s.DeleteLogo(); err != nil {
		t.Fatal(err)
	}
	if v, _ := s.LogoVersion(); v != "" {
		t.Fatalf("after delete %q", v)
	}
}
