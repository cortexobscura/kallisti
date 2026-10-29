# See who has access to a private channel from its menu

Requested by MOONMAN (2026-10-29).

## Summary

"Members" in a private channel's menu, for everyone who can see the
channel (members, channel moderators, staff): a read-only list of its
members and the staff, without entering it (today the list shows only
while one is in the channel; staff and channel moderators get the
editable Members dialog). The server sends the list to these users
already (sidecar privatechannels.go): web only.

## Acceptance Criteria

- Menu item for members who cannot edit; the editable dialog stays for
  staff and channel moderators.
- Unit test for who gets which item; browser check as a plain member.

## Result (2026-10-29)

Done: "Members…" (read-only) in the menu of text and voice channels for members who do not manage them. Checked in a browser on a local server. Not on the phone channel sheet (it has no Members item at all yet).
