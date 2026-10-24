# README update

## Summary

The README lists the parts and how to deploy, but its feature table, the Mumble section and the deployment notes are from 2026-10-18. Shorten the feature part to a summary with a link to the features page, and bring the technical parts up to date.

## Requirements

- A short feature summary (a few lines) and a link to kallisti.chat/features.
- "What works where": built-in server vs. Murmur vs. Mumble clients, with the newer features (threads, categories, invite-only channels, fediverse login, personal emoji, streams, recording).
- Deployment: UDP 8189 for streams from OBS; the optional settings (ffmpeg, recordings, Web Push); the Cloudflare DNS-01 and behind-proxy variants; the same quick start as the docs.
- The desktop app in the list of parts.

## Acceptance Criteria

- Every claim matches the code and the docs; links work.
