# Watch your own WHIP stream

## Summary

A WHIP stream belongs to the account's voice session, and nobody may watch their own stream. For screen shares that is right (there is a self-preview), but a WHIP stream has none.

## Requirements

- The server lets a session watch its own stream when it is an ingest (WHIP).
- The client allows the click on its own LIVE badge when the stream is from WHIP (not its own screen share).

## Acceptance Criteria

- Server test: own ingest stream may be watched, own screen share may not.
