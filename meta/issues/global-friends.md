# Friends across servers

Feature request (2026-10-31): a friends list that goes with the user from
server to server.

## Summary

Friends are accounts by their home identity (name@home server: an account
logged in with its home server is known that way on other servers). The
list lives with the user's home account. The client shows, for each
friend, on which of the user's servers they are online (and in which voice
channel), and opens a direct chat there.

## Decisions (2026-10-31)

- A plain list (like following): no request to accept. It shows only what
  the user can see on their servers anyway.
- Presence from the servers the user is connected to (the client sees it
  there already); no server-to-server sharing.

- A friend is an account on one server: its origin, its user id there,
  and the name last seen (the same person on two servers: two entries).

## Design

- Kept with the home account on its home server (`/api/friends`, like
  `/api/servers`: own site only, not for guest accounts); without a home
  account on the device.
- "Add friend" / "Remove friend" in a registered user's menu (any server
  in the rail).
- A Friends view: online ones first (from the servers in the rail: their
  user id online there; voice channel if visible), then the others; each
  with Message (a `/dm/<user id>` link for that server's client) and
  Remove.

## Acceptance Criteria

- Server: store and handler tests (own site, guests refused, limits).
- Client: unit tests for presence matching and the list; checked in a
  browser with two servers.
