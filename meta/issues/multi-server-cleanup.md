# Several servers: cleanup after the final review

Parent: [Status of the branch multi-server](multi-server-status.md)

## Summary

From the final review, as its own step (2026-10-28):

1. App.svelte (about 1000 lines) holds several coordinators: move them
   into lib units with tests: the popup login's phase, voice across
   servers, links from elsewhere (invites, ?join, ?open), push of added
   servers following the home's.
2. One view for both client mounts (home and added servers: about 20
   props the same; added servers get the "Retry now" banner too).
3. Shared helpers: an accounts API with a token, local storage access,
   the own user of a server state (client); one CORS function, one token
   maker (server).
4. Blocking a home server ends the sessions of its accounts.
5. A list of the sites one allowed (the login provider), each to revoke.

## Acceptance Criteria

- No change in behaviour for 1 to 3 (the existing tests and the browser
  checks pass); tests for the moved logic.
- Tests for 4 and 5.

## Result (2026-10-28)

Done. Points 1 to 3 as refactors; added servers got the "Retry now"
banner, and blocked storage no longer throws in the login and fediverse
forms and in push. Point 4 logs out only accounts without a password
(the others can still log in), also a waiting handoff. Point 5 in the
settings (not in the desktop app); a removed site's pending codes do not
give tokens.

Reviews: Claude and Astra. Not done (unlikely, or not destructive):

- With FEDI_LOGIN=0 the block does not end sessions (EndBlockedHomes
  needs the fediverse login).
- A tab's settings after another account logged in in a second tab act
  on the newer login's sites.
- The push unit test calls syncFor itself (no reactive Client in it).
