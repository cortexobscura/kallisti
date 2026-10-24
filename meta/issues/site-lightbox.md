# Site: enlarge screenshots

## Summary

The screenshots on the site are the right size for the page, but small to read. A click opens one at full size.

## Requirements

- Screenshots in feature sections and the hero (index.html and features.html) open in a lightbox on click, Enter or Space: the same file (they are 2x), as large as the window allows, with its alt text as caption.
- Closes with a click anywhere, Escape, or a close button; focus goes back to the screenshot.
- Plain JavaScript and a native <dialog>, no library; works without JavaScript as before (the pictures just do not enlarge).

## Acceptance Criteria

- Checked in a headless browser on both pages: open by click and by keyboard, close by click, Escape and the button; the page does not scroll behind it.
- Pushed only after the user's OK (the push publishes the site).
