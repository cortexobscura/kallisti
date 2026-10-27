# Mentions of names with emoji

## Summary

Found in a security review (2026-10-27): ChatPane reads a mention's
`textContent` to find the user; Twemoji turns emoji in the text into
images first, so a mention of `🎮gamer` looks up `gamer`. Keep the
mentioned identity apart from the rendered text (a data attribute).

## Acceptance Criteria

- A mention of a name with emoji opens that user; a test.
