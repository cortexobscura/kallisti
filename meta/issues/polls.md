# Polls

## Summary

Polls in text channels, threads and direct messages, as in Discord: a
question with answers, people vote, the result shows when it ends.
Design, not started.

## Requirements

- Create: a question (up to 300 characters), 2 to 10 answers (up to 55
  characters, each with an optional emoji), single or multiple choice, a
  duration (1 hour to 2 weeks). From the composer's "+" menu.
- A poll is a message (like a sticker: its own data next to the message,
  and a text for paths that only know text). Not editable; deletable as
  any message (own, moderators).
- Vote, change or take back a vote while it runs; counts show after one's
  own vote or when the poll ends. The author can end it early.
- The server ends polls on time (also after a restart) and posts the
  result as a short message.
- Votes update live for everyone in the channel.
- Mumble clients see the poll as text (question and answers, "vote in
  Kallisti") and the result message; they cannot vote.
- Search, replies, quotes, pins and push notifications use the text.

## Acceptance Criteria

- Server tests: create, vote rules (single/multiple, change, take back,
  closed poll), ending on time and after a restart, result message,
  permissions (read-only guests, private channels, blocked users).
- Web tests for the poll form and the display; checked in a browser with
  two accounts.

## Notes

- Open questions for the user: anonymous votes or names shown (Discord
  shows names); may guests vote; a limit of open polls per channel.
- Built-in server only (as stickers).
