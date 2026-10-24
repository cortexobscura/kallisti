# Stickers in the web app: manage

Part of [Stickers](stickers.md).

## Summary

Staff add, rename and delete stickers in the moderation dialog, next to the server's emoji.

## Requirements

- A "Stickers" tab (or section) in the moderation dialog, shown with the emoji rights: the list with each sticker's picture, name, description and emoji; add, rename, delete.
- Adding resizes in the browser with the emoji code, target 320x320 (the size becomes a parameter); animated files kept as they are if they fit 512 KB.
- Errors as for emoji (name taken, too large, the limit).

## Acceptance Criteria

- Tests for the pure parts (form checks, the resize sizes at 320).
- Checked in a headless browser: add, rename and delete a sticker.
