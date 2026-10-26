# Docker: run as a chosen user and group id

## Summary

A user: "i think i need a way to set the user and group id to something
specific". Likely a bind mount for `/data` owned by a host user (as with
PUID/PGID in linuxserver.io images): the image runs as `kallisti`, made
by `adduser` without a fixed id, so its files on the host get whatever
id that is. The Helm chart already assumes 1000:1000.

The user's feedback: Compose's `user:` works. Make it easy anyway:
`PUID` and `PGID` in `deploy/.env` (default 1000:1000), and a note in
the Docker install docs.

## Acceptance Criteria

- The image's user has a fixed id (1000:1000, as the chart says).
- Running with another id works and is documented (compose `user:` or
  PUID/PGID), tested with a bind mount.
