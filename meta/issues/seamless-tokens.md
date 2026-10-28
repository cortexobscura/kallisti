# Tokens without typing

Parent: [Join other servers with your account](join-other-servers.md)

## Summary

The first click on a server in a new browser gets a token through a short
popup to the home server (approved at once after the first time; the
token comes back by postMessage, deep links in the apps). The home client
registers as the handler of web+kallisti: links, so invite pages offer
one click. Invite links inside a Kallisti app are handled directly.

## Done (2026-10-27)

- The popup (web): the add dialog's "Continue with your account at
  <home>" opens the server's page with ?kallisti-connect=<home>; its
  fediverse login runs as a handoff (no cookie there); the token is made
  when the popup takes it and goes by postMessage to the provider's
  origin only; the dialog takes it only from that popup and origin. The
  popup page has no account or servers of its own.

Still open: the web+kallisti: handler, invite links inside the apps, the
desktop app (no popup there: deep links).

## Acceptance Criteria

- Tests for each part; checked across two local servers.
