# Stream sound drifts after lost packets (Chrome)

## Summary

Reported during the TV show (2026-10-26): sound drifts from the picture
after a while; closing and opening the stream fixes it. The server's sync
values were fine (a few ms).

Cause, checked in headless Chrome: its AudioDecoder ignores the input
timestamps and counts decoded samples from the first one. After lost
packets the output timestamps do not jump forward (packets 20-29 after a
loss of 10-19 came out at 200-380 ms, not 400-580 ms), so every lost
packet moves the sound 20 ms earlier against the picture. A new decoder
(reopening) starts over.

## Acceptance Criteria

- Each decoded output gets the timestamp of its packet, in order, not
  the decoder's counted one (voice uses the same decoder).
- Test for the mapping; checked in headless Chrome with lost packets.
