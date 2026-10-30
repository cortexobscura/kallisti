# A message for browsers too old for the app

Found with the Android 7 crash (2026-10-30): on a WebView or browser too
old for the app's code, the page stays empty, with no hint why.

## Summary

A plain script (old syntax, no module) checks for a feature the app needs.
Without it, the page says the browser is too old: in the Android app,
update Android System WebView in the Play Store; elsewhere, update the
browser.

## Acceptance Criteria

- Unit test: the message shows without the feature, not with it; the
  Android app gets the WebView hint.
- Seen in the Android 7 emulator (WebView 52).
