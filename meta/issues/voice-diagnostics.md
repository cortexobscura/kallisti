# Voice diagnostics: how a speaker reaches the server

## Summary

A user (niffwe) sounded muffled from the web client and fine from a
Mumble client. Nothing showed why. The user (2026-10-26): add this, and
more diagnostics overall.

The built-in server keeps, per speaker, what their voice packets look
like over the last talk (Opus bandwidth from narrow band to full band,
mono or stereo, frame length, bitrate, packets) and which client sends
them. Admins and moderators see it for a user in the app ("Voice
details" in the user menu), with a hint for narrow band (often a
Bluetooth headset in its phone mode).

## Acceptance Criteria

- Server: the stats from the Opus TOC byte, per session, over a recent
  window; tests with packets of each bandwidth.
- A request for a user's stats, for admins and moderators only.
- The web client shows them in a dialog from the user menu.
