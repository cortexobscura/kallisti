# Server name in the settings

Parent: [Admins change the server's name and logo](server-name-and-logo.md)

## Summary

An admin sets the server's name in the server settings. It is stored in
the database and wins over `SERVER_NAME` (the default) after restarts.
The root channel gets the name at once (Mumble clients see the rename);
the web client shows it in the rail, the channel list header, the tab
title and on the login page.

## Acceptance Criteria

- Admins only; the name validated like a channel name.
- Live for connected clients (web and Mumble), kept after a restart.
- Tests for storage, precedence over `SERVER_NAME`, and the rename.
