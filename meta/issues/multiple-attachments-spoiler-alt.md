# Files: spoiler and description

Parent: [Several files in one message](multiple-attachments.md)

## Summary

As in Discord: in the composer each file can be marked as a spoiler and
get a description (alt text, up to 1024 characters). A spoiler image is
blurred with a "SPOILER" label until a click; the description is the
image's alt text and shows in the viewer.

## Acceptance Criteria

- Spoiler and description are kept with the file and sent to all clients.
- A spoiler shows blurred until clicked, per viewer, also in the grid.
- Screen readers read the description.
