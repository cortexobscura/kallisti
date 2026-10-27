# Admins change the server's name and logo

## Summary

Admins set the server's name and logo in the server settings, in place
of "Kallisti" and the "K" in the top left. Decisions:

- The name from the settings wins; `SERVER_NAME` is only the default
  until an admin sets one. Stored in the database.
- The name is the root channel's name (Mumble clients see it too), the
  tab title and the login greeting ("Welcome to <name>"). "Kallisti"
  stays where it means the software.
- The logo shows in the server rail, as the tab icon, on the login page
  and as the installed app's icon (browsers cache that one: changes
  show late).

Sub-issues:

- [Server name in the settings](server-name-setting.md)
- [Server logo](server-logo.md)

## Acceptance Criteria

- Both sub-issues done.
