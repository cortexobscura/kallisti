package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func jsonBody(v any) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func decode(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var v map[string]any
	json.NewDecoder(resp.Body).Decode(&v)
	return v
}

func TestQRLoginOffer(t *testing.T) {
	a, srv := setup(t)
	resp, _ := post(t, srv, "/api/accounts", map[string]string{"name": "alice", "password": "correct horse"}, nil)
	alice := sessionCookie(resp)

	// The logged-in device shows a code.
	resp, v := post(t, srv, "/api/qr/start", map[string]string{"device": "Firefox on Linux"}, alice)
	token, owner := v["code"].(string), v["owner"].(string)
	if resp.StatusCode != http.StatusOK || v["kind"] != "offer" || len(token) < 40 || len(owner) < 40 || token == owner {
		t.Fatalf("start: %d %v", resp.StatusCode, v)
	}
	wait := func(secret string, c *http.Cookie) map[string]any {
		_, v := post(t, srv, "/api/qr/wait", map[string]string{"secret": secret}, c)
		return v
	}
	if v := wait(owner, alice); v["state"] != "waiting" {
		t.Fatalf("owner waits: %v", v)
	}
	// A phone (logged out) scans it.
	resp, v = post(t, srv, "/api/qr/scan", map[string]string{"token": token, "device": "Kallisti app on iPhone"}, nil)
	claim, _ := v["claim"].(string)
	if resp.StatusCode != http.StatusOK || v["kind"] != "offer" || len(claim) < 40 {
		t.Fatalf("scan: %d %v", resp.StatusCode, v)
	}
	if resp, _ := post(t, srv, "/api/qr/scan", map[string]string{"token": token, "device": "x"}, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("scanned twice: %d", resp.StatusCode)
	}
	if v := wait(owner, alice); v["state"] != "scanned" || v["device"] != "Kallisti app on iPhone" {
		t.Fatalf("owner sees the scan: %v", v)
	}
	if v := wait(claim, nil); v["state"] != "scanned" || v["token"] != nil {
		t.Fatalf("phone waits: %v", v)
	}
	// Only the owner answers, logged in as the account.
	if resp, _ := post(t, srv, "/api/qr/answer", map[string]any{"secret": owner, "yes": true}, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("answer without the session: %d", resp.StatusCode)
	}
	if resp, _ := post(t, srv, "/api/qr/answer", map[string]any{"secret": claim, "yes": true}, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("the phone answers itself: %d", resp.StatusCode)
	}
	if resp, _ := post(t, srv, "/api/qr/answer", map[string]any{"secret": owner, "yes": true}, alice); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("answer: %d", resp.StatusCode)
	}
	// The phone gets its session once (the app asks for the token).
	req, _ := http.NewRequest("POST", srv.URL+"/api/qr/wait", jsonBody(map[string]string{"secret": claim}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Kallisti-Session", "token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	got := decode(t, resp)
	phone := sessionCookie(resp)
	if got["state"] != "done" || got["token"] == nil || phone == nil || got["name"] != "alice" {
		t.Fatalf("phone logs in: %v", got)
	}
	if acc, ok := a.AccountFor(requestWith(phone)); !ok || acc.Name != "alice" {
		t.Fatalf("phone's session: %+v", acc)
	}
	if v := wait(claim, nil); v["state"] != "expired" {
		t.Fatalf("taken twice: %v", v)
	}
	if v := wait(owner, alice); v["state"] != "done" || v["device"] != "Kallisti app on iPhone" {
		t.Fatalf("owner after: %v", v)
	}
}

func TestQRLoginRequest(t *testing.T) {
	a, srv := setup(t)
	resp, _ := post(t, srv, "/api/accounts", map[string]string{"name": "alice", "password": "correct horse"}, nil)
	alice := sessionCookie(resp)

	// A logged-out computer shows a code.
	resp, v := post(t, srv, "/api/qr/start", map[string]string{"device": "Chrome on Windows"}, nil)
	token, owner := v["code"].(string), v["owner"].(string)
	if resp.StatusCode != http.StatusOK || v["kind"] != "request" {
		t.Fatalf("start: %d %v", resp.StatusCode, v)
	}
	// Scanning it needs a login; the logged-in phone sees whom it logs in.
	if resp, _ := post(t, srv, "/api/qr/scan", map[string]string{"token": token, "device": "phone"}, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("scan logged out: %d", resp.StatusCode)
	}
	resp, v = post(t, srv, "/api/qr/scan", map[string]string{"token": token, "device": "Kallisti app on Android"}, alice)
	if resp.StatusCode != http.StatusOK || v["kind"] != "request" || v["device"] != "Chrome on Windows" || v["claim"] != nil {
		t.Fatalf("scan: %d %v", resp.StatusCode, v)
	}
	_, v = post(t, srv, "/api/qr/wait", map[string]string{"secret": owner}, nil)
	if v["state"] != "scanned" || v["device"] != "Kallisti app on Android" {
		t.Fatalf("computer sees the scan: %v", v)
	}
	// The phone confirms (by the token, with its session).
	if resp, _ := post(t, srv, "/api/qr/answer", map[string]any{"secret": token, "yes": true}, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("answer logged out: %d", resp.StatusCode)
	}
	if resp, _ := post(t, srv, "/api/qr/answer", map[string]any{"secret": token, "yes": true}, alice); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("answer: %d", resp.StatusCode)
	}
	resp, v = post(t, srv, "/api/qr/wait", map[string]string{"secret": owner}, nil)
	computer := sessionCookie(resp)
	if v["state"] != "done" || computer == nil || v["name"] != "alice" {
		t.Fatalf("computer logs in: %v", v)
	}
	if acc, ok := a.AccountFor(requestWith(computer)); !ok || acc.Name != "alice" {
		t.Fatalf("computer's session: %+v", acc)
	}
}

