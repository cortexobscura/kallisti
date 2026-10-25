# Log in with a fediverse account

## Summary

Many users come from the fediverse. Let them register, log in, and link existing accounts with a fediverse account through the Mastodon OAuth API (Mastodon, Pleroma/Akkoma, GoToSocial, Pixelfed, Friendica, ...), asking only for the profile.

## Requirements

- Flow: the user enters an instance or `@user@domain`. Handles use WebFinger to discover the API server from the ActivityPub self-link, including when the account domain differs from the server hostname. Bare server addresses and profile URLs still work without discovery. The server registers an app there (`POST /api/v1/apps`, cached per instance), sends the user to the instance's OAuth authorize page with scope `profile` (fallback `read:accounts` for older servers), handles the callback, reads `GET /api/v1/accounts/verify_credentials`.
- Identity: API server host + remote account id (not the name). The account approved in OAuth owns the login, even if a different handle was typed. Existing links and app registrations keep their keys; no database migration is needed.
- Display: discover the authenticated user's canonical handle through the API server's WebFinger endpoint. A different account domain must confirm the same subject and actor, and the actor must belong to the authenticated server. If this optional lookup fails, display `@user@server` instead. First-password confirmation uses the stored API host, not the displayed handle's domain.
- Discovery: at most two WebFinger lookups per discovery operation, with a shared 10-second deadline, at most three redirects per lookup, and a 64 KiB response limit. All requests use the existing HTTPS and address guards. OAuth credentials are never sent to WebFinger endpoints.
- Log in: a linked identity logs in to its Kallisti account (normal web session).
- Register: an unlinked identity creates a new account in one step; the fedi username is suggested and can be changed on a clash. Same new-account rules as normal registration (captcha/message rules apply as they do today, decide sensibly).
- Link and unlink in user settings for existing accounts. An account must keep at least one way to log in (password or a linked identity).
- Accounts without a password can set one in settings (needed for native Mumble clients).
- Opt-in: show the linked handle on the profile as a verified `@user@host`.
- SSRF protection: HTTPS only, no private/loopback/link-local addresses (checked at connect time, also after redirects and DNS), timeouts, response size limits.
- Web client first.

## Acceptance Criteria

- Server tests for the OAuth flow against a fake instance (app registration, state/CSRF check, callback, verify_credentials), linking/unlinking, registration, SSRF guard.
- Separate account/API host fixtures cover delegated login, account switching, canonical handles, existing links, and first-password confirmation. Discovery tests cover subject/actor mismatches, malformed and oversized responses, redirect limits, cancellation, and address guards.
- Client: login button, register step, settings link/unlink, opt-in handle; tests for pure logic.
- Headless browser check against a fake local instance.

## Notes

- Later: the desktop app (deep link or system browser), Misskey's MiAuth.
- Decisions: new accounts and linking both allowed; handle on the profile is opt-in; web first.
