# Polls: the web client

Parent: [Polls](polls.md)

## Summary

"Create poll" in the composer's "+" menu opens a form (question, answers
with emoji, multiple choice, anonymous, duration). The poll shows in the
chat: answers to pick, a Vote button, then bars with counts (after one's
own vote or the end), who voted (named polls), the time left, "End poll"
for the author.

## Acceptance Criteria

- Form validation as on the server; the poll is sent like a message.
- Voting, changing and taking back a vote; live updates.
- Says before voting whether votes are anonymous.
- Checked in a browser with two accounts.
