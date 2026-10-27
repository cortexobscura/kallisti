# Push from added servers

Parent: [Status of the branch multi-server](multi-server-status.md)

## Summary

Added servers notify while the page is open (their clients stay mounted).
With the page closed only the home server pushes: lib/push has one
subscription per browser (scope "/", cookie login). Each added server
needs its own service worker registration (a scope per server, as
Chatto's FDR-013: one subscription per server key), subscribing with its
token, and a click that opens the home site with that server shown (the
service worker posts or opens "/?open=<link>").

## Acceptance Criteria

- A mention on an added server pushes with the page closed; its click
  shows that server and the message.
- Removing the server or a home logout ends its subscription.
