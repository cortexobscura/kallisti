# Push notifications with the server's name and logo

Parent: [Admins change the server's name and logo](server-name-and-logo.md)

## Summary

Push notifications have "Kallisti" as their title and Kallisti's icon.
With a name or logo set by admins, they use those: the server sends
them with each push, the service worker shows them.

## Acceptance Criteria

- Title: the server's name (Kallisti's without one); icon: the logo's
  192 px icon (Kallisti's without one). The badge (a silhouette on
  Android) stays Kallisti's.
- The service worker takes only an icon from this server's logo path.
- Tests on both sides.
