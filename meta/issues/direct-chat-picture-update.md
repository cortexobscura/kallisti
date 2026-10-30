# Direct chats: a new profile picture does not show in the list

Reported by kaia (2026-10-30): after a direct chat started with someone
without a profile picture, their new picture does not show in the
sidebar's direct chat list.

## Acceptance Criteria

- Cause found and a test for it; the list shows a picture set (or
  removed) later, without a reload.

## Result (2026-10-30)

Cause: the direct chat list (and the quick switcher's direct chats) asked for the picture without "registered", so only an online partner's picture showed; offline ones fell back to initials, also after a reload. Partners are always registered: now they get their saved picture (no-cache URL, so a new one shows). Reproduced and checked in a browser on a local server (picture set while online, partner then offline, and after a reload).
