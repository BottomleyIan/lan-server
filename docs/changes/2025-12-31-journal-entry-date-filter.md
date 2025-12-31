# Journal entry date filter update

## What changed
- Added `journal_date` to `journal_entries` and populate it during journal sync.
- Updated journal entry list filtering to use a single date pattern across `journal_date`, `scheduled_at`, and `deadline_at`.
- Adjusted handlers to build the date filter pattern and updated sqlc output.

## Why it changed
- Align date filtering so year/month/day constraints match any of the three relevant dates as a single predicate.

## New conventions/decisions
- Date filtering now uses a `LIKE` pattern built from query params, matching journal, scheduled, or deadline dates.

## Follow-ups / TODOs
- None.
