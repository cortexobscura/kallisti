# Tokens without typing

Parent: [Join other servers with your account](join-other-servers.md)

## Summary

The first click on a server in a new browser gets a token through a short
popup to the home server (approved at once after the first time; the
token comes back by postMessage, deep links in the apps). The home client
registers as the handler of web+kallisti: links, so invite pages offer
one click. Invite links inside a Kallisti app are handled directly.

## Acceptance Criteria

- Tests for each part; checked across two local servers.
