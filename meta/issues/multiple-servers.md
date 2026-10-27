# Several servers in one client

Parent: [Join other servers with your account](join-other-servers.md) (step 1)

## Summary

The client holds a session token and a connection per server (Kallisti
servers; plain Mumble servers in the desktop app later), shows them in the
rail, keeps text and unread marks per server, and one voice channel across
all of them: voice goes on while you switch. "Add a server" opens that
server's normal login (password, guest, fediverse) inside the client. Web
and apps.

## Decisions (2026-10-27)

- All added servers stay connected (unread marks, mentions, instant
  switching); users see you online on each.
- From another site only "Log in" works (step 2: new accounts and guests
  on a server's own site); joining with the home account is step 3.

## Steps

1. A server session of its own: App's logic for its one connection
   (reconnect, place, streams, drafts, files, errors) moves into a unit
   that App holds; no change for users. Tests for the unit.
2. A list of sessions and the rail: switching shows another session's
   client; the others stay connected.
3. "Add a server": address, that server's login (token mode), the list
   and tokens kept on the device; removing a server, logging out of one.
4. One voice channel across servers; unread marks, mentions and
   notifications of servers in the background.

## Acceptance Criteria

- Add, switch and remove servers; the list kept on the device (the home
  account takes it over in step 4).
- Joining voice on one server leaves voice on another.
- Tests for the per-server state and the single voice channel.
