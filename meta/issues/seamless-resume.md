# Coming back to the app: no "Reconnecting" and no empty frame

Raised by the user (2026-10-30), with two iOS videos: back in the app, the
last screen shows, then "Connection lost. Reconnecting…" (or the small
sign), an empty channel ("Welcome to #!") for a moment, then the app.

## Findings

- No page reload: the screen is the live page.
- The socket closed in the background, and the time hidden counted as
  lost: the banner showed at once.
- Or the socket was dead without a close: nothing noticed it until the
  close came late.
- A new connection replaced the old view as soon as it was in, before its
  text channels and messages came: the empty frame.

## Summary

- Shown again, the loss time starts again (quiet), with an attempt at once;
  the first attempt of any loss goes at once.
- Shown again, a connection that looks alive gets a ping; without an answer
  in 2 s it counts as lost.
- A new connection replaces the old view once it is ready to show (text
  side known, the open text channel's messages loaded; the session asks
  for them), at the latest after 5 s.

## Acceptance Criteria

- Unit tests for the readiness, the time hidden, the ping on return and the
  first attempt.
- Reproduced and checked in Chromium with a frozen page and 300 ms latency
  (closed and dead sockets): no banner, no empty frame.
