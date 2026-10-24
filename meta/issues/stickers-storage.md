# Stickers: storage and upload

Part of [Stickers](stickers.md).

## Summary

The server keeps a server's stickers, lets staff add, rename and delete them, and tells clients about them.

## Requirements

- Table `stickers`: id (AUTOINCREMENT), name (emoji name rule, unique among stickers that are not deleted), description (optional, up to 100 characters), emoji (one related emoji, optional), pack, type, data (BLOB; NULL once deleted), author, created_at.
- A pack name (optional; the picker groups by it). At most 200 stickers (not deleted); 512 KB; PNG, GIF, WebP or JPEG (detected from the content, as emoji); the pixel check of emoji.
- Upload as emoji: a WebSocket request gives a single-use token (name, description, emoji checked first), then POST /api/stickers with the token; GET /api/stickers/{id} serves the image (immutable, cached for a year; 404 once deleted).
- WebSocket requests: rename (name, description, emoji), delete (keeps the row and name, removes the image).
- Who may manage emoji may manage stickers (mayManageEmoji); nobody without the built-in server.
- `ready` carries the stickers (id, name, description, emoji, url); a change is sent to everyone.

## Acceptance Criteria

- Store tests: add, the limit, a taken name, rename, delete keeps the name and frees it, ids not reused.
- Upload tests: size and type checks, the token, serving, 404 after delete.
- Sidecar tests: permissions, ready, the broadcast after add, rename and delete.
