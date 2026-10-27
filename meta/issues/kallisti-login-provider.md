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

## Acceptance Criteria

- Tests for each part; checked across two local servers.
