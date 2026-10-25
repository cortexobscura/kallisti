# Watched streams come back after a reload and a lost connection

## Summary

Reported by the user (2026-10-26): a page reload ends the stream one
watches, and a stream that cuts out (a server restart) is not watched
again when it comes back.

A watched stream that ends is already watched again when the same
user's comes back within 30 s (lib/stream/rewatch), but that memory
lives in the Client, which is new after each reconnect and each reload.

## Acceptance Criteria

- After a reload, the streams one watched are watched again (once the
  voice channel is back), if they are back within 30 s.
- After a lost connection (a server restart), the same, counted from
  the moment the connection was lost.
- Streams the viewer closed are not watched again; nothing older than
  30 s comes back.
