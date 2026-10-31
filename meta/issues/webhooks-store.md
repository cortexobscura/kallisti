# Webhooks: store and management

## Requirements

- Table webhooks: id, channel, name, token_hash (SHA-256 of a random
  token, as invite links), creator, created_at. Deleted with its channel.
- Manage requests for staff (MayManage): list a channel's webhooks,
  create (name, at most 80 characters; at most 10 per channel), rename,
  delete. The token goes once to its creator (the URL cannot be shown
  again: a new webhook if it is lost).
- messages.webhook_id: the webhook a message came from (its name may be
  another per post); on the wire as `webhook: <id>` (for its picture).

## Acceptance Criteria

- Go tests: store round trip, permissions, limits, token only at
  creation, channel delete.
