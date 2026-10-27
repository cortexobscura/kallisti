# Message HTML: forms

## Summary

Found in a security review (2026-10-27), older than the multi-server work:
Mumble message HTML may keep `<form action="https://...">` and its
controls (DOMPurify keeps them; `formaction` goes). Nothing loads by
itself, but a message can show a fake login form that sends to another
site when submitted.

## Acceptance Criteria

- Message HTML without forms and form controls; a test.
