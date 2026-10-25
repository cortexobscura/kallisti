# Several uploads per message: store and protocol

Parent: [Several files in one message](multiple-attachments.md)

## Summary

The store keeps a list of uploads per message, in order (today:
`messages.upload_id`). `send` takes a list of upload keys (at most 10,
each unused and the sender's); messages carry a list of uploads.

## Acceptance Criteria

- Old messages keep their file (migration of `upload_id`).
- `send` with 0 to 10 keys; more, a used key or another user's key is
  refused, and no upload is marked used then.
- Deleting a message (also purges) deletes all its files.
- Clients from before the change still show the first file (`upload`
  stays next to the new list, while old clients are around).
- History, search, pins, replies and threads give all files.
