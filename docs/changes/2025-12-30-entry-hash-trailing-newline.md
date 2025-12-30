# Entry hash uses raw content

- What changed: update/delete now parse journal files using the raw content (no trailing newline trim) so hash matching is consistent for the last entry.
- Why it changed: newly appended entries were failing hash lookups due to a trailing newline difference.
- New conventions/decisions: file parsing for update/delete uses the untrimmed content to match stored hashes.
- Follow-ups / TODOs: none.
