# Android app: the keyboard pushes the header away

## Summary

Reported by pwm (APK): when the keyboard opens, the whole page moves up
and the channel header goes under the status bar. The app draws edge to
edge (Android 15+), so the keyboard no longer shrinks the window and the
WebView pans the page to show the input.

## Acceptance Criteria

- With the keyboard open the header stays visible and the composer sits
  right above the keyboard (no gap for the gesture bar).
- Checked on an Android phone.
