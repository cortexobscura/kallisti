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

Tried (headless Firefox and Chromium on Linux, the real image, a camera
stream from Chromium's fake camera): with a mouse, Firefox shows the
controls on hover and they work. With a touch screen (`hasTouch`,
`(hover: none)` matches), hover still shows them, but the first click on
a control does nothing (it only shows them); the second works. Firefox
also lays out the sidebar's live badge and stream preview differently
(the badge below the name, the preview at the left edge).

## Acceptance Criteria

- Reproduced in Firefox (headless), cause found.
- Controls show on hover in Firefox; a test for the cause.
