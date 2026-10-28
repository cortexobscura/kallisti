# Remove the fallback for plain-named session cookies

## Summary

The __Host- cookies came with a move of old sessions (auth.go
plainCookiesUntil, sessionCookie, the rename in /api/me), so nobody had to
log in again. After 2026-11-28 all those sessions have ended: remove the
code, its test and store.SessionExpires if nothing else uses it.

## Acceptance Criteria

- The fallback code is gone; the cookie tests pass.
