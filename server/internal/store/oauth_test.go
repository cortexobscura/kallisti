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
	c := OAuthClient{ID: "cid", SecretHash: "sh", Name: "Kallisti", RedirectURI: "https://b.example/api/fedi/callback", Scopes: "read:accounts", Website: ""}
	if err := s.CreateOAuthClient(c, now); err != nil {
		t.Fatal(err)
	}
	if got, err := s.OAuthClient("cid"); err != nil || got.RedirectURI != c.RedirectURI || got.SecretHash != "sh" || got.Used {
		t.Fatalf("client %+v %v", got, err)
	}
	if _, err := s.OAuthClient("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown client: %v", err)
	}
	// Consent per redirect URI, not per client (a site may register again).
	if ok, _ := s.HasOAuthGrant(acc.ID, c.RedirectURI); ok {
		t.Fatal("granted before consent")
	}
	if err := s.GrantOAuth(acc.ID, c.RedirectURI, now); err != nil {
		t.Fatal(err)
	}
	if ok, _ := s.HasOAuthGrant(acc.ID, c.RedirectURI); !ok {
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
	if err := s.MarkOAuthClientUsed("cid", now); err != nil {
		t.Fatal(err)
	}
	// Unused clients go after a day, used ones after a month without a login.
	s.CreateOAuthClient(OAuthClient{ID: "old", SecretHash: "x", Name: "x", RedirectURI: "https://c.example/cb", Scopes: "read:accounts"}, now.Add(-48*time.Hour))
	s.CreateOAuthClient(OAuthClient{ID: "idle", SecretHash: "x", Name: "x", RedirectURI: "https://d.example/cb", Scopes: "read:accounts"}, now.Add(-90*24*time.Hour))
	s.MarkOAuthClientUsed("idle", now.Add(-60*24*time.Hour))
	if n, err := s.DeleteUnusedOAuthClients(now.Add(-24*time.Hour), now.Add(-30*24*time.Hour)); err != nil || n != 2 {
		t.Fatalf("unused: %d %v", n, err)
	}
	if _, err := s.OAuthClient("cid"); err != nil {
		t.Fatal("a used client was deleted")
	}
	// Full: the least recently used goes.
	s.CreateOAuthClient(OAuthClient{ID: "new", SecretHash: "x", Name: "x", RedirectURI: "https://e.example/cb", Scopes: "read:accounts"}, now.Add(time.Hour))
	if err := s.DeleteLeastUsedOAuthClient(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OAuthClient("cid"); !errors.Is(err, ErrNotFound) {
		t.Fatal("not the least recently used went")
	}
	if n, err := s.CountOAuthClients(); err != nil || n != 1 {
		t.Fatalf("count %d %v", n, err)
	}
}
