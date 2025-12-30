# Tasks date filters include scheduled/deadline

- What changed: task date filters now also match scheduled_at and deadline_at date prefixes in addition to the journal date.
- Why it changed: date filtering should include tasks scheduled or due on a given date, not just the journal date.
- New conventions/decisions: year/month/day filters match either the journal date or the scheduled/deadline date.
- Follow-ups / TODOs: none.
