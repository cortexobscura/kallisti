// Package homes holds which other servers' accounts may log in here (with
// another Kallisti server or a fediverse instance as their home): all by
// default, an admin's block list, or an allow list (then only those).
package homes

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"kallisti.chat/server/internal/fedi"
)

// MaxHosts: the most hosts in a list.
const MaxHosts = 1000

type Rules struct {
	// "all" ("": the same), "block" (all but Hosts and their subdomains),
	// "allow" (only Hosts, exactly).
	Mode  string   `json:"mode"`
	Hosts []string `json:"hosts"`
}

// Allows: an account from this host (a home server's, as fediverse logins
// have it: see fedi.ParseInstance) may log in.
func (r Rules) Allows(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	switch r.Mode {
	case "block":
		return !slices.ContainsFunc(r.Hosts, func(h string) bool { return host == h || strings.HasSuffix(host, "."+h) })
	case "allow":
		return slices.Contains(r.Hosts, host)
	}
	return true
}

// Normalize checks rules from an admin: each host as logins have it (lower
// case, IDN in ASCII, no root dot or :443; from a URL or handle too), once,
// empty lines dropped.
func Normalize(r Rules) (Rules, error) {
	switch r.Mode {
	case "", "all", "block", "allow":
	default:
		return Rules{}, errors.New("unknown mode")
	}
	if r.Mode == "" {
		r.Mode = "all"
	}
	hosts := []string{}
	for _, raw := range r.Hosts {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		inst, err := fedi.ParseInstance(raw, false)
		if err != nil { // a local or test server, with its port
			inst, err = fedi.ParseInstance(raw, true)
		}
		if err != nil || strings.ContainsAny(inst.Host, " ") {
			return Rules{}, fmt.Errorf("%q is not a server's address", strings.TrimSpace(raw))
		}
		if !slices.Contains(hosts, inst.Host) {
			hosts = append(hosts, inst.Host)
		}
	}
	if len(hosts) > MaxHosts {
		return Rules{}, fmt.Errorf("at most %d servers", MaxHosts)
	}
	return Rules{Mode: r.Mode, Hosts: hosts}, nil
}
