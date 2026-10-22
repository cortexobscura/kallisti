# Login: no server password field when the server has none

## Summary

The login page always shows "Server password (only if the server has one)". A Kallisti server with the built-in Mumble server knows if it has a password (MUMBLE_PASSWORD), so the field can go when there is none.

## Requirements

- The server says if guests may need a server password: GET /api/server gives `{"password": bool}`. Built-in server: true only with MUMBLE_PASSWORD. Murmur: true (unknown).
- The login page shows the field only when the answer is true, or when there is no answer (older servers).

## Acceptance Criteria

- Server test for both answers.
- Client test: field hidden for `password: false`, shown for `true` and on errors.
