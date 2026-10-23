# WHIP streams: sound drops after a while

## Summary

Sound from an OBS/ffmpeg stream goes away and comes back in spurts. The sender stamps sound by sample count (the capture device's clock) and pictures by the system clock; they drift apart. The viewer's playout keeps both on one clock, and beyond its 1.5 s lead it drops the late sound (simulated: 0.1 % drift, all sound dropped after about 25 minutes).

## Requirements

- The server keeps the pictures' timestamps on the sound's timeline (pictures can move a few ms unseen, sound cannot).
- Streams without sound and browser streams stay as they are.

## Acceptance Criteria

- Unit test: with sound 0.1 % slow for an hour, picture and sound timestamps stay within 50 ms of each other relative to their arrival.
