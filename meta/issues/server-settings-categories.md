# Server settings in categories

## Summary

The server settings dialog has ten tabs in a row. As the user settings
(issue settings-categories): a side menu on desktop, a list with pages
on phones. Both dialogs share one frame component.

## Requirements

- Groups: Server (Overview: name and logo; Emoji; Stickers; Threads;
  Message rules), Members (Registered users; Bans; Guests), Monitoring
  (Connections; Recordings). Each category only for who may use it, as
  the tabs now.
- The frame (header, menu, phone pages, back and Escape, focus, the
  full-screen phone dialog) is one component; the user settings use it
  too, unchanged for the user.
- Everything that is there today stays.

## Acceptance Criteria

- Vitest for the shared pure parts.
- Playwright: both dialogs at desktop and phone sizes; each server
  category opens; the user settings checks still pass.
