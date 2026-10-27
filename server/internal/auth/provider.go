package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"kallisti.chat/server/internal/config"
	"kallisti.chat/server/internal/fedi"
	"kallisti.chat/server/internal/store"
)

// Kallisti as a login provider (meta/issues/kallisti-login-provider.md):
// the part of the Mastodon API that fediverse logins use, so another
// Kallisti server's fediverse login takes accounts from here unchanged.
// Other servers register an app (/api/v1/apps); the browser comes to
// /oauth/authorize, the user allows the site once (the web app shows the
// question: /?authorize=...), then at once; the site trades the code
// (single use, PKCE S256 only) for a short token that reads the account
// (verify_credentials). WebFinger finds name@host.
//
// An app is its redirect URI: its client id comes from it, nothing is kept
// (no table anyone could fill, no app anyone could take away: sites keep
// the first app that worked). Its secret is not checked: as for public
// clients, PKCE and the exact redirect protect the code.

const (
	codeLifetime  = 2 * time.Minute
	tokenLifetime = 10 * time.Minute
)

// clientIDFor is the app of a redirect URI.
func clientIDFor(redirect string) string {
	sum := sha256.Sum256([]byte("kallisti oauth client " + redirect))
	return hex.EncodeToString(sum[:16])
}

// The scopes we know: all read the account, nothing else.
var providerScopes = []string{"read", "read:accounts", "profile"}

// Provider turns it on (see Service.Provider).
type Provider struct {
	// Public: the server's public address (the issuer; WebFinger's host).
	Public string
	codes  pending[oauthCode]
}

type oauthCode struct {
	account                            int64
	client, redirect, scopes, pkceS256 string
}

// issue keeps a code for the account (single use); "" when full.
func (p *Provider) issue(account int64, client, redirect, scopes, challenge string, now time.Time) string {
	return p.codes.put(oauthCode{account, client, redirect, scopes, challenge}, now, now.Add(codeLifetime))
}

// authRequest is a checked /oauth/authorize request.
type authRequest struct {
	client, redirect, scopes, state, pkce string
}

// site is who asks: the redirect's origin (consent is kept per redirect
// URI: another app on the site with another path asks again).
func (q authRequest) site() string { return originOf(q.redirect) }

func originOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// withQuery is the redirect URI with these parameters added.
func withQuery(raw string, add url.Values) string {
	u, _ := url.Parse(raw)
	q := u.Query()
	for k, v := range add {
		if v[0] != "" {
			q[k] = v
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

var errBadRequest = errors.New("invalid request")

// parseAuthorize checks an authorize request; an error never redirects
// (the redirect URI may be anyone's).
func (a *Service) parseAuthorize(q url.Values) (authRequest, error) {
	requested := strings.Fields(q.Get("scope"))
	if len(requested) == 0 {
		requested = []string{"read"}
	}
	var scopes []string // known ones, once each: bounded
	for _, s := range requested {
		if !slices.Contains(providerScopes, s) {
			return authRequest{}, errBadRequest
		}
		if !slices.Contains(scopes, s) {
			scopes = append(scopes, s)
		}
	}
	r := authRequest{client: q.Get("client_id"), redirect: q.Get("redirect_uri"), scopes: strings.Join(scopes, " "), state: q.Get("state"), pkce: q.Get("code_challenge")}
	if !validRedirect(r.redirect) || originOf(r.redirect) == originOf(a.Provider.Public) || r.client != clientIDFor(r.redirect) || q.Get("response_type") != "code" ||
		q.Get("code_challenge_method") != "S256" || len(r.pkce) != 43 || len(r.state) > 512 {
		return authRequest{}, errBadRequest
	}
	return r, nil
}

// codeFor the account: the redirect URI with a new code.
func (a *Service) codeFor(acc store.Account, q authRequest) (string, bool) {
	code := a.Provider.issue(acc.ID, q.client, q.redirect, q.scopes, q.pkce, a.Now())
	return withQuery(q.redirect, url.Values{"code": {code}, "state": {q.state}}), code != ""
}

// serveProvider answers what other servers ask (no JSON, no cookies: they
// are not browsers); false: not one of its routes.
func (a *Service) serveProvider(w http.ResponseWriter, r *http.Request) bool {
	if a.Provider == nil {
		return false
	}
	get, post := r.Method == http.MethodGet, r.Method == http.MethodPost
	if post {
		config.LimitBody(w)
	}
	switch r.URL.Path {
	case "/api/v1/instance":
		if !get {
			return false
		}
		u, _ := url.Parse(a.Provider.Public)
		title := "Kallisti"
		if a.ServerName != nil {
			title = a.ServerName()
		}
		// Not "4.3" or later: fediverse logins then ask for read:accounts.
		writeJSON(w, http.StatusOK, map[string]any{"uri": u.Host, "title": title, "version": "4.0.0 (compatible; Kallisti)", "registrations": false})
	case "/api/v1/apps":
		if !post {
			return false
		}
		a.registerClient(w, r)
	case "/oauth/authorize":
		if !get {
			return false
		}
		a.authorize(w, r)
	case "/oauth/token":
		if !post {
			return false
		}
		a.token(w, r)
	case "/api/v1/accounts/verify_credentials":
		if !get {
			return false
		}
		a.verifyCredentials(w, r)
	case "/.well-known/webfinger":
		if !get {
			return false
		}
		a.webFinger(w, r)
	default:
		return false
	}
	return true
}

// form reads a form or JSON body (Mastodon takes both).
func form(w http.ResponseWriter, r *http.Request) (url.Values, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if ct := r.Header.Get("Content-Type"); strings.HasPrefix(ct, "application/json") {
		var m map[string]any
		if json.NewDecoder(r.Body).Decode(&m) != nil {
			return nil, false
		}
		v := url.Values{}
		for k, x := range m {
			if s, ok := x.(string); ok {
				v.Set(k, s)
			}
		}
		return v, true
	}
	return r.PostForm, r.ParseForm() == nil
}

// registerClient: an app is its redirect URI (see clientIDFor); the same
// one for each registration.
func (a *Service) registerClient(w http.ResponseWriter, r *http.Request) {
	f, ok := form(w, r)
	name, redirect, scopes, website := strings.TrimSpace(f.Get("client_name")), f.Get("redirect_uris"), f.Get("scopes"), f.Get("website")
	for _, s := range strings.Fields(scopes) {
		ok = ok && slices.Contains(providerScopes, s)
	}
	if !ok || name == "" || utf8.RuneCountInString(name) > 100 || len(website) > 300 || !validRedirect(redirect) {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "invalid app"})
		return
	}
	id := clientIDFor(redirect)
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "name": name, "website": website, "redirect_uri": redirect, "client_id": id, "client_secret": randomHex(32)})
}

// validRedirect: another Kallisti server's fediverse login callback
// (https, http only on this computer for development): that fixed path
// only, so no other page of a site (an open redirect, an address that
// leaks) can take codes in its name; the consent names the site.
func validRedirect(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || len(raw) > 2000 || u.Host == "" || u.Path != "/api/fedi/callback" || u.RawQuery != "" || u.Fragment != "" ||
		u.User != nil || u.Opaque != "" {
		return false
	}
	return u.Scheme == "https" || u.Scheme == "http" && config.IsLocalhost(u.Scheme+"://"+u.Host)
}

// authorize: allowed before, a code at once; else the web app asks (and
// logs in first). Invalid requests never redirect.
func (a *Service) authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	q, err := a.parseAuthorize(r.URL.Query())
	if err != nil {
		http.Error(w, "This login request is not valid.", http.StatusBadRequest)
		return
	}
	if acc, ok := a.AccountFor(r); ok && !acc.Temporary && a.allow(r) {
		if granted, _ := a.Store.HasOAuthGrant(acc.ID, q.redirect); granted && fedi.ValidUsername(acc.Name) {
			if to, ok := a.codeFor(acc, q); ok {
				http.Redirect(w, r, to, http.StatusFound)
				return
			}
		}
	}
	http.Redirect(w, r, "/?authorize="+url.QueryEscape(r.URL.RawQuery), http.StatusFound)
}

