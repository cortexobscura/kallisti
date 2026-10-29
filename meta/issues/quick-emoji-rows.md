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

Done. Row 1: habits, emoji used more than once lately (counts fade by 0.98 per later use, 50 kept; at least 1.5), most first, filled from 20 defaults so the row stays put; then "+" (and "×" on wide screens). Row 2: recent ones not in row 1, 10 wide, 9 on phones. Without counts yet (from before), row 1 takes the recent ones. Status emoji are recent, not counted. (A first fade of 0.9 made row 1 a recent list: found in review.) Checked in a browser at desktop and phone widths.
