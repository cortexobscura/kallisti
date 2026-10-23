# Streams: a message when the browser cannot play them

## Summary

Without WebCodecs (VideoDecoder), the client has no stream client, and a click on LIVE does nothing.

## Requirements

- A click on LIVE then shows that this browser cannot play streams, with a hint that Lockdown Mode turns this off.

## Acceptance Criteria

- Unit test for the message choice; checked in a browser without VideoDecoder.
