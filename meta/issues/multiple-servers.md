# Several servers in one client

Parent: [Join other servers with your account](join-other-servers.md) (step 1)

## Summary

The client holds a session token and a connection per server (Kallisti
servers; plain Mumble servers in the desktop app later), shows them in the
rail, keeps text and unread marks per server, and one voice channel across
all of them: voice goes on while you switch. "Add a server" opens that
server's normal login (password, guest, fediverse) inside the client. Web
and apps.

## Acceptance Criteria

- Add, switch and remove servers; the list kept on the device (the home
  account takes it over in step 4).
- Joining voice on one server leaves voice on another.
- Tests for the per-server state and the single voice channel.
