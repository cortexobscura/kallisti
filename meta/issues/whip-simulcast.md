# Streams from OBS: simulcast

## Summary

OBS 32.1 and newer can send a WHIP stream as up to four encodings at once (simulcast): the same picture at, for example, 100 %, 66 % and 33 % of the size and bitrate. The server could then give each viewer the layer that fits, without encoding anything itself. Now the server refuses such offers ("turn simulcast off in OBS", `server/internal/whip/offer.go`).

## Requirements

- WHIP: accept an offer with simulcast (`a=rid` / `a=simulcast:send`), receive one H.264 track per layer, and ask each layer for keyframes (PLI per SSRC).
- Hub: a stream has layers, each with its own size (`Config`). A viewer has a current layer; a switch waits for a keyframe of the new layer, and the viewer gets the new size with it (as a quality change now).
- Choosing the layer: by the viewer's player size first (a small tile or a phone gets a small layer). Later, also step down when frames pile up for a viewer (the hub already sees full viewers).
- Sound: one track, the same for all layers.
- Without simulcast, streams work as now.
- Previews: from the largest layer's keyframes (see `whip-stream-previews`).

## Acceptance Criteria

- An OBS offer with three layers is accepted; a test sends three layers and two viewers with different player sizes get different layers.
- A layer switch starts at a keyframe, without a broken picture.
- Tested with OBS 32.1 or ffmpeg sending simulcast, if possible.

## Notes

- Related: [Streams: temporal layers for slow viewers](stream-temporal-layers.md) aims at the same problem for browser streams; for WHIP, simulcast lets OBS do the work.
- OBS guide: https://obsproject.com/kb/whip-streaming-guide
- Worth it when viewers watch on phones or weak connections; with a few viewers on good connections one stream is enough.
