# Names of accounts from other servers

Parent: [Status of the branch multi-server](multi-server-status.md)

## Summary

From testing (2026-10-28):
- "Add a server" says "A Kallisti server where you have an account": not
  true, the popup makes one. Any Kallisti server.
- A login through another server (another Kallisti, or the fediverse)
  asks for a new name. It should not: the name is the one at home with the
  home's host, as the fediverse writes it (alice@chat.example.org). The
  web app shows the host part faded (as Discord shows its numbers).

Decisions: local names (password, guest, Mumble) no longer take "@", so
nobody can pose as alice@chat.example.org; existing ones stay. A name
taken already (rare) falls back to the name step. Mentions of @alice do
not match @alice@host. The browser check for new accounts stays (it runs
by itself, no click).

Accepted (from the reviews): the name stays when the handle changes at
home (a later owner of that handle gets the name step); such an account
cannot unlink. With an external Murmur, its own clients are not checked
(the bridge refuses guests with "@"). Before deploying to a server with
older accounts: check for names with "@" (they would look remote).

## Acceptance Criteria

- Tests: remote names, no "@" in local names, mentions, signup without a
  name (and its fallback).
- Checked in the browser across two servers.
