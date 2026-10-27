# Desktop (Linux, Wayland): GPU use while minimised

## Summary

GitHub #3, follow-up after the hidden-window fix: on KDE Plasma (Wayland)
the app still uses the GPU while minimised, as long as no other window
has focus; with another window focused it drops to zero. Wayland does not
tell a client that it is minimised, so the hidden state may never be set.

## Requirements

- Reproduce in a headless Wayland session (KWin) with the Linux build.
- Find which signals reach the app on minimise: page visibility, Tauri
  window events, focus, GTK/WebKitGTK state, frame callbacks.
- Fix with the simplest signal that works, or document why it cannot be
  done.

## Acceptance Criteria

- Measured before and after (painting stops while minimised), or a clear
  answer with the reason.
- A reply on GitHub #3, text approved by the user.

Closed 2026-10-27: implemented (see git log); not every platform was checked on a real device.
