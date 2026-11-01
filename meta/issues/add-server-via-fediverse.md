# Add a server with the linked fediverse account

## Summary

An account on the home server named after its fediverse home
(Lars@shitposter.world) cannot add another Kallisti server with "+":
the home server refuses to log it in elsewhere (provider.go,
providerAccount), so the popup fails. Recommendation of Astra and
GLM-5.3 (2026-11-01): the popup on the other server logs in with the
linked fediverse account itself, and that server asks before it gives
the home's page the session.

## Requirements

- /api/me tells the client the linked fediverse account's server (the
  stored API host, not the name's host).
- "+" dialog: an account with a linked fediverse account can continue
  with it ("Continue as <handle>"); accounts that can log in elsewhere
  keep "Continue with your account at <home>".
- The popup on the other server: a fediverse login with that server; the
  session goes to the home's page (the address in the popup link), kept
  apart from the login's server.
- Before it gives the session to a page that is not the login's server,
  the other server asks on its own page: "Allow <home> to use your
  account <name> here?" Only a yes hands it over; the question is bound
  to the flow (single use, expiry).
- The other server finds an existing account by the fediverse identity
  (also one with a local name); else a new one (default name: the
  handle). Blocks of fediverse servers (allowsIdentity) apply.

## Acceptance Criteria

- Tests: a remote-named home account adds a server; an existing account
  there is found; a page that was not asked about gets no session; a
  no answer hands over nothing; a blocked fediverse server is refused.
- Checked headless with two local servers and a fake instance.
- Astra review.

## Notes

- Names stay as they are; a fixed handle with a display name is
  issues/handle-and-display-name.md.
