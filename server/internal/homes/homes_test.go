package homes

import (
	"slices"
	"testing"
)

func TestRules(t *testing.T) {
	// Entries as login hosts are: lower case, IDN in ASCII, no root dot or
	// :443; URLs and handles give their host.
	r, err := Normalize(Rules{Mode: "block", Hosts: []string{" Evil.Example ", "evil.example.", "", "https://bad.example:443/x", "@me@Bücher.example", "chat.local:8443", "http://Straße.example./", "http://xn--bcher-kva.example"}})
	if err != nil || !slices.Equal(r.Hosts, []string{"evil.example", "bad.example", "xn--bcher-kva.example", "chat.local:8443", "xn--strae-oqa.example"}) {
		t.Fatalf("normalize: %+v %v", r, err)
	}
	for _, c := range []struct {
		r    Rules
		host string
		want bool
	}{
		{Rules{}, "any.example", true},
		{Rules{Mode: "all", Hosts: []string{"x.example"}}, "x.example", true},
		{r, "evil.example", false},
		{r, "EVIL.example.", false},
		{r, "social.evil.example", false}, // a block covers subdomains
		{r, "notevil.example", true},
		{r, "good.example", true},
		{Rules{Mode: "allow", Hosts: []string{"home.example"}}, "home.example", true},
		{Rules{Mode: "allow", Hosts: []string{"home.example"}}, "sub.home.example", false}, // an allow is exact
		{Rules{Mode: "allow", Hosts: []string{"home.example"}}, "other.example", false},
		{Rules{Mode: "allow"}, "home.example", false},
	} {
		if got := c.r.Allows(c.host); got != c.want {
			t.Errorf("%+v allows %q: %v, want %v", c.r, c.host, got, c.want)
		}
	}
	for _, bad := range []Rules{{Mode: "maybe"}, {Mode: "block", Hosts: []string{"not a host"}}, {Mode: "block", Hosts: []string{"ftp://x.example"}}} {
		if _, err := Normalize(bad); err == nil {
			t.Errorf("%+v: no error", bad)
		}
	}
	many := make([]string, MaxHosts+1)
	for i := range many {
		many[i] = "h" + string(rune('a'+i%26)) + string(rune('a'+i/26%26)) + string(rune('a'+i/676)) + ".example"
	}
	if _, err := Normalize(Rules{Mode: "block", Hosts: many}); err == nil {
		t.Error("too many hosts: no error")
	}
}
