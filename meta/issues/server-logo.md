# Server logo

Parent: [Admins change the server's name and logo](server-name-and-logo.md)

## Summary

An admin uploads a logo (a square picture) or removes it. It shows in
the server rail in place of the initials, as the tab icon, on the login
page and as the installed app's icon (the manifest's icons). Without a
logo: the initials and Kallisti's icons, as now.

## Acceptance Criteria

- Admins only; size and format limits; resized on the server to the
  icon sizes.
- Served with a version in the URL, so a new logo shows without stale
  caches (the installed app's icon may still lag: the browser decides).
- Tests for upload, limits, removal and serving.
