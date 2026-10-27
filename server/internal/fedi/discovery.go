package fedi

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const (
	webFingerPath = "/.well-known/webfinger"
	acctPrefix    = "acct:"
	maxDiscovery  = 64 << 10
)

type finger struct {
	username string
	domain   Instance
	actor    string
	server   Instance
}

func (f finger) handle() string { return "@" + f.username + "@" + f.domain.Host }

// parseHandle keeps the account domain separate from the API server.
func parseHandle(input string, dev bool) (string, Instance, error) {
	user, host, ok := strings.Cut(strings.TrimPrefix(input, "@"), "@")
	if !ok || !ValidUsername(user) || strings.ContainsAny(host, "/@?#\\") || strings.TrimSpace(host) != host {
		return "", Instance{}, ErrInstance
	}
	inst, err := ParseInstance(host, dev)
	return user, inst, err
}

// actorServer rejects credentials and fragments rather than silently dropping
// them as ParseInstance does for user-entered profile URLs.
func actorServer(actor string, dev bool) (Instance, error) {
	u, err := url.Parse(actor)
	if err != nil || u.Host == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" ||
		u.Scheme != "https" && !(dev && u.Scheme == "http") {
		return Instance{}, ErrInstance
	}
	return ParseInstance(actor, dev)
}

func (c *Client) webFinger(ctx context.Context, inst Instance, username string, dev bool) (finger, error) {
	resource := acctPrefix + username + "@" + inst.Host
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, inst.Base+webFingerPath+"?resource="+url.QueryEscape(resource), nil)
	if err != nil {
		return finger{}, err
	}
	req.Header.Set("Accept", "application/jrd+json")
	var doc struct {
		Subject string `json:"subject"`
		Links   []struct {
			Rel  string `json:"rel"`
			Type string `json:"type"`
			Href string `json:"href"`
		} `json:"links"`
	}
	if err := c.doLimit(req, &doc, maxDiscovery); err != nil {
		return finger{}, err
	}
	if !strings.HasPrefix(doc.Subject, acctPrefix) {
		return finger{}, fmt.Errorf("%w: missing WebFinger account subject", ErrUnreachable)
	}
	user, domain, err := parseHandle(strings.TrimPrefix(doc.Subject, acctPrefix), dev)
	if err != nil || !strings.EqualFold(user, username) {
		return finger{}, fmt.Errorf("%w: unexpected WebFinger subject", ErrUnreachable)
	}
	f := finger{username: user, domain: domain}
	for _, link := range doc.Links {
		if link.Rel != "self" {
			continue
		}
		media, params, err := mime.ParseMediaType(link.Type)
		if err != nil || media != "application/activity+json" &&
			!(media == "application/ld+json" && params["profile"] == "https://www.w3.org/ns/activitystreams") {
			continue
		}
		server, err := actorServer(link.Href, dev)
		if err != nil || f.actor != "" && f.actor != link.Href {
			return finger{}, fmt.Errorf("%w: invalid WebFinger actor link", ErrUnreachable)
		}
		f.actor, f.server = link.Href, server
	}
	if f.actor == "" {
		return finger{}, fmt.Errorf("%w: no WebFinger actor link", ErrUnreachable)
	}
	return f, nil
}

// confirmSubject asks the account domain to confirm a server's claim. A server
// cannot claim a verified handle on another domain merely by returning it.
func (c *Client) confirmSubject(ctx context.Context, f finger, dev bool) error {
	canonical, err := c.webFinger(ctx, f.domain, f.username, dev)
	if err != nil {
		return err
	}
	if canonical.handle() != f.handle() || canonical.actor != f.actor {
		return fmt.Errorf("%w: WebFinger delegation does not match", ErrUnreachable)
	}
	return nil
}

// ResolveInstance discovers a handle's API host. Bare hosts and URLs keep
// working without WebFinger. Discovery has one shared deadline, at most two
// lookups, bounded redirects, and the same address guard as OAuth requests.
func (c *Client) ResolveInstance(ctx context.Context, input string, dev bool) (Instance, error) {
	input = strings.TrimSpace(input)
	if strings.Contains(input, "/") || !strings.Contains(input, "@") {
		return ParseInstance(input, dev)
	}
	user, domain, err := parseHandle(input, dev)
	if err != nil {
		return Instance{}, ErrInstance
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	f, err := c.webFinger(ctx, domain, user, dev)
	if err != nil {
		return Instance{}, fmt.Errorf("%w: WebFinger lookup: %w", ErrUnreachable, err)
	}
	if f.domain.Host != domain.Host {
		if err := c.confirmSubject(ctx, f, dev); err != nil {
			return Instance{}, fmt.Errorf("%w: %w", ErrUnreachable, err)
		}
	}
	return f.server, nil
}

// DiscoverHandle finds the canonical address of the authenticated user. Callers
// may keep Account.Handle's server-based address if optional discovery fails.
// Neither the initially typed handle nor an unconfirmed cross-domain subject
// is sufficient to label an account as verified.
func (c *Client) DiscoverHandle(ctx context.Context, inst Instance, username string, dev bool) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	f, err := c.webFinger(ctx, inst, username, dev)
	if err != nil {
		return "", err
	}
	if f.server != inst {
		return "", fmt.Errorf("%w: WebFinger actor is not on the authenticated server", ErrUnreachable)
	}
	if f.domain.Host != inst.Host {
		if err := c.confirmSubject(ctx, f, dev); err != nil {
			return "", err
		}
	}
	return f.handle(), nil
}
