# Polls: custom emoji, no pings, emoji picker

Parent: [Polls](polls.md)

## Summary

The user's decisions (2026-10-26) after the first version:

- Answers may have the server's custom emoji (and one's personal
  emoji, as reactions do), not only Unicode emoji.
- Polls never ping: no mentions, no @everyone/@here, no mention
  highlight or notification from the poll's text.
- The form picks an answer's emoji with the emoji picker instead of a
  free text field (so it is always valid).

## Acceptance Criteria

- The server takes custom and personal emoji in answers, as for reactions.
- Mentions in a poll ping nobody (server push and follows, client).
- The form's emoji comes from the picker; the poll shows custom emoji.
