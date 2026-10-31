# Webhooks

## Summary

A secret URL that posts messages into a text channel, for CI, RSS,
monitoring and the like. Discord-compatible, so tools with a "Discord
webhook" option work with the URL alone. User decisions (2026-10-31):
admins and moderators manage them; Discord format (Slack's `{text}` too);
a post may set the name, not the picture (the webhook's own picture: a
picture URL would make every reader load a foreign image); text channels
only.

## Requirements

- Sub-issues: store and management, the endpoint, the web client, the
  picture.

## Acceptance Criteria

- [x] The sub-issues are done; curl with Discord's JSON posts a message that
  readers and Mumble users see.
