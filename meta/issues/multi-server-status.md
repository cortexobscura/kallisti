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
  gone; a failed check ends nothing.
- The list on the home account (/api/servers); the device keeps only the
  tokens; a new device shows the servers and asks for a login there.
- The rail outside the client: the same beside the connecting, error and
  login screens; while a connection is lost the rail (on phones the menu
  to the drawer) still takes input.
- Voice continues while another server is on screen: the clients of all
  connected servers stay mounted (one shown); the others show the voice
  server's bar; mute, deafen and push-to-talk act on it; joining voice
  elsewhere leaves it (and its rejoin after a reconnect).
- Servers in the background: their unread marks and mentions in the rail,
  their notifications while the page is open.
- Step 3, Kallisti as a login provider: the Mastodon OAuth subset that
  fediverse logins use (other Kallisti servers log in with it unchanged),
  a consent screen once per site, then at once; admins' block and allow
  lists of home servers.
- Step 5, logins without typing (web): "Continue with your account" in
  the add dialog; the server's page in a popup logs in through the home
  server and hands its token to the home page (no cookie there).

## Still to do

- Step 4 part 2: servers you join update the list at home, and the rail
  on their own sites (a design question: see home-server-list.md).
- Step 5, the rest: the web+kallisti: link handler, invite links inside
  the apps, the desktop app (no popup: deep links).
- [Push from added servers](push-from-added-servers.md) while the page is
  closed.

## Accepted (2026-10-27)

Unlikely and self-healing, or not destructive: not handled (a rule for
this work: no handlers for two tabs acting in the same instant).

- A logout and a login in two tabs at nearly the same time: the late
  logout answer can clear the newer login's cookie (log in again). A
  cookie name per login (Chatto) would fix it: the patch was written and
  dropped as not worth its complexity.
- A failed home logout still ends the added servers; a guest's logout
  guarded by its name only; a database error during a logout check.
- On the test server and in the browser tests only: the local servers
  use http, so the tests bypass the page's CSP (production is https).
