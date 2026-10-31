# Large GIFs: an MP4 preview, shown like a picture

Raised by the user (2026-10-31): large uploads. Pictures already get WebP
previews; GIFs get none (the animation). Measured on the last week's 7
large GIFs: 32.8 MB as GIF, 3.7 MB as MP4 at most 640 px (WebP: 14.7 MB).

## Summary

The server makes an MP4 (H.264, no sound, at most 640 px) of a large GIF
as its preview. The chat shows it like a picture: playing, looping,
muted, no controls; a click opens the original GIF.

## Acceptance Criteria

- Server tests for the job (which GIFs, the format) and the served type.
- Client tests for the shown element; checked in a browser.
- Existing large GIFs get previews too (if the server makes missing ones).
