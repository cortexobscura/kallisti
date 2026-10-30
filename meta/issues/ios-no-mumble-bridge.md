# iOS app without the Mumble bridge (no export compliance question)

Asked by the user (2026-10-30): App Store Connect asks about encryption for
every build. The bridge brings its own (rustls, Mumble's OCB2-AES), so the
honest answer is not "none" while it is in the iOS app.

## Summary

Leave the bridge (and its crypto crates) out of the iOS build; on iOS the
app connects only to Kallisti servers (WebKit's HTTPS/WSS/WebRTC), the
start screen does not offer plain Mumble servers. Info.ios.plist (and the
generated plist) say `ITSAppUsesNonExemptEncryption` false.

## Acceptance Criteria

- The iOS build has no bridge, rustls or aes code; other platforms unchanged.
- The iOS start screen takes Kallisti servers only, with a clear message for
  a Mumble address.
- TestFlight upload without the export compliance question.
