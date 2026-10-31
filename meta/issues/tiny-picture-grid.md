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

## Cause and fix (2026-10-31)

Reproduced in Chromium and WebKit (Playwright): with no text or a short
text, the message body (a flex item) sizes to its content, so
`min(520px, 100%)` resolved against the content. Chromium made the grid
207 px wide; WebKit 210 px until the pictures loaded, 492 px after. A long
text made the body wide, which is why the first tries looked normal.

Fix: `width: 520px; max-width: 100%`. Checked in both engines: 520 px
(desktop, with and without text, loaded or not), 270 px on a 390 px
phone. No unit test: jsdom has no layout; the Playwright check stays
outside the repo.

## Acceptance Criteria

- [x] Reproduced first; the fix checked in the same engine.
