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
