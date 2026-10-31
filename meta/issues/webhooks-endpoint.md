# Webhooks: the endpoint

## Requirements

- POST /api/webhooks/{id}/{token} (Discord's path): JSON or multipart
  with `payload_json` (files are ignored). Fields: content, username
  (at most 80), embeds (title, url, description, fields, footer, author:
  turned into Markdown text), Slack's text. Unknown fields ignored.
- 204, or the message with `?wait=true`; errors as Discord's
  `{"message","code"}` with 400/401/404/429 (Retry-After).
- No pings at all (as poll results: tools mention with Discord ids,
  which match no name); content at most 2000 characters (400 above, as
  Discord); per webhook 5 posts per 2 s (Discord's limit).
- Delivered as other messages (readers, push, Mumble: "Name: text").

## Acceptance Criteria

- [x] Go tests: formats, errors, limits, delivery, Mumble text.
