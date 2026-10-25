# Stream sound without WebCodecs AudioDecoder

## Summary

The iOS app (WKWebView) has no `AudioDecoder`, so watched streams had no
sound (voice played: it falls back to libopus). Stream sound now uses the
voice decoder (WebCodecs with Opus, else libopus).

## Acceptance Criteria

- Kallisti TV plays with sound in the iOS app.
- A test plays stream sound without `AudioDecoder`.

## Notes

- Done: checked on the user's iPhone.
- Going live with sound still needs `AudioEncoder`.
