# Composer: pick, paste and drop several files

Parent: [Several files in one message](multiple-attachments.md)

## Summary

The file picker takes several files; paste and drop add files. The
composer shows the waiting files in a row (image previews, file cards),
each with a remove button. Up to 10; the 11th is refused with a note.

## Acceptance Criteria

- Picking, pasting and dropping several files works; remove one of them.
- Each file uploads with its own progress and retry; send waits for all.
- Waiting files are kept per chat across chat changes, reconnects and
  reloads, as the one file is today (`lib/pendingFiles`).
- Files over the size limit are refused one by one, the rest stays.
