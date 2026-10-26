# Stats for a watched stream

Parent: [More diagnostics](more-diagnostics.md)

## Summary

A small overlay on a watched stream ("Stats"): the delay behind the
stream (the playout's lead), pictures and sound received per second,
late pictures and sound, pictures skipped to the next keyframe, lost
sound packets (gaps in the timestamps), and the bitrate.

## Acceptance Criteria

- Counted in the stream client, shown on demand from the stream tile.
- Tests for the counting.
