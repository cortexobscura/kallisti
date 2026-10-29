# Pictures in a grid show tiny for some users

Reported by bajax (2026-10-29): two pictures in one message (with a long
link above them) showed as tiny squares on their screen, normal for
others.

## Summary

The grid is `width: min(520px, 100%)` with square cells: if its parent
sizes to its content, a browser can shrink it to almost nothing. Not
reproduced in Chromium (desktop and phone widths: normal sizes; 270 px at
390 px wide). WebKit could not run in the sandbox. Next: reproduce in
WebKit or the iOS simulator (ask which device bajax uses), then fix.

## Acceptance Criteria

- Reproduced first; the fix checked in the same engine.
