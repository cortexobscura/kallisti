# Stream controls on phones

## Summary

On phones (Android app, also the installed web app) the stream tile's
controls are hard to use: they show on hover, the buttons and the volume
slider are small, and in full screen the bar sits under the status bar,
so the button to leave full screen cannot be reached.

## Requirements

- Touch: a tap on the stream shows or hides its controls (they hide again
  after a few seconds); in focus view a tap no longer switches streams by
  accident while the controls show.
- Touch targets at least 44 px; a volume slider that is easy to drag.
- Full screen: the controls keep clear of the status bar, notch and home
  indicator (safe areas); leaving full screen is always reachable.
- Desktop with a mouse stays as it is.

## Acceptance Criteria

- Checked on the user's Android phone and in the iOS simulator (PWA):
  controls reachable in normal view and in full screen.
- Tests for the show/hide logic.
