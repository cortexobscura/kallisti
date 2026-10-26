# Firefox: no stream controls on hover

## Summary

A report: in Firefox, nothing shows when the pointer is over a watched
stream, so there is no way to reach full screen, picture-in-picture or
the volume. The controls show on `.tile:hover`, `:focus-within` or a tap
(`.shown`, only when `(hover: none)` matches); see `StreamStage.svelte`.

Possible causes, not checked yet:

- Firefox matches `(hover: none)` on some systems (a touch screen
  laptop, some Linux setups). Then the tools get `pointer-events: none`
  until a tap, and a mouse click picks the stream instead.
- Something over the tile takes the hover in Firefox.

Unknown: the reporter's Firefox version and system.

## Acceptance Criteria

- Reproduced in Firefox (headless), cause found.
- Controls show on hover in Firefox; a test for the cause.
