package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"

	"kallisti.chat/server/internal/fedi"
	"kallisti.chat/server/internal/store"
)

// providerSetup: P, a Kallisti server as a login provider (TLS, as
// instances must be), and B, a Kallisti server whose fediverse login uses
// it. B is on "localhost", P on 127.0.0.1: the cookie jar keeps them apart.
func providerSetup(t *testing.T) (p *Service, pSrv *httptest.Server, b *Service, bURL string) {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	p = &Service{Store: s, Now: time.Now}
	pSrv = httptest.NewTLSServer(p)
	t.Cleanup(pSrv.Close)
	p.AllowedOrigins = []string{pSrv.URL}
	p.Provider = &Provider{Public: pSrv.URL}
	b, bSrv := setup(t)
	bURL = strings.Replace(bSrv.URL, "127.0.0.1", "localhost", 1)
	client := fedi.NewClient(func(netip.AddrPort) bool { return true }, pSrv.Client().Transport.(*http.Transport).TLSClientConfig, false)
	b.Fedi = &FediLogin{Client: client, Public: bURL, AppName: "B", Dev: true}
	b.AllowedOrigins = []string{bURL}
	return p, pSrv, b, bURL
}

// A Kallisti account at P logs in at B through B's fediverse login: once
// with P's consent, then at once.
func TestKallistiAsALoginProvider(t *testing.T) {
	_, pSrv, _, bURL := providerSetup(t)
	br := newBrowser(t, &httptest.Server{URL: bURL})
	br.c.Transport = pSrv.Client().Transport
	if resp, v := br.do("POST", pSrv.URL+"/api/accounts", map[string]string{"name": "lain", "password": "longpassword"}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("account at P: %d %v", resp.StatusCode, v)
	}
	login := func() *http.Response {
		resp, v := br.do("POST", "/api/fedi/start", map[string]any{"instance": strings.TrimPrefix(pSrv.URL, "https://")})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("start at B: %d %v", resp.StatusCode, v)
		}
		resp, _ = br.do("GET", v["url"].(string), nil)
		return resp
	}
	// First time: P asks (its page, with the request), then the user allows.
	resp := login()
	consent := resp.Header.Get("Location")
	if resp.StatusCode != http.StatusFound || !strings.HasPrefix(consent, "/?authorize=") {
		t.Fatalf("authorize: %d %q", resp.StatusCode, consent)
	}
	query, _ := url.QueryUnescape(strings.TrimPrefix(consent, "/?authorize="))
	resp, v := br.do("GET", pSrv.URL+"/api/oauth/request?"+query, nil)
	if resp.StatusCode != http.StatusOK || v["account"] != "lain" || v["site"] != bURL {
		t.Fatalf("request: %d %v", resp.StatusCode, v)
	}
	resp, v = br.do("POST", pSrv.URL+"/api/oauth/approve", map[string]string{"query": query})
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(v["redirect"].(string), bURL+"/api/fedi/callback?") {
		t.Fatalf("approve: %d %v", resp.StatusCode, v)
	}
	resp, _ = br.do("GET", v["redirect"].(string), nil)
	wantRedirect(t, resp, bURL+"/?fedi=register")
	if resp, v = br.do("POST", "/api/fedi/signup", map[string]any{"name": "lain"}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("signup at B: %d %v", resp.StatusCode, v)
	}
	if resp, v = br.do("GET", "/api/me", nil); v["name"] != "lain" {
		t.Fatalf("me at B: %d %v", resp.StatusCode, v)
	}
	// Again (logged out at B): P approves at once, B knows the account.
	br.do("POST", "/api/logout", map[string]string{})
	resp = login()
	if loc := resp.Header.Get("Location"); resp.StatusCode != http.StatusFound || !strings.HasPrefix(loc, bURL+"/api/fedi/callback?") {
		t.Fatalf("second authorize: %d %q", resp.StatusCode, loc)
	}
	resp, _ = br.do("GET", resp.Header.Get("Location"), nil)
	wantRedirect(t, resp, bURL+"/")
	if _, v = br.do("GET", "/api/me", nil); v["name"] != "lain" {
		t.Fatalf("me at B again: %v", v)
	}
}

