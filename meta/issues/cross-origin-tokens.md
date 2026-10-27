# Cross-origin access with tokens

Parent: [Join other servers with your account](join-other-servers.md)

## Summary

Servers accept token requests and WebSockets from any origin (cookie
sessions stay limited to ALLOWED_ORIGINS); the web app's CSP allows
connections and pictures from other servers. Safe with tokens: no
cookies, no cross-site forgery.

## Acceptance Criteria

- Tests for each part; checked across two local servers.
