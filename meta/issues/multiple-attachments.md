# Several files in one message

## Summary

A message has at most one file now (`Message.Upload`, one waiting file
per chat in the composer). Discord allows up to 10 files per message,
images too, and shows several images as a grid with a viewer.

Do the same: up to 10 files per message, each file with the size limit
of today (the daily quota stays as it is).

## Sub-issues

- [Several uploads per message: store and protocol](multiple-attachments-server.md)
- [Composer: pick, paste and drop several files](multiple-attachments-composer.md)
- [Image grid and a viewer with next and previous](multiple-attachments-grid.md)
- [Mumble bridge: several images in one message](multiple-attachments-mumble.md)
- [Edit: remove files from a message](multiple-attachments-edit.md)
- [Files: spoiler and description](multiple-attachments-spoiler-alt.md)

## Acceptance Criteria

- All sub-issues done.
- A message with 10 files sends, shows, edits, replies, pins and
  deletes like one with a file today.
