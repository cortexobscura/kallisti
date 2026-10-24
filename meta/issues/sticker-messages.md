# Sticker messages

Part of [Stickers](stickers.md).

## Summary

A message can be a sticker: no text, a sticker id. All clients see it in history; Mumble clients get a text.

## Requirements

- Column `messages.sticker_id`; a post with a sticker has no text and no upload; the sticker must exist and not be deleted (else refused).
- Sticker messages cannot be edited; replies to them and reactions on them work as for other messages; threads and direct messages too.
- History and live messages carry the sticker id; clients look up name and image in the sticker list (a deleted sticker: its name only).
- Mumble clients and the text of notifications and previews: "[sticker: name]".
- Message rules (slow mode, limits for new accounts) count a sticker as a message.

## Acceptance Criteria

- Tests: a sticker post is stored and sent with its id; text with a sticker, an unknown and a deleted sticker are refused; an edit is refused; Mumble gets the text; slow mode applies.
