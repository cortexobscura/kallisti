# Links that join a voice channel

Requested by MOONMAN (2026-10-29).

## Summary

A link to a voice channel (like the text channel links
`/channels/<id>`): whoever opens it sees the server, then a dialog
"Join <channel> on <server>?" (Join / Cancel), and is in the channel after
Join. "Copy link" in the channel menu for voice and text channels (on
desktop too: today only the phone sheet has it, for text).

- The link gives no access: a private channel's link works only for its
  members (others: "This channel does not exist or is private"). Access
  comes from invite links.
- Invite links that join voice today join at once: they get the same
  dialog.
- Another server's link opens that server first (as message links do);
  no account there: the add dialog first.
- Already in a call: the dialog says it leaves that call.

## Acceptance Criteria

- Unit tests: making and parsing the link; the routing (this server,
  another server, unknown channel, no access).
- The dialog before every join from a link (also invites).
- Browser check: link opened logged in, logged out (after the login),
  from another server, while in a call.
