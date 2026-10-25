package fedi

import (
	"context"
	"testing"
)

func TestDiscoveryProductionURLs(t *testing.T) {
	for _, actor := range []string{
		"http://social.example/users/alice", "https://127.0.0.1/users/alice",
		"https://[::1]/users/alice", "https://localhost/users/alice",
		"https://social.example:8443/users/alice", "https://user:pass@social.example/users/alice",
		"/users/alice", "https://social.example/users/alice#fragment",
	} {
		if _, err := actorServer(actor, false); err == nil {
			t.Errorf("accepted unsafe actor URL %q", actor)
		}
	}
	inst, err := actorServer("https://Social.Example:443/users/alice", false)
	if err != nil || inst.Host != "social.example" || inst.Base != "https://social.example" {
		t.Fatalf("HTTPS actor: %+v %v", inst, err)
	}
	if _, err := actorServer("http://127.0.0.1:8080/users/alice", true); err != nil {
		t.Fatalf("development HTTP actor: %v", err)
	}
}

func TestDiscoveryHandleSyntax(t *testing.T) {
	c := NewClient(nil, nil, false)
	for _, input := range []string{"@@alice@example.org", "alice@@example.org", "@example.org", "a b@example.org", "alice@example.org?bad", "alice@localhost", "alice@127.0.0.1", "alice@example.org:444", "alice@"} {
		if _, err := c.ResolveInstance(context.Background(), input, false); err != ErrInstance {
			t.Errorf("invalid handle %q: %v", input, err)
		}
	}
	user, domain, err := parseHandle("@Alice@BÜCHER.example", false)
	if err != nil || user != "Alice" || domain.Host != "xn--bcher-kva.example" {
		t.Fatalf("normalized handle: %q %+v %v", user, domain, err)
	}
}
