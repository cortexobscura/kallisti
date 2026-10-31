# QR login: server grants

## Requirements

- POST /api/qr/start {device}: with a session (not a guest): an offer
  (the scanner logs in); without one: a request (the showing device logs
  in). Answers the token (in the code) and an owner secret (to the
  showing device only). In memory, 2 minutes, bounded; the login rate
  limit per IP.
- POST /api/qr/scan {token, device}: an offer: the scanner gets a claim
  secret, the owner sees "scanned by <device>"; a request: the scanner
  must be logged in (not a guest) and gets the showing device's name to
  confirm. Once per code.
- POST /api/qr/answer {yes}: an offer: by its owner (owner secret, still
  logged in as that account); a request: by the scanner (its session,
  the token).
- POST /api/qr/wait {owner | claim}: the state (waiting, scanned,
  refused, expired); the device that logs in gets its session there once
  (the cookie, or the token for the apps), and the code is gone.
- Device names: at most 64 characters, shown as text.

## Acceptance Criteria

- [x] Go tests: both directions, refusals, expiry, single use, guests, wrong
  secrets, an owner logged out meanwhile, limits.
