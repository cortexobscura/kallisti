# Kallisti as a login provider

Parent: [Join other servers with your account](join-other-servers.md)

## Summary

Every Kallisti server speaks the Mastodon OAuth subset (instance, apps,
authorize, token, verify_credentials) and webfinger, with a consent page
("Let chat.bob.org see you as lain?") that approves at once after the
first time. The login and invite pages offer "Continue with your Kallisti
account" through the existing fediverse login; a linked local account,
"name@home" only on a conflict. Other home servers are accepted by default;
admins set an optional block list and an optional allow list (then only
those) in the server settings.

## Done (2026-10-27)

- Endpoints: /api/v1/instance, /api/v1/apps, /oauth/authorize,
  /oauth/token, verify_credentials, WebFinger (auth/provider.go).
- An app is its redirect URI (client id = its hash, nothing stored, the
  secret not checked: PKCE S256 and the exact redirect protect the code);
  only another server's /api/fedi/callback, never this server's own.
- Consent per account and redirect URI; the web app asks (/?authorize=),
  the approval takes a ticket for exactly the request shown. Codes single
  use, 2 minutes; tokens 10 minutes, by hash. Guests and names other
  servers refuse (fedi.ValidUsername) are refused before the question.
- Block and allow lists (package homes; Server settings): a block covers
  subdomains, an allow is exact; checked at the login's start, callback,
  the handle's domain and the sign-up.

## From Chatto and Fluxer (2026-10-27)

- The authorization code: one use, valid 2 minutes, bound to the client's
  PKCE verifier; a wrong verifier also uses it up (Fluxer's origin handoff).
- The home server keeps only hashes of its tokens (Chatto stores HMACs).
- Short access tokens with a rotating refresh token; a refresh token used
  twice ends the whole login there (Chatto ADR-079). A home logout or a
  block then reaches the other servers within minutes.
- A server that blocks a home server closes the open connections of its
  accounts.

## Acceptance Criteria

- Tests for each part; checked across two local servers.
