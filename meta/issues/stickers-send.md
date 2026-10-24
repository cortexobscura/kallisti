# Stickers in the web app: send and show

Part of [Stickers](stickers.md).

## Summary

Everyone can send a sticker from the picker; messages show it large.

## Requirements

- The emoji button's picker gets a "Stickers" tab with the server's stickers (name as label, search by name, description and emoji); a click sends the sticker as its own message (the text in the box stays).
- A sticker message shows the picture at up to 160x160 (scaled down, never up), name as alt text and on hover; a deleted sticker shows its name.
- Replies, the reply preview, notifications, the chat preview text and the thread list show "[sticker: name]" or a small picture.
- On phones the picker tab works as on desktop.

## Acceptance Criteria

- Tests for the pure parts (sticker lookup, the preview text, the picker search).
- Checked in a headless browser: send a sticker, see it in two clients, after a reload, and after the sticker was deleted.
