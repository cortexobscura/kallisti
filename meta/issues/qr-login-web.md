# QR login: web

## Requirements

- Settings → "Log in on another device": shows the code (offer); after
  a scan, "Log in on <device>?" Yes / No; then done.
- Login screen → "Log in with a QR code": shows the code (request);
  "Confirm on your phone…"; logged in when confirmed.
- The link /qr/<token> (a phone's camera opens it): a logged-out page
  logs in after the owner confirms; a logged-in page asks "Log in on
  <device>?" (with the phishing warning).
- QR drawn as SVG on the client (a small library).

## Acceptance Criteria

- Vitest for the pure parts; Playwright with two browsers, both
  directions.
