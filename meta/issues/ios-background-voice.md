# iOS: does voice keep running in the background?

## Summary

Step 1 of the mobile apps. Find out if the web client inside a Tauri iOS
app (WKWebView) keeps sending and playing voice when the app is in the
background or the phone is locked. The answer decides how the iOS app is
built.

## Requirements

- A Tauri iOS build of the desktop app (same crate), microphone
  permission, UIBackgroundModes audio.
- Join a voice channel, send the app to the background, lock the phone.
- Check on the server side whether voice packets keep coming, and on a
  second client whether the phone still plays what others say.

## Acceptance Criteria

- A clear answer for background and for locked: the microphone keeps
  sending or not, playback keeps going or not, and what stops it.
- If it stops: the options (native audio part, CallKit, other) with an
  estimate.

## Notes

- Test on the user's iPhone (iOS 27); the simulator only for the build
  and first checks.
- No store release, no push in this step.
