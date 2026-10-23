# WHIP streams survive a browser refresh

## Summary

A WHIP stream belongs to the account's voice session. A browser refresh ends that session, so the stream ends for everyone, and OBS must reconnect.

## Requirements

- When the stream's voice session goes, the WHIP connection waits a short time for a new voice session of the same account and continues the stream there.
- Viewers who watched it watch the continued stream again by themselves.

## Acceptance Criteria

- Server test: the ingest ends with the session and starts again on the account's new session; the WHIP session ends if none comes in time.
- Client: a watched stream that ends and comes back from the same user in the same channel is watched again.
