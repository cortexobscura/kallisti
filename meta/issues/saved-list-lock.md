# One tab at a time writes the saved server list

Parent: [Status of the branch multi-server](multi-server-status.md)

## Summary

Two tabs that change the saved list (and its tokens) in the same instant
can lose one change: read, change and write are separate steps. Wrap them
in a lock across tabs (Web Locks, `navigator.locks.request`), read again
inside the lock. Without Web Locks (old browsers), keep today's behaviour.

## Acceptance Criteria

- Test: two writes that overlap both end in the list.
