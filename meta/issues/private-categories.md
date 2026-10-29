# Private categories: members for all channels in them

Requested by shibao (2026-10-29).

## Summary

A category can be private, with members like a private channel: its
members see and use all text channels in it, also ones made later. The
category's Mumble channel already holds its channels' mirrors, and Mumble
groups and ACLs inherit to them, so the members group goes on the
category's Mumble channel (server SetPrivate takes it already; the UI does
not offer it).

Decided (2026-10-29):

- A private channel inside a private category: its own members and the
  category's (the channel's group inherits the category's).
- Voice: no change. Categories hold text channels; a private voice
  channel already passes access to its subchannels.

## Acceptance Criteria

- Server tests: a member of the category reads and writes in all its
  channels (also a new one), non-members see none; @everyone, push,
  search, unread and readers follow it; moving a channel in or out
  changes its access.
- Web: Private switch and Members dialog for categories; lock icons on
  the category and its channels; the readers list.

## Result (2026-10-29)

Done. A private category closes the channels in it, not itself (so a private channel's own members still reach it; Mumble apps see the empty category channel). Only channels directly in a category inherit its members (a members group SuperUser made higher up gives no access). A public channel may not leave a private category, and a private category with channels may not be deleted (the history would become public). Channel moderators of a channel in it keep it. Category invites: a member, its first channel opens. Reviews: Claude and Astra. Checked in a browser on a local server (category member, channel-only member, outsider, invite).
