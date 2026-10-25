# Fediverse login: Sharkey works only once

## Summary

Reported by Rain (melonbread.xyz, Sharkey): the first login went
through; every later one (verify in settings, a login in another
browser) ends on Sharkey's "An error has occurred".

Misskey-family servers (Sharkey, Misskey) give a Mastodon app whose
client_id is base64 of an auth session URL (`<base>/auth/<id>`), and
that session is used up by one login. We keep the app per instance and
send the same client_id again.

## Acceptance Criteria

- A client_id that is such a session URL is never used twice: each
  login registers a new app on that server.
- Mastodon, Pleroma and Akkoma apps stay reused as before.
