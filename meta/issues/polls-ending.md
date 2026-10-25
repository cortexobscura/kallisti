# Polls: ending on time and the result

Parent: [Polls](polls.md)

## Summary

The minute loop ends polls whose time is up (also after a restart): the
poll shows as ended, and a short result message answers it (by the
poll's author, no ping), also to Mumble.

## Acceptance Criteria

- Ending on time, early by the author, and after a restart; once only.
- The result names the winning answer(s) with votes and share, or "no votes".
