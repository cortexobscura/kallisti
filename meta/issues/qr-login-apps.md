# QR login: the apps' scanner

## Requirements

- Android/iOS app: "Scan a QR code" on the login screen and in the
  account settings (tauri-plugin-barcode-scanner; camera permission
  texts). The code's link also names the server: the app connects to it
  if it has none yet.
- Desktop apps: no scanner; they show codes like the web.

## Acceptance Criteria

- The apps build; the user checks scanning on a phone.
