# Stickers

## Summary

A sticker is a large picture sent as a message of its own (as in Discord):
about 160×160 px, on its own line, with no text in the same message. Staff
add server stickers, as they add custom emoji now. Kallisti already has
most of the parts: custom and personal emoji keep their images as blobs in
SQLite, staff upload them with a token (`/api/emoji`), and the picker
shows them. A sticker is a new kind of the same thing, with a larger size
and a message of its own.

## Requirements

- Storage: stickers next to custom emoji (same blobs), with a name (the
  emoji name rule), an optional description, and one related emoji (for
  suggestions). At most 512 KB, PNG, GIF, WebP or JPEG (APNG and animated
  WebP kept as they are). A limit per server (for example 60).
- Managing: who may manage emoji may add, rename and delete stickers (no
  new permission), in the same settings page as custom emoji.
- Sending: a message with only a sticker reference (for example
  `<sticker:name:id>`, the id of its image, so a later rename or delete
  does not break old messages). The server refuses a sticker message with
  other text. Mumble clients and old web clients see a text fallback
  (`:name:` or "[sticker: name]").
- Showing: about 160 px (scaled down, never up), on its own line; the
  name as alt text and on hover. A deleted sticker shows its name.
- Picker: a "Stickers" tab next to emoji. Typing a message whose text is
  only one emoji suggests stickers with that related emoji (optional).
- Resizing on upload (see [Resize custom emoji when adding them](emoji-resize.md)):
  the same code, target 320×320.
- Not: Discord's sticker packs, Lottie, stickers as reactions, stickers in
  other servers.

## Sub-issues

- [Stickers: storage and upload](stickers-storage.md)
- [Sticker messages](sticker-messages.md)
- [Stickers in the web app: manage](stickers-manage.md)
- [Stickers in the web app: send and show](stickers-send.md)

## Design (2026-10-24, from the code)

- Stored like personal emoji, not like server emoji: a deleted sticker keeps its row (image removed, name kept), so old messages show its name; ids are never reused, so images are cached forever.
- A sticker message is a message with a `sticker_id` column and no text (as uploads have `upload_id`), not a text code: markdown, word filters and the emoji rewriting do not touch it.
- Resizing in the browser as for emoji (target 320x320 instead of 96); animated files kept as they are.
- Managed by whoever may manage emoji (admins and moderators, built-in server only).

## Acceptance Criteria

- Staff add a sticker; another user sends it from the picker; all web
  clients show it at about 160 px, also after a reload and in history.
- A sticker message with text, or a reference to a sticker that does not
  exist, is refused (tests).
- A Mumble client sees the text fallback.
- A deleted sticker: old messages show its name, not a broken picture.
- Tests for the reference format (parse and write, as `emoji.Ref`), the
  size and type checks, and the limit per server.

## Notes

- The Discord comparison (`meta/research/discord-comparison.md`) had
  stickers as "skip". With custom emoji in place, they are a small
  feature: a data type, a message kind and a picker tab.
- Discord: 5 to 60 server stickers (by boost level), 512 KB, 320×320 px,
  PNG/APNG/GIF/Lottie, at most 3 per message (in practice one).
- If the wish is only "large animated pictures", jumbo custom emoji with
  a larger limit may be enough; ask before building.
