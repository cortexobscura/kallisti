# A lost connection: quiet for the first 10 seconds

## Summary

The user (2026-10-26): a lost connection is too noisy, and the server is
usually back in a few seconds. For the first 10 seconds nothing
changes but a small grey "Reconnecting…" sign (after 2 seconds); the
app tries again every second. After that, as before: the banner with
"Retry now" and the growing waits.

## Acceptance Criteria

- No banner in the first 10 s of a loss; a retry each second then.
- The small sign shows from 2 s, the banner from 10 s.
- After 10 s the waits start again from the first (1, 2, 5 … s).
