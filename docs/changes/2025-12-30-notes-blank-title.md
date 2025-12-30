# Notes allow blank titles

- What changed: ingestion now keeps misc entries even when the parsed title is blank; only task entries still require a title.
- Why it changed: notes should be returned even when their heading is blank (e.g., tag-only entries).
- New conventions/decisions: blank titles are allowed for notes but not for tasks.
- Follow-ups / TODOs: none.
