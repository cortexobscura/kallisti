# A new cookie name for each login

Parent: [Status of the branch multi-server](multi-server-status.md)

## Summary

A logout answer clears the session cookie by its name. When the answer
comes late (tab A logs out, tab B logs in meanwhile), it deletes B's newer
cookie. Chatto's approach (ADR-017 there): each login sets a cookie with a
new name (`kallisti_session_<random>`, with `__Host-` over HTTPS); a logout
clears only the names its request carried; with several valid cookies the
newest session counts; a login clears the older names its request carried.
The old single name is still read and cleared, so nobody is logged out.
Also useful for main (older than the branch).

## Acceptance Criteria

- Test: a logout carrying only the old cookie does not clear a newer
  login's cookie.
- Test: with two valid cookies the newer session counts.
- The number of cookie names read is bounded.
