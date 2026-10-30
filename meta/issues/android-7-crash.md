# The Android app closes at once on Android 7

Raised by a user (2026-10-30): on a very old phone (Android 7) the APK
installs, opens and closes at once.

## Summary

Reproduce in the emulator (Android 7.0, API 24, arm64), find the crash in
the log, and fix it if it is in our code or a setting. If Tauri or the
system WebView cannot run there, say so in the Android docs.

## Acceptance Criteria

- The crash is reproduced and its cause known.
- Fixed, or the Android docs give the lowest working version.
