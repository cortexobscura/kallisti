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
