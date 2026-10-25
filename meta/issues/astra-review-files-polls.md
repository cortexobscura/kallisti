# Fixes from the Astra review of files and polls

## Summary

Astra reviewed the undeployed work (several files per message, polls).
The user's decision (2026-10-26): fix these, leave the edge cases (an
interrupted poll end, slow batches outliving their tokens, crossing vote
counts, Mumble text size, rollback) alone.

- A poll deleted while it ends (or a vote reads it back) can crash the
  server: guard a poll that is gone.
- Ten long descriptions exceed the bridge's 16 KiB text frame and drop
  the connection (voice too): a larger frame limit.
- Waiting files are kept in IndexedDB only while all together fit one
  file's limit: check each file.
- On a one-file server a new pick does not replace the waiting file.
- Mumble's catch-up of missed direct messages lacks their files.
- A crafted poll reply still pings: polls never ping.
- A poll retry gets a new nonce (a second poll); Cancel works while
  posting; a description typed right before Send is lost on touch
  screens; poll buttons stay active during a time-out; older servers are
  offered polls and file removal.

## Acceptance Criteria

- Each point fixed, with a test where it can have one.
