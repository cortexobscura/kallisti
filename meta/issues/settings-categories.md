# Settings in categories

## Summary

The settings dialog is one long page. The user chose layout A of the
mockups (ops/mockups/settings.html, 2026-11-01): categories with a side
menu on desktop; on phones a list of categories, each its own page with
a back arrow.

## Requirements

- Categories: User settings (Profile, Account), App settings (Voice &
  video, Notifications, Appearance). Empty ones do not show (no account:
  no Account).
- Desktop: a larger dialog, the menu on the left, one category on the
  right; Log out at the bottom of the menu (with an account).
- Phones: the list (with "Edit profile ›" on top), a category page with
  "‹ Settings"; Escape or the back arrow returns to the list.
- Opened for the account (the first password after a fediverse
  confirmation; a guest's "keep this account"): on Account.
- Everything that is there today stays (no setting lost).

## Acceptance Criteria

- Vitest for the pure parts (which categories, where it starts).
- Playwright: desktop and phone sizes, each category opens, back works,
  the account start.
