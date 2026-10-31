# Webhooks: web client

## Requirements

- Webhook messages: an "APP" tag after the name; no user menu for the
  name; the webhook's picture.
- Channel settings (staff): Webhooks: the list (name, created by, when),
  create (shows the URL once, with Copy and a note), rename, delete.

## Acceptance Criteria

- [x] Vitest for the parsing and pure parts; Playwright: create, curl a
  post, it shows with the tag, delete.
