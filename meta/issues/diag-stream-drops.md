# Stream frames dropped per viewer

Parent: [More diagnostics](more-diagnostics.md)

## Summary

The stream server drops frames for a viewer whose queue is full (32
frames), sound too, without a trace. Count them per viewer and show the
count in the connections view.

## Acceptance Criteria

- A test: a full viewer's dropped frames are counted.
- The connections view shows them beside the voice frames dropped.
