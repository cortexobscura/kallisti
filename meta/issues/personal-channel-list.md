# Personal channel list: hide, show again, own order

Feature request (2026-10-31): users choose which channels of a server they
see (hide, unhide) and in which order.

## Summary

Per user and server: a set of hidden channels and an own order, on top of
the server's order. Hidden channels are out of the sidebar, quick switcher
and unread counts; a place lists them to show them again.

## Decisions (2026-10-31)

- Kept with the account on that server (registered users, also guest
  accounts): the same on every device. Without an account: on the device.
- Text and voice channels. A hidden channel with a mention of the user
  shows until it is read.
- The own order sorts channels inside their category (text) or parent
  (voice); new channels keep the server's place after the ordered ones.

## Acceptance Criteria

- Unit tests for the server's storage and wire, and for the client's
  hide/order logic; hidden channels stay reachable (links, search, the
  list to show them again).
