# Case-insensitive tag filters

- What changed: tag filters in notes/tasks/journals queries now compare case-insensitively.
- Why it changed: tag-only results were missing entries due to mixed-case tags.
- New conventions/decisions: tag matching is case-insensitive at query time.
- Follow-ups / TODOs: consider backfilling existing tags to lowercase for consistency.
