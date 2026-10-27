# Fixes from the multi-server branch that main could use

## Summary

The reviews of the multi-server work (branch `multi-server`) found
weaknesses that are older than it, or useful without it. They are fixed
there and could come to main on their own (cherry-picks, checked again):

- A fediverse callback with a wrong state cancelled a login in progress
  (`4b58e49`).
- A fediverse link that was in progress survived a logout everywhere
  or a password change (part of `966418e`).
- `__Host-` cookies over HTTPS, so a sibling subdomain cannot plant a
  session or a login flow (`053e1d4`, with the clearing of old cookies in
  `966418e`; users log in once more after it).
- Message HTML: no style elements (a message could restyle the page),
  links and titles that say url( kept (`7bfc7e1`, `ec1d8e1`, `8febb92`;
  the rest there matters only with the looser CSP of the branch).
- The login check telling "no login" from "unknown" (`b0ee296`), a late
  401 or logout clearing only the token it used (`285a881`), late logout
  answers acting only on their login (`50498b3`), the login screen's
  answers counting only for their server (in `8e7f59e`, `b38be78`).
- Readable close reasons on the login screen (`b44ff26`, `7bfc7e1`,
  `00c6910`, `208bff6`, `1dd7c0a`).
- Svelte's browser build in tests, so effects can be tested (`541a6c6`).

## Acceptance Criteria

- Each brought over in its own commit with its tests, reviewed again
  (Claude and Astra).
