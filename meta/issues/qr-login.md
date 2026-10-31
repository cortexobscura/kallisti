# Log in with a QR code

## Summary

One device logs another in by a QR code, so nobody types a password on a
phone. User decisions (2026-10-31): both directions (a logged-in device
shows a code the phone scans; a logged-out device shows a code a
logged-in phone scans); the logged-in device always confirms ("Log in on
<device>?"); the code works in the app's scanner and with the phone's
camera (it is a link).

## Requirements

- Sub-issues: server grants, web (show, confirm, the /qr/ link), the
  apps' scanner.
- A code is a link https://<server>/qr/<token>: valid 2 minutes, once.
- Guests neither give nor get a login by code.

## Acceptance Criteria

- Both directions work between two browsers (Playwright), and with the
  Android/iOS app's scanner (the user checks on a device).

## Notes

- Phishing (a stranger's code shown as "scan to win"): the logged-in
  device's question names the other device and warns to scan only codes
  from one's own screen.
