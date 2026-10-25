package fedi_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"kallisti.chat/server/internal/fedi"
	"kallisti.chat/server/internal/fedi/fedifake"
)

func fingerJSON(t *testing.T, w http.ResponseWriter, subject, actor, media string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/jrd+json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"subject": subject,
		"links":   []map[string]string{{"rel": "self", "type": media, "href": actor}},
	}); err != nil {
		t.Errorf("write WebFinger response: %v", err)
	}
}

func TestResolveInstanceDiscovery(t *testing.T) {
	for _, media := range []string{"application/activity+json", `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`} {
		t.Run(media, func(t *testing.T) {
			api, inst := fakeInstance(t, &fedifake.Instance{})
			var domain string
			var requests atomic.Int32
			apex := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.Header.Get("Accept") != "application/jrd+json" || r.Header.Get("Authorization") != "" {
					t.Errorf("discovery headers: %v", r.Header)
				}
				if r.URL.Query().Get("resource") != "acct:alice@"+domain {
					t.Errorf("resource: %s", r.URL.RawQuery)
				}
				if r.URL.Path == "/.well-known/webfinger" {
					http.Redirect(w, r, "/discovery?"+r.URL.RawQuery, http.StatusFound)
					return
				}
				fingerJSON(t, w, "acct:alice@"+domain, api.URL+"/users/alice", media)
			}))
			defer apex.Close()
			domain = strings.TrimPrefix(apex.URL, "https://")
			c := testClient(apex)
			for _, input := range []string{"@alice@" + domain, "  alice@" + domain + "  "} {
				got, err := c.ResolveInstance(context.Background(), input, true)
				if err != nil || got != inst {
					t.Fatalf("resolve %q: %+v %v", input, got, err)
				}
			}
			if requests.Load() != 4 {
				t.Fatalf("unexpected lookup count: %d", requests.Load())
			}
			// Direct servers and profile URLs do not need discovery.
			for _, input := range []string{api.URL, api.URL + "/@alice"} {
				got, err := c.ResolveInstance(context.Background(), input, true)
				if err != nil || got != inst || requests.Load() != 4 {
					t.Fatalf("direct server: %+v %v", got, err)
				}
			}
		})
	}
}

func TestDiscoveryRejectsInvalidDocuments(t *testing.T) {
	for _, failure := range []string{"missing-subject", "wrong-user", "wrong-domain", "missing-link", "wrong-type", "relative", "credentials", "fragment", "conflicting-links", "too-large", "not-json", "not-found"} {
		t.Run(failure, func(t *testing.T) {
			var domain string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				subject, actor, media := "acct:alice@"+domain, "https://social.example/users/alice", "application/activity+json"
				switch failure {
				case "missing-subject":
					subject = ""
				case "wrong-user":
					subject = "acct:bob@" + domain
				case "wrong-domain":
					subject = "acct:alice@elsewhere.example"
				case "missing-link":
					actor = ""
				case "wrong-type":
					media = "text/html"
				case "relative":
					actor = "/users/alice"
				case "credentials":
					actor = "https://user:secret@social.example/users/alice"
				case "fragment":
					actor += "#other"
				case "conflicting-links":
					if _, err := fmt.Fprintf(w, `{"subject":%q,"links":[{"rel":"self","type":"application/activity+json","href":%q},{"rel":"self","type":"application/activity+json","href":"https://other.example/users/alice"}]}`, subject, actor); err != nil {
						t.Errorf("write conflicting actor links: %v", err)
					}
					return
				case "too-large":
					if _, err := fmt.Fprintf(w, `{"subject":%q,"padding":%q}`, subject, strings.Repeat("a", 65<<10)); err != nil {
						t.Errorf("write oversized response: %v", err)
					}
					return
				case "not-json":
					if _, err := fmt.Fprint(w, "not JSON"); err != nil {
						t.Errorf("write malformed response: %v", err)
					}
					return
				case "not-found":
					http.NotFound(w, r)
					return
				}
				fingerJSON(t, w, subject, actor, media)
			}))
			defer srv.Close()
			domain = strings.TrimPrefix(srv.URL, "https://")
			if _, err := testClient(srv).ResolveInstance(context.Background(), "@alice@"+domain, true); err == nil {
				t.Fatal("accepted invalid discovery document")
			}
		})
	}
}

