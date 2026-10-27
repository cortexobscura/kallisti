# Several servers: status of the branch (2026-10-27)

Parent: [Join other servers with your account](join-other-servers.md)

Work on several servers in one client lives on the branch `multi-server`
(main does not have it): bigger than planned, it must not hold up the
rest. test.kallisti.chat runs this branch.

## Done on the branch

- Step 2, other sites connect with tokens (server): every origin, cookie
  only on the server's own origins; guests and new accounts only on the
  own site and in the apps; logins from other sites limited per account
  (counted before the password work, no lockout of the own site); the
  page's CSP allows https: connections, pictures and media; message HTML
  loads nothing remote; `__Host-` cookies over HTTPS; /api/me names the
  session (a fingerprint); a logout can name the session and account it
  means (409 for another).
- Step 1, several servers in the client: a server session unit
  (lib/serverSession), the hub of added servers (lib/serverHub), the saved
  list with its owner (home origin, account, session), the add dialog,
  the rail with all servers, "Try again", push only for the home server.
  Ending rules: another home login, or none (also a login that expired in
  an open tab), ends the previous one's servers: logged out there, tokens
  and list gone; a failed check ends nothing; stale tabs touch only their
  own tokens and list.

## Still to do (decided)

- The rail outside the client (usable while a server connects, fails or
  reconnects; on phones in the channel drawer).
- Voice continues while another server is on screen: voice bar, mute,
  deafen and push-to-talk follow the voice connection; joining voice
  elsewhere leaves it.
- Unread marks, mentions and notifications of servers in the background.
- Steps 3 to 5 (login provider, the list on the home account, tokens
  without typing).

## Accepted (2026-10-27)

Unlikely and self-healing, or not destructive: not handled (a rule for
this work: no handlers for two tabs acting in the same instant).

- A logout and a login in two tabs at nearly the same time: the late
  logout answer can clear the newer login's cookie (log in again). A
  cookie name per login (Chatto) would fix it: the patch was written and
  dropped as not worth its complexity.
- Two tabs writing the saved list in the same instant: one entry can be
  lost. Web Locks alone do not fix it (localStorage is not consistent
  across tab processes); IndexedDB would. Step 4 replaces the list.
- Another owner's empty list record taken over by a stale tab; no "+"
  until the next successful check without a fingerprint; a failed home
  logout still ends the added servers; a guest's logout guarded by its
  name only; a database error during a logout check.
- On the test server and in the browser tests only: the local servers
  use http, so the tests bypass the page's CSP (production is https).