func TestQRLoginRefusals(t *testing.T) {
	a, srv := setup(t)
	resp, _ := post(t, srv, "/api/accounts", map[string]string{"name": "alice", "password": "correct horse"}, nil)
	alice := sessionCookie(resp)
	resp, _ = post(t, srv, "/api/accounts", map[string]string{"name": "bob", "password": "correct horse"}, nil)
	bob := sessionCookie(resp)
	a.GuestAccounts = true
	resp, _ = post(t, srv, "/api/guest", map[string]string{"name": "visitor"}, nil)
	guest := sessionCookie(resp)

	// No: the phone learns it, nobody logs in.
	_, v := post(t, srv, "/api/qr/start", map[string]string{"device": "laptop"}, alice)
	token, owner := v["code"].(string), v["owner"].(string)
	_, v = post(t, srv, "/api/qr/scan", map[string]string{"token": token, "device": "phone"}, nil)
	claim := v["claim"].(string)
	if resp, _ := post(t, srv, "/api/qr/answer", map[string]any{"secret": owner, "yes": true}, bob); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("another account answers: %d", resp.StatusCode)
	}
	post(t, srv, "/api/qr/answer", map[string]any{"secret": owner, "yes": false}, alice)
	if resp, v := post(t, srv, "/api/qr/wait", map[string]string{"secret": claim}, nil); v["state"] != "refused" || sessionCookie(resp) != nil {
		t.Fatalf("refused: %v", v)
	}

	// A logged-in device does not change accounts by an offer; the code
	// stays for the device it is meant for.
	_, v = post(t, srv, "/api/qr/start", map[string]string{"device": "laptop"}, alice)
	if resp, _ := post(t, srv, "/api/qr/scan", map[string]string{"token": v["code"].(string), "device": "phone"}, bob); resp.StatusCode != http.StatusConflict {
		t.Fatalf("offer to a logged-in device: %d", resp.StatusCode)
	}
	if resp, _ := post(t, srv, "/api/qr/scan", map[string]string{"token": v["code"].(string), "device": "phone"}, nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("code used up: %d", resp.StatusCode)
	}

	// An owner logged out meanwhile answers nothing.
	_, v = post(t, srv, "/api/qr/start", map[string]string{"device": "laptop"}, alice)
	owner2 := v["owner"].(string)
	post(t, srv, "/api/qr/scan", map[string]string{"token": v["code"].(string), "device": "phone"}, nil)
	post(t, srv, "/api/logout-all", map[string]string{}, alice)
	if resp, _ := post(t, srv, "/api/qr/answer", map[string]any{"secret": owner2, "yes": true}, alice); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("logged out owner: %d", resp.StatusCode)
	}

	// Guests neither show nor confirm.
	if resp, _ := post(t, srv, "/api/qr/start", map[string]string{"device": "x"}, guest); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("guest shows: %d", resp.StatusCode)
	}
	_, v = post(t, srv, "/api/qr/start", map[string]string{"device": "x"}, nil)
	if resp, _ := post(t, srv, "/api/qr/scan", map[string]string{"token": v["code"].(string), "device": "x"}, guest); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("guest confirms: %d", resp.StatusCode)
	}

	// Expiry; long names; unknown secrets.
	_, v = post(t, srv, "/api/qr/start", map[string]string{"device": "x"}, alice)
	old := a.Now
	a.Now = func() time.Time { return old().Add(qrLifetime + time.Second) }
	if resp, _ := post(t, srv, "/api/qr/scan", map[string]string{"token": v["code"].(string), "device": "x"}, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expired: %d", resp.StatusCode)
	}
	a.Now = old
	long := strings.Repeat("a", maxDeviceName+1)
	for _, bad := range []string{long, "\u202eenohp", "a\u200bb", " "} {
		if resp, _ := post(t, srv, "/api/qr/start", map[string]string{"device": bad}, nil); resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("name %q: %d", bad, resp.StatusCode)
		}
	}
	if _, v := post(t, srv, "/api/qr/wait", map[string]string{"secret": "nope"}, nil); v["state"] != "expired" {
		t.Fatalf("unknown: %v", v)
	}
}

