# Mumble bridge: several images in one message

Parent: [Several files in one message](multiple-attachments.md)

## Summary

A Mumble text with several images becomes one message with all of them
(today: the first on the text, the others in messages of their own). A
web message with several files goes to Mumble with all of them, within
Mumble's message size limit.

## Acceptance Criteria

- Mumble to web: one message, all images, in order.
- Web to Mumble: all images (or links where they do not fit), one text.
