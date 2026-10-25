# Android app (prototype)

## Summary

Step 1 of the mobile apps for Android: a Tauri Android build of the same
crate, as for iOS, installed on the user's phone.

## Requirements

- Tauri Android project (gen/android tracked), the user's phone as test
  device, a debug APK installed with adb.
- Microphone permission for the WebView; voice works in the app.
- Voice keeps running with the screen off and in the background: a
  foreground service (type microphone) with a notification ("In voice:
  <channel>") while in voice.

## Acceptance Criteria

- The APK runs on the phone: log in, chat, voice, Kallisti TV with sound.
- Voice in the background and with the screen off for several minutes,
  both directions.

## Notes

- Signing for releases and CI builds come later.
