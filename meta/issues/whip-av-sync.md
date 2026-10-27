# WHIP streams: picture and sound out of sync

## Summary

Kallisti TV (ffmpeg over WHIP) is out of sync on some items (music videos, some commercials) and in sync on others; the files themselves are in sync. Sound also went away during one break and came back; after a refresh it played for a second. The bot pairs each picture with its sound and stamps both by count, so the suspect is the server's A/V correction (`whip/sync.go`), which moves pictures by the arrival delay it sees, or the viewer's playout.

## Requirements

- First a measurement: the server logs per WHIP stream, every 10 s, the sound's and the pictures' drift and the correction it applies.
- Then a fix from what the log shows.

## Acceptance Criteria

- The log shows whether the correction moves while the stream is out of sync.
- After the fix, music videos and commercials from Kallisti TV are in sync for a viewer.

## Notes

- See [WHIP streams: sound drops after a while](whip-av-drift.md) for why the correction exists (OBS stamps sound and pictures on different clocks).
- 2026-10-24: the server's drift correction stays within ±5 ms on air (log), also in breaks.
- A local reconstruction (old bot in podman with libass, the same Nadia parts and 13 clips, local server with a trace) played the double clip again, but the viewer's playout (replayed from the trace) dropped no sound, and the RTCP sender reports put the server's picture 6 ms before its sound for the whole run. Not the ads; what differs on air: VAAPI, the NAS's 2 cores, the uplink. The server's clocks start at each track's first arrival, so a late first picture would put every picture late: the log now has `picture_late_ms` from the sender reports.

Closed 2026-10-27: implemented (see git log); not every platform was checked on a real device.
