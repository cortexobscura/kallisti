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

## Open points from the reviews (not fixed)

- The last two commits (logout naming the account) have not had their
  two reviews yet.
- A delayed logout answer's Set-Cookie can delete a newer login's cookie
  (server; older than this work).
- Two tabs writing the saved list in the same instant are not atomic (a
  lock across tabs, Web Locks, would be needed).
- Another owner's empty list record can be taken over by a stale tab
  (its owner's next add is then refused until a reload).
- Without a session fingerprint (a check failed) there is no "+" until
  the next check succeeds; nothing asks again by itself.
- A home logout that fails still ends the added servers (the user asked
  to log out).
- On the test server and in the browser tests only: the local servers
  use http, so the tests bypass the page's CSP (production is https).
