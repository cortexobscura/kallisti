# Shorten long links in messages

Raised by the user (2026-10-29): very long links (a proxy link of about
300 characters) fill several lines.

## Summary

Show long links shortened: the host and the start of the path, then "…"
(the full address on hover and when copied; the link itself unchanged).

## Acceptance Criteria

- Unit test for the shortening; the full link still opens and copies.
