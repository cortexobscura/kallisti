# Polls: store and protocol

Parent: [Polls](polls.md)

## Summary

A poll is a message with its own data (`polls`, `poll_votes`); its text
(question and answers) serves Mumble, search, quotes and push. `send`
takes `poll`; `vote` and `end-poll` are new requests; a `poll` event
carries the counts to each reader with their own votes.

## Acceptance Criteria

- Create with the limits of the parent issue; guests cannot create or vote.
- Vote rules: single or multiple choice, change, take back, none after the end.
- Anonymous polls never send names; named polls send who voted for what.
- The author ends it early; polls cannot be edited; deleting removes the votes.
