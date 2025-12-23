# Task timestamp output formatting

- What changed
  - Converted task `scheduled_at` and `deadline_at` outputs to ISO formats (`YYYY-MM-DD` or RFC3339) instead of Logseq strings.
- Why it changed
  - API consumers should receive consistent ISO-style timestamps rather than Logseq display formats.
- New conventions/decisions
  - Logseq timestamps with times are converted to RFC3339 in UTC; date-only values stay `YYYY-MM-DD`.
- Follow-ups / TODOs
  - Decide if time zone handling should preserve local time offsets instead of UTC.
