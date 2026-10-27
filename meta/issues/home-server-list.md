# The server list on the home account

Parent: [Join other servers with your account](join-other-servers.md)

## Summary

The home account stores the list of servers you joined. Servers you
joined read and update it server to server with the login's token (scope:
identity and list). Every site you are logged into shows the full rail;
joining or leaving anywhere changes the list at home.

## Part 1 now (2026-10-27)

All of it ships together, so the list on the device (step 1) is replaced
now, not kept as a prototype:

1. Server: a table of the account's servers (origin, name, order); GET,
   add and remove at /api/servers with the home login (own site and apps;
   not guests); origins checked, a limit per account.
2. Client: the rail's list comes from the home account (at start, after
   own changes, when the page shows again); the device list with its owner
   rules goes. The device keeps only the tokens of the added servers; a
   home logout still logs out there and deletes them, the list stays.
   On a new device the servers show and ask for a login (until step 5).

Part 2 (servers you join update the list) comes with step 3's tokens.

## Acceptance Criteria

- Tests for each part; checked across two local servers.
