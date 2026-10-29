# Rethink the split of text and voice channels

Raised by the user (2026-10-29), after the requests for voice links and
private categories.

## Summary

Text and voice channels can do different things, and people do not
expect that: categories, notification settings, mute and links are for
text channels; subchannels and recording for voice channels; a voice
channel's chat is not a full text channel (no links to it, for example).
The cause is the model: voice channels are the Mumble tree, text
channels are Kallisti's own (with Mumble mirrors).

To work out: one kind of channel with text and voice (as Discord's text
in voice channels, or the other way round), or the same features for
both; what Mumble clients see; migration of existing servers.

## Acceptance Criteria

- A list of today's differences, and a design with its migration,
  agreed before any code.
