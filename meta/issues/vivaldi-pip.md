# Vivaldi and ungoogled Chromium: the picture-in-picture button does nothing

## Summary

A report: in Vivaldi, the stream tile's picture-in-picture button does
not open the floating window. The console shows "Uncaught (in promise)
AbortError: The play() request was interrupted by a new load request."
(index bundle, line 20: the PiP code and the stream sound's helper
video.)

Our PiP plays a new, detached video element with the canvas's capture
stream, then asks for PiP; a failure is caught and the button stays off.
Both play() calls on that bundle line catch their errors, so the
uncaught one likely comes from a play() we do not call ourselves:
Vivaldi's own video features (its PiP button on videos) may wrap or
call play() on our video and reload it. Not checked: no Vivaldi here.

The reporter sees the same in ungoogled Chromium. PiP works in headless
Chromium 156 (Playwright, the real image: a camera stream and a shared
tab with sound) and in Helium. So likely something of the reporter's
in both browsers: an extension, or a setting.

Open: Vivaldi version, does Vivaldi's own PiP button show on the stream,
does it happen with Vivaldi's "Picture-in-Picture button on videos"
setting off, does PiP work for the same user in Chrome.

Closed: it works in a private window, so it is something of the
reporter's (an extension most likely), not Kallisti.

## Acceptance Criteria

- Cause found with the reporter's answers or a Vivaldi test.
- PiP opens in Vivaldi, or a clear message when the browser refuses it.