func TestDiscoverHandleConfirmsTheAccountDomain(t *testing.T) {
	for _, failure := range []string{"", "other-user", "other-actor", "other-subject", "unavailable", "foreign-actor"} {
		t.Run(failure, func(t *testing.T) {
			var apiURL, domain string
			apex := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				subject, actor := "acct:alice@"+domain, apiURL+"/users/alice"
				switch failure {
				case "other-actor":
					actor += "-else"
				case "other-subject":
					subject = "acct:alice@other.example"
				case "unavailable":
					http.NotFound(w, r)
					return
				}
				fingerJSON(t, w, subject, actor, "application/activity+json")
			}))
			defer apex.Close()
			domain = strings.TrimPrefix(apex.URL, "https://")
			api := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				subject, actor := "acct:alice@"+domain, apiURL+"/users/alice"
				if failure == "other-user" {
					subject = "acct:bob@" + domain
				}
				if failure == "foreign-actor" {
					actor = apex.URL + "/users/alice"
				}
				fingerJSON(t, w, subject, actor, "application/activity+json")
			}))
			defer api.Close()
			apiURL = api.URL
			inst, err := fedi.ParseInstance(apiURL, true)
			if err != nil {
				t.Fatal(err)
			}
			c := testClient(api)
			handle, err := c.DiscoverHandle(context.Background(), inst, "alice", true)
			if failure != "" {
				if err == nil || handle != "" {
					t.Fatalf("unconfirmed handle: %q %v", handle, err)
				}
				return
			}
			if err != nil || handle != "@alice@"+domain {
				t.Fatalf("canonical handle: %q %v", handle, err)
			}
			// A server-domain handle may return its canonical account
			// domain, provided that domain confirms the same actor.
			resolved, err := c.ResolveInstance(context.Background(), "@alice@"+inst.Host, true)
			if err != nil || resolved != inst {
				t.Fatalf("server-domain alias: %+v %v", resolved, err)
			}
		})
	}
}

func TestDiscoveryNetworkLimits(t *testing.T) {
	for _, redirects := range []int{3, 4} {
		t.Run(fmt.Sprint(redirects), func(t *testing.T) {
			var domain string
			var count atomic.Int32
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if int(count.Add(1)) <= redirects {
					http.Redirect(w, r, r.URL.String(), http.StatusFound)
					return
				}
				fingerJSON(t, w, "acct:alice@"+domain, "https://social.example/users/alice", "application/activity+json")
			}))
			defer srv.Close()
			domain = strings.TrimPrefix(srv.URL, "https://")
			_, err := testClient(srv).ResolveInstance(context.Background(), "alice@"+domain, true)
			wantError := redirects == 4
			if failed := err != nil; failed != wantError || count.Load() != 4 {
				t.Fatalf("redirects: count=%d error=%v", count.Load(), err)
			}
		})
	}
	t.Run("cancelled", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer srv.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, err := testClient(srv).ResolveInstance(ctx, "alice@"+strings.TrimPrefix(srv.URL, "https://"), true)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline: %v", err)
		}
	})
	t.Run("address-guard", func(t *testing.T) {
		var reached atomic.Bool
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reached.Store(true)
		}))
		defer srv.Close()
		c := fedi.NewClient(nil, srv.Client().Transport.(*http.Transport).TLSClientConfig, false)
		_, err := c.ResolveInstance(context.Background(), "alice@"+strings.TrimPrefix(srv.URL, "https://"), true)
		if err == nil || reached.Load() {
			t.Fatalf("private address reached: %v", err)
		}
	})
	t.Run("redirect-guard", func(t *testing.T) {
		var reached atomic.Bool
		blocked := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reached.Store(true)
		}))
		defer blocked.Close()
		start := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, blocked.URL, http.StatusFound)
		}))
		defer start.Close()
		allowed := netip.MustParseAddrPort(strings.TrimPrefix(start.URL, "https://"))
		c := fedi.NewClient(func(ap netip.AddrPort) bool { return ap == allowed }, start.Client().Transport.(*http.Transport).TLSClientConfig, false)
		_, err := c.ResolveInstance(context.Background(), "alice@"+allowed.String(), true)
		if err == nil || reached.Load() {
			t.Fatalf("redirect bypassed address guard: %v", err)
		}
	})
}