// What the token endpoint refuses: a code twice, a wrong verifier, another
// redirect, another client's code, a wrong secret.
func TestProviderCodesAndTokens(t *testing.T) {
	p, pSrv, _, _ := providerSetup(t)
	ctx := context.Background()
	client := fedi.NewClient(func(netip.AddrPort) bool { return true }, pSrv.Client().Transport.(*http.Transport).TLSClientConfig, false)
	inst := fedi.Instance{Host: strings.TrimPrefix(pSrv.URL, "https://"), Base: pSrv.URL}
	const redirect = "https://b.example/api/fedi/callback"
	app, err := client.RegisterApp(ctx, inst, "B", redirect, "")
	if err != nil {
		t.Fatal(err)
	}
	other, _ := client.RegisterApp(ctx, inst, "C", "https://c.example/api/fedi/callback", "")
	acc, _ := p.Store.CreateAccount(store.Account{Name: "lain", PasswordHash: "h", CertPEM: []byte("c"), KeyPEM: []byte("k")}, time.Now())
	code := func(c fedi.App, uri, verifier string) string {
		return p.Provider.issue(acc.ID, c.ClientID, uri, "read:accounts", fedi.Challenge(verifier), time.Now())
	}
	// Registering again gives the same app (it is its redirect URI).
	if again, _ := client.RegisterApp(ctx, inst, "B again", redirect, ""); again.ClientID != app.ClientID {
		t.Errorf("another app for the same redirect: %q, %q", again.ClientID, app.ClientID)
	}
	v := fedi.NewVerifier()
	c1 := code(app, redirect, v)
	token, err := client.Token(ctx, inst, app, redirect, c1, v)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := client.VerifyCredentials(ctx, inst, token); err != nil || got.Username != "lain" {
		t.Fatalf("verify: %+v %v", got, err)
	}
	if _, err := client.Token(ctx, inst, app, redirect, c1, v); err == nil {
		t.Error("a code worked twice")
	}
	if _, err := client.Token(ctx, inst, app, redirect, code(app, redirect, v), fedi.NewVerifier()); err == nil {
		t.Error("a wrong verifier worked")
	}
	if _, err := client.Token(ctx, inst, app, "https://evil.example/api/fedi/callback", code(app, redirect, v), v); err == nil {
		t.Error("another redirect worked")
	}
	if _, err := client.Token(ctx, inst, other, redirect, code(app, redirect, v), v); err == nil {
		t.Error("another client's code worked")
	}
	if _, err := client.Token(ctx, inst, fedi.App{ClientID: other.ClientID, ClientSecret: "x"}, "https://c.example/api/fedi/callback", code(app, redirect, v), v); err == nil {
		t.Error("another app with its own redirect took the code")
	}
	if _, err := client.VerifyCredentials(ctx, inst, "not-a-token"); err == nil {
		t.Error("a made-up token worked")
	}
	// The account's handle, by WebFinger.
	if h, err := client.DiscoverHandle(ctx, inst, "lain", true); err != nil || h != "@lain@"+inst.Host {
		t.Fatalf("webfinger: %q %v", h, err)
	}
}

