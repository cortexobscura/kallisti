package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"kallisti.chat/server/internal/names"
	"kallisti.chat/server/internal/store"
)

// maxAccountServers: the most servers in an account's list.
const maxAccountServers = 100

// servers: the account's list of servers it joined (the rail; see
// store.AccountServers: each with the account's name there), from our own
// site and the apps; guests have none
// (nothing of theirs survives on a shared device).
func (a *Service) servers(w http.ResponseWriter, r *http.Request) {
	if !a.ownSite(w, r, "the server list is only on your home server's own site") {
		return
	}
	acc, ok := a.AccountFor(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not logged in"})
		return
	}
	if acc.Temporary {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "set a password first: guest accounts keep no server list"})
		return
	}
	// account: the one the client means (a stale window after another
	// login in this browser must not change that one's list).
	var body struct {
		store.AccountServer
		Account string `json:"account"`
	}
	if r.Method != http.MethodGet && json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body) != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if r.Method == http.MethodGet {
		body.Account = r.URL.Query().Get("account")
	}
	if body.Account != "" && names.Key(body.Account) != names.Key(acc.Name) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "another login is active here now"})
		return
	}
	var err error
	switch r.Method {
	case http.MethodGet:
		var list []store.AccountServer
		if list, err = a.Store.AccountServers(acc.ID); err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"servers": list})
			return
		}
	case http.MethodPost:
		body.Name = strings.TrimSpace(body.Name)
		if !a.otherOrigin(body.Origin) || !validListName(body.Name) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "not a server's address and name"})
			return
		}
		if err = a.Store.AddAccountServer(acc.ID, body.AccountServer, maxAccountServers); errors.Is(err, store.ErrTooManyServers) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "your list is full"})
			return
		} else if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{})
			return
		}
	case http.MethodDelete:
		if err = a.Store.RemoveAccountServer(acc.ID, body.Origin); err == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	default:
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "cannot change the server list"})
}

// otherOrigin: a canonical web origin (scheme and lower-case host, no
// path) of another server than this one.
func (a *Service) otherOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || len(origin) > 300 {
		return false
	}
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err != nil || n < 1 || n > 65535 {
			return false
		}
	}
	defaultPort := map[string]string{"https": "443", "http": "80"}[u.Scheme]
	return origin == u.Scheme+"://"+strings.ToLower(u.Host) && u.Port() != defaultPort && !strings.HasSuffix(u.Hostname(), ".") &&
		!slices.Contains(a.AllowedOrigins, origin)
}

// validListName: the account's name on that server (its rules may differ
// from ours: loosely checked).
func validListName(name string) bool {
	return name != "" && utf8.RuneCountInString(name) <= 64 && !strings.ContainsFunc(name, unicode.IsControl)
}