// oauthRequest (GET /api/oauth/request?<authorize query>): what the web
// app asks the user; oauthApprove (POST, {query}): allowed, the redirect.
func (a *Service) oauthRequest(w http.ResponseWriter, r *http.Request, approve bool) {
	raw := r.URL.RawQuery
	if approve {
		var body struct {
			Query string `json:"query"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		raw = body.Query
	}
	values, _ := url.ParseQuery(raw)
	q, err := a.parseAuthorize(values)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "this login request is not valid"})
		return
	}
	acc, ok := a.AccountFor(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not logged in"})
		return
	}
	if acc.Temporary {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "set a password first: a guest account cannot log in elsewhere"})
		return
	}
	// Other servers take names as the fediverse makes them (see
	// fedi.ValidUsername).
	if !fedi.ValidUsername(acc.Name) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "other servers do not take names with characters other than letters, digits, _ . and -"})
		return
	}
	if approve && !a.allow(r) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many attempts, try again later"})
		return
	}
	if !approve {
		writeJSON(w, http.StatusOK, map[string]any{
			"site": q.site(), "account": acc.Name,
			"deny": withQuery(q.redirect, url.Values{"error": {"access_denied"}, "state": {q.state}}),
		})
		return
	}
	if err := a.Store.GrantOAuth(acc.ID, q.redirect, a.Now()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "please try again"})
		return
	}
	to, ok := a.codeFor(acc, q)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "please try again"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"redirect": to})
}

func tokenError(w http.ResponseWriter, code string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": code})
}

// token trades a code for a short token (the same app and redirect URI,
// the PKCE verifier).
func (a *Service) token(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	f, ok := form(w, r)
	if !ok || f.Get("grant_type") != "authorization_code" {
		tokenError(w, "unsupported_grant_type")
		return
	}
	id, redirect := f.Get("client_id"), f.Get("redirect_uri")
	if u, _, ok := r.BasicAuth(); ok {
		id = u
	}
	now := a.Now()
	code, ok := a.Provider.codes.get(f.Get("code"), now, true)
	sum := sha256.Sum256([]byte(f.Get("code_verifier")))
	if !ok || id != clientIDFor(redirect) || code.client != id || code.redirect != redirect ||
		subtle.ConstantTimeCompare([]byte(base64.RawURLEncoding.EncodeToString(sum[:])), []byte(code.pkceS256)) != 1 {
		tokenError(w, "invalid_grant")
		return
	}
	token := randomHex(32)
	if err := a.Store.CreateOAuthToken(tokenHash(token), code.account, id, code.scopes, now.Add(tokenLifetime)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server_error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"access_token": token, "token_type": "Bearer", "scope": code.scopes, "created_at": now.Unix(), "expires_in": int(tokenLifetime.Seconds())})
}

func (a *Service) verifyCredentials(w http.ResponseWriter, r *http.Request) {
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	acc, _, err := a.Store.AccountByOAuthToken(tokenHash(token), a.Now())
	if token == "" || err != nil || acc.Temporary {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "The access token is invalid"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": strconv.FormatInt(acc.ID, 10), "username": acc.Name, "acct": acc.Name, "display_name": acc.Name,
		"avatar": "", "avatar_static": "", "url": a.Provider.Public + "/users/" + url.PathEscape(acc.Name),
	})
}

// webFinger: acct:name@host for accounts here (not guests).
func (a *Service) webFinger(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(a.Provider.Public)
	resource := strings.TrimPrefix(r.URL.Query().Get("resource"), "acct:")
	at := strings.LastIndex(resource, "@")
	name, host, ok := resource[:max(at, 0)], resource[at+1:], at > 0
	acc, err := a.Store.AccountByName(name)
	if !ok || !strings.EqualFold(host, u.Host) || err != nil || acc.Temporary || !fedi.ValidUsername(acc.Name) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/jrd+json")
	json.NewEncoder(w).Encode(map[string]any{
		"subject": "acct:" + acc.Name + "@" + u.Host,
		"links": []map[string]string{
			{"rel": "self", "type": "application/activity+json", "href": a.Provider.Public + "/users/" + url.PathEscape(acc.Name)},
		},
	})
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
