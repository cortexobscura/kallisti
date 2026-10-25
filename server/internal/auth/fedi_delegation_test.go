package auth

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"

	"kallisti.chat/server/internal/fedi"
	"kallisti.chat/server/internal/fedi/fedifake"
)

// The account domain serves discovery only; the API is on another host.
func delegatedFediSetup(t *testing.T) (*Service, *httptest.Server, *fedifake.Instance, string, string) {
	t.Helper()
	a, srv := setup(t)
	f := &fedifake.Instance{}
	var apiURL, domain string
	discovery := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("OAuth credentials sent to WebFinger")
		}
		resource := r.URL.Query().Get("resource")
		user, _, ok := strings.Cut(strings.TrimPrefix(resource, "acct:"), "@")
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"subject": "acct:" + user + "@" + domain,
			"links":   []map[string]string{{"rel": "self", "type": "application/activity+json", "href": apiURL + "/users/" + user}},
		})
	}
	api := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/webfinger" {
			discovery(w, r)
			return
		}
		f.ServeHTTP(w, r)
	}))
	t.Cleanup(api.Close)
	apiURL = api.URL
	apex := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/webfinger" {
			t.Errorf("account domain received an API request: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		discovery(w, r)
	}))
	t.Cleanup(apex.Close)
	domain = strings.TrimPrefix(apex.URL, "https://")
	a.Fedi = &FediLogin{
		Client: fedi.NewClient(func(netip.AddrPort) bool { return true }, api.Client().Transport.(*http.Transport).TLSClientConfig, false),
		Public: srv.URL, AppName: "Kallisti", Dev: true,
	}
	home = srv.URL + "/"
	return a, srv, f, api.URL, domain
}

func TestFediDelegatedLogin(t *testing.T) {
	a, srv, f, inst, domain := delegatedFediSetup(t)
	b := newBrowser(t, srv)
	start := map[string]any{"instance": "@alice@" + domain}
	resp, v := b.do("POST", "/api/fedi/start", start)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delegated start: %d %v", resp.StatusCode, v)
	}
	authorize, err := url.Parse(v["url"].(string))
	if err != nil || "https://"+authorize.Host != inst {
		t.Fatalf("authorize on wrong server: %v %v", authorize, err)
	}
	// The approved account, not the typed handle, owns the login.
	callback, err := f.Approve(authorize.String(), "bob")
	if err != nil {
		t.Fatal(err)
	}
	resp, _ = b.do("GET", callback, nil)
	wantRedirect(t, resp, home+"?fedi=register")
	if resp, v := b.do("GET", "/api/fedi/signup", nil); resp.StatusCode != http.StatusOK || v["handle"] != "@bob@"+domain {
		t.Fatalf("authenticated handle: %d %v", resp.StatusCode, v)
	}
	if resp, v := b.do("POST", "/api/fedi/signup", map[string]any{"name": "bob"}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("signup: %d %v", resp.StatusCode, v)
	}
	acc, err := a.Store.AccountByName("bob")
	if err != nil {
		t.Fatal(err)
	}
	id, err := a.Store.FediIdentityOf(acc.ID)
	if err != nil || id.Host != strings.TrimPrefix(inst, "https://") || id.Handle != "@bob@"+domain {
		t.Fatalf("stored identity: %+v %v", id, err)
	}
	// A direct server login must find the same account, not create another.
	other := newBrowser(t, srv)
	wantRedirect(t, other.fediFlow(f, map[string]any{"instance": inst}, "bob"), home)
	if _, v := other.do("GET", "/api/me", nil); v["name"] != "bob" {
		t.Fatalf("direct login: %v", v)
	}
	// The client derives this from the displayed handle. Reauth must use
	// the stored API host instead, without registering an app on the apex.
	wantRedirect(t, b.fediFlow(f, map[string]any{"instance": domain, "purpose": "reauth"}, "bob"), home+"?fedi=reauth")
	if resp, v := b.do("POST", "/api/password", map[string]any{"new": "bobs-password"}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("first password: %d %v", resp.StatusCode, v)
	}
}

func TestFediDelegationPreservesExistingLinks(t *testing.T) {
	a, srv, f, inst, domain := delegatedFediSetup(t)
	b := newBrowser(t, srv)
	b.fediFlow(f, map[string]any{"instance": inst}, "alice")
	if resp, v := b.do("POST", "/api/fedi/signup", map[string]any{"name": "alice"}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("signup: %d %v", resp.StatusCode, v)
	}
	acc, err := a.Store.AccountByName("alice")
	if err != nil {
		t.Fatal(err)
	}
	id, err := a.Store.FediIdentityOf(acc.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Model an account linked before discovery was supported.
	if err := a.Store.UpdateFediHandle(id.Host, id.RemoteID, "@alice@"+id.Host); err != nil {
		t.Fatal(err)
	}
	other := newBrowser(t, srv)
	wantRedirect(t, other.fediFlow(f, map[string]any{"instance": "alice@" + domain}, "alice"), home)
	updated, err := a.Store.FediIdentityOf(acc.ID)
	if err != nil || updated.Host != id.Host || updated.RemoteID != id.RemoteID || updated.Handle != "@alice@"+domain {
		t.Fatalf("existing identity: %+v %v", updated, err)
	}
	if _, v := other.do("GET", "/api/me", nil); v["name"] != "alice" {
		t.Fatalf("existing account: %v", v)
	}
	if _, err := a.Store.FediApp(domain, a.Fedi.redirectURI()); err == nil {
		t.Fatal("app was cached under the account domain")
	}
	// A wrong approved account still cannot authorize a first password.
	wantRedirect(t, other.fediFlow(f, map[string]any{"instance": domain, "purpose": "reauth"}, "bob"), home+"?fedi=error&reason=other")
	if resp, _ := other.do("POST", "/api/password", map[string]any{"new": "alice-password"}); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("wrong account authorized a password: %d", resp.StatusCode)
	}
}

func TestFediDelegatedLink(t *testing.T) {
	_, srv, f, _, domain := delegatedFediSetup(t)
	b := newBrowser(t, srv)
	if resp, v := b.do("POST", "/api/accounts", map[string]any{"name": "local", "password": "local-password"}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("local account: %d %v", resp.StatusCode, v)
	}
	start := map[string]any{"instance": "@alice@" + domain, "link": true, "password": "wrong-password"}
	if resp, _ := b.do("POST", "/api/fedi/start", start); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("link without password: %d", resp.StatusCode)
	}
	start["password"] = "local-password"
	wantRedirect(t, b.fediFlow(f, start, "alice"), home+"?fedi=linked")
	other := newBrowser(t, srv)
	wantRedirect(t, other.fediFlow(f, map[string]any{"instance": "@alice@" + domain}, "alice"), home)
	if _, v := other.do("GET", "/api/me", nil); v["name"] != "local" || v["fedi"].(map[string]any)["handle"] != "@alice@"+domain {
		t.Fatalf("linked login: %v", v)
	}
}
