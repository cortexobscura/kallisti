# The whole page scrolls away

## Summary

A report (Chromium-based Helium on Linux, 1347×627): after scrolling the
messages or the member list, the whole app moves up. The channel header
is gone at the top, an empty band shows at the bottom, and the page has
its own scrollbar. The app is `height: 100%` on html, body and #app, so
the document should never scroll: something reaches below the viewport
and makes it scrollable, and a list at its end hands the wheel on to it
(scroll chaining).

## Acceptance Criteria

- The overflowing element found and a test for it.
- The document does not scroll in the app (desktop and phone layout);
  pages that need to scroll (login on a small screen) still do.