// At most maxQRPending codes at once; expired ones make room.
func TestQRGrantLimit(t *testing.T) {
	var q qrGrants
	now := time.Now()
	for i := range maxQRPending {
		if !q.add(&qrGrant{token: fmt.Sprint("t", i), owner: fmt.Sprint("o", i), expires: now.Add(time.Minute)}, now) {
			t.Fatalf("full at %d", i)
		}
	}
	if q.add(&qrGrant{token: "t", owner: "o", expires: now.Add(time.Minute)}, now) {
		t.Fatal("over the limit")
	}
	later := now.Add(2 * time.Minute)
	if !q.add(&qrGrant{token: "t", owner: "o", expires: later.Add(time.Minute)}, later) {
		t.Fatal("expired ones kept their room")
	}
}

// A logout everywhere (or a new password) ends approved codes not taken
// yet: their session is vouched for by the approving one.
func TestQRLoginRevoked(t *testing.T) {
	_, srv := setup(t)
	resp, _ := post(t, srv, "/api/accounts", map[string]string{"name": "alice", "password": "correct horse"}, nil)
	alice := sessionCookie(resp)
	// An offer.
	_, v := post(t, srv, "/api/qr/start", map[string]string{"device": "laptop"}, alice)
	owner := v["owner"].(string)
	_, v2 := post(t, srv, "/api/qr/scan", map[string]string{"token": v["code"].(string), "device": "phone"}, nil)
	post(t, srv, "/api/qr/answer", map[string]any{"secret": owner, "yes": true}, alice)
	// A request.
	_, r := post(t, srv, "/api/qr/start", map[string]string{"device": "computer"}, nil)
	post(t, srv, "/api/qr/scan", map[string]string{"token": r["code"].(string), "device": "phone"}, alice)
	post(t, srv, "/api/qr/answer", map[string]any{"secret": r["code"].(string), "yes": true}, alice)

	post(t, srv, "/api/logout-all", map[string]string{}, alice)
	for _, secret := range []string{v2["claim"].(string), r["owner"].(string)} {
		resp, got := post(t, srv, "/api/qr/wait", map[string]string{"secret": secret}, nil)
		if got["state"] != "refused" || sessionCookie(resp) != nil {
			t.Errorf("after the logout: %v", got)
		}
	}
}

// Taking a session needs a device without one (as scanning does).
func TestQRLoginTakenWithoutSession(t *testing.T) {
	_, srv := setup(t)
	resp, _ := post(t, srv, "/api/accounts", map[string]string{"name": "alice", "password": "correct horse"}, nil)
	alice := sessionCookie(resp)
	resp, _ = post(t, srv, "/api/accounts", map[string]string{"name": "bob", "password": "correct horse"}, nil)
	bob := sessionCookie(resp)
	_, v := post(t, srv, "/api/qr/start", map[string]string{"device": "laptop"}, alice)
	_, s := post(t, srv, "/api/qr/scan", map[string]string{"token": v["code"].(string), "device": "phone"}, nil)
	post(t, srv, "/api/qr/answer", map[string]any{"secret": v["owner"].(string), "yes": true}, alice)
	if resp, _ := post(t, srv, "/api/qr/wait", map[string]string{"secret": s["claim"].(string)}, bob); resp.StatusCode != http.StatusConflict || sessionCookie(resp) != nil {
		t.Fatalf("taken by a logged-in device: %d", resp.StatusCode)
	}
	if _, got := post(t, srv, "/api/qr/wait", map[string]string{"secret": s["claim"].(string)}, nil); got["state"] != "done" {
		t.Fatalf("still there for the phone: %v", got)
	}
}