// The authorize page refuses what it cannot check, without redirecting
// anywhere: an unknown client, another redirect, no PKCE.
func TestAuthorizeRefusesBadRequests(t *testing.T) {
	_, pSrv, _, _ := providerSetup(t)
	ctx := context.Background()
	client := fedi.NewClient(func(netip.AddrPort) bool { return true }, pSrv.Client().Transport.(*http.Transport).TLSClientConfig, false)
	inst := fedi.Instance{Host: strings.TrimPrefix(pSrv.URL, "https://"), Base: pSrv.URL}
	app, _ := client.RegisterApp(ctx, inst, "B", "https://b.example/api/fedi/callback", "")
	good := url.Values{"client_id": {app.ClientID}, "redirect_uri": {"https://b.example/api/fedi/callback"}, "response_type": {"code"}, "scope": {"read:accounts"}, "state": {"s"}, "code_challenge": {fedi.Challenge(fedi.NewVerifier())}, "code_challenge_method": {"S256"}}
	c := &http.Client{Transport: pSrv.Client().Transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	try := func(change func(url.Values)) int {
		q := url.Values{}
		for k, v := range good {
			q[k] = v
		}
		change(q)
		resp, err := c.Get(pSrv.URL + "/oauth/authorize?" + q.Encode())
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if code := try(func(url.Values) {}); code != http.StatusFound {
		t.Fatalf("good: %d", code)
	}
	for name, change := range map[string]func(url.Values){
		"another client": func(q url.Values) { q.Set("client_id", clientIDFor("https://c.example/api/fedi/callback")) },
		"another path": func(q url.Values) {
			q.Set("redirect_uri", "https://b.example/open-redirect")
			q.Set("client_id", clientIDFor("https://b.example/open-redirect"))
		},
		"this server": func(q url.Values) {
			q.Set("redirect_uri", pSrv.URL+"/api/fedi/callback")
			q.Set("client_id", clientIDFor(pSrv.URL+"/api/fedi/callback"))
		},
		"other redirect":  func(q url.Values) { q.Set("redirect_uri", "https://evil.example/api/fedi/callback") },
		"no PKCE":         func(q url.Values) { q.Del("code_challenge") },
		"plain PKCE":      func(q url.Values) { q.Set("code_challenge_method", "plain") },
		"another scope":   func(q url.Values) { q.Set("scope", "write") },
		"no code request": func(q url.Values) { q.Set("response_type", "token") },
	} {
		if code := try(change); code != http.StatusBadRequest {
			t.Errorf("%s: %d, want 400", name, code)
		}
	}
}

// Guests and names other servers do not take are refused before consent;
// provider tokens and session tokens do not work in each other's place.
func TestProviderRefusals(t *testing.T) {
	p, pSrv, _, _ := providerSetup(t)
	ctx := context.Background()
	client := fedi.NewClient(func(netip.AddrPort) bool { return true }, pSrv.Client().Transport.(*http.Transport).TLSClientConfig, false)
	inst := fedi.Instance{Host: strings.TrimPrefix(pSrv.URL, "https://"), Base: pSrv.URL}
	app, _ := client.RegisterApp(ctx, inst, "B", "https://b.example/api/fedi/callback", "")
	query := url.Values{"client_id": {app.ClientID}, "redirect_uri": {"https://b.example/api/fedi/callback"}, "response_type": {"code"}, "state": {"s"}, "code_challenge": {fedi.Challenge(fedi.NewVerifier())}, "code_challenge_method": {"S256"}}.Encode()
	ask := func(name string, temporary bool) int {
		acc, err := p.Store.CreateAccount(store.Account{Name: name, PasswordHash: "h", Temporary: temporary, CertPEM: []byte("c"), KeyPEM: []byte("k")}, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		session := randomHex(32)
		p.Store.CreateSession(tokenHash(session), acc.ID, time.Now().Add(time.Hour))
		req, _ := http.NewRequest("GET", pSrv.URL+"/api/oauth/request?"+query, nil)
		req.AddCookie(&http.Cookie{Name: "__Host-" + CookieName, Value: session})
		resp, err := pSrv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if code := ask("lain", false); code != http.StatusOK {
		t.Fatalf("lain: %d", code)
	}
	if code := ask("visitor", true); code != http.StatusForbidden {
		t.Errorf("a guest: %d, want 403", code)
	}
	if code := ask("[odd]", false); code != http.StatusForbidden {
		t.Errorf("a name others refuse: %d, want 403", code)
	}
	// A session token is no provider token, and the other way round.
	acc, _ := p.Store.AccountByName("lain")
	p.Store.CreateSession(tokenHash("session-token"), acc.ID, time.Now().Add(time.Hour))
	if _, err := client.VerifyCredentials(ctx, inst, "session-token"); err == nil {
		t.Error("a session token read the account as a provider token")
	}
	p.Store.CreateOAuthToken(tokenHash("provider-token"), acc.ID, app.ClientID, "read", time.Now().Add(time.Hour))
	req, _ := http.NewRequest("GET", pSrv.URL+"/api/me", nil)
	req.Header.Set("Authorization", "Bearer provider-token")
	resp, err := pSrv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("a provider token as a session: %d, want 401", resp.StatusCode)
	}
}
