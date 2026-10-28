# Several servers: fixes from the final review

Parent: [Status of the branch multi-server](multi-server-status.md)

## Summary

The final review of the whole branch (Claude and Astra, 2026-10-28), the
points before merging:

1. An added server whose token ended (30 days, or ended on its site) is
   stuck: back to "Log in on this device", the list unchanged.
2. Comments and docs from earlier designs.
3. A home login ended elsewhere: the added servers' tokens here are logged
   out and forgotten (not only closed).
4. A known remote account's confirmed domain counts for the block list,
   also when the handle lookup fails.
5. Push off ends every added server's subscription, mounted or not.
6. The voice bar of another server shows its recording.
7. The server list takes names as long as remote names.
8. A consent asked for survives a login through another server.
9. A voice invite on another server does not leave the voice you are in.

Later (own step): App.svelte split, one view for both client mounts,
shared helpers (client and server), ending sessions of blocked homes, a
list of allowed sites, App-level tests.

## Done (2026-10-28)

All nine, then a check of the fixes by both: "session expired" only for a
token the store does not know (not a busy database); a confirmed handle
kept on every path, the stored one used only when the lookup fails;
consent pages open no servers; one's own call shown before another
device's. Left (two actions within a second): a delayed home answer after
a quick new login; an invite while another server reconnects.

## Acceptance Criteria

- A test for each point where the logic allows one; all reviewed.
