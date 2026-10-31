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

## Acceptance Criteria

- To be set with the design (identity matching across servers, where the
  list is kept).
