# Quick reactions: two rows, most used and recent

Requested by kaia (2026-10-29): the 8 quick reactions rarely hold the one
she wants.

## Summary

Two rows instead of one. Row 1: the 8 most used emoji (a count per emoji
that fades with later uses, so the row stays put), then "+" (and "×" on
wide screens). Row 2: the most recent ones not in row 1, 10 on wide
screens, 9 on phones (the buttons line up with row 1 there). Gaps from the
default list. Counts stay on the device, like the recent list.

## Acceptance Criteria

- Unit tests: counting (fading, a size limit), the two rows (most used,
  then recent; defaults fill; no emoji twice; custom emoji that are gone
  left out).
- Browser check on desktop and phone widths.

## Result (2026-10-29)

Done. Row 1: the 8 most used (counts fade by 0.9 per later use, 50 kept), then "+" (and "×" on wide screens); row 2: recent ones not in row 1, 10 wide, 9 on phones. 20 defaults fill gaps before there is a history. Without counts yet (from before), row 1 takes the recent ones. Checked in a browser at desktop and phone widths.
