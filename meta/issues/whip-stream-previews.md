# Previews for WHIP streams

## Summary

Stream previews are made by the streaming browser. WHIP streams have no browser, so they never have a preview.

## Requirements

- With ffmpeg available (the same as for media previews), the server makes a small JPEG from a keyframe: soon after the start, then every 15 s, one at a time per stream.
- Without ffmpeg: no previews, as now.

## Acceptance Criteria

- Unit tests for the timing and the ffmpeg arguments; a test with real ffmpeg where it is installed.
