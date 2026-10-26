# The speaker's microphone in the connections view

Parent: [More diagnostics](more-diagnostics.md)

## Summary

The web client tells the server what the browser says about its
microphone: its name, sample rate, channels, and the filters that run
(the browser's noise suppression, echo cancellation, gain; RNNoise or
DeepFilterNet; music mode). Admins see it per connection. A sample rate
of 16 or 8 kHz shows a Bluetooth headset in its phone mode at once.

## Acceptance Criteria

- Sent when the microphone starts or changes; checked and capped on the
  server; kept per connection, gone with it.
- The connections view shows it (admins only, as the view).
