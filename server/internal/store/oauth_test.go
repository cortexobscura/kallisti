package store

import (
	"errors"
	"testing"
	"time"
)

func TestOAuthProvider(t *testing.T) {
	s := open(t)
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	acc, _ := s.CreateAccount(Account{Name: "lain", PasswordHash: "h", CertPEM: []byte("c"), KeyPEM: []byte("k")}, now)
	const redirect = "https://b.example/api/fedi/callback"
	// Consent per redirect URI.
	if ok, _ := s.HasOAuthGrant(acc.ID, redirect); ok {
		t.Fatal("granted before consent")
	}
	if err := s.GrantOAuth(acc.ID, redirect, now); err != nil {
		t.Fatal(err)
	}
	if ok, _ := s.HasOAuthGrant(acc.ID, redirect); !ok {
		t.Fatal("no grant")
	}
	if ok, _ := s.HasOAuthGrant(acc.ID, "https://b.example/other"); ok {
		t.Fatal("another path of the site is granted")
	}
	// Tokens by hash, until they expire.
	if err := s.CreateOAuthToken("th", acc.ID, "cid", "read:accounts", now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	got, scopes, err := s.AccountByOAuthToken("th", now)
	if err != nil || got.ID != acc.ID || scopes != "read:accounts" {
		t.Fatalf("token: %+v %q %v", got, scopes, err)
	}
	if _, _, err := s.AccountByOAuthToken("th", now.Add(11*time.Minute)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired token: %v", err)
	}
}
