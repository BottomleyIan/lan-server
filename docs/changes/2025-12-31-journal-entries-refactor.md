# Journal entries refactor

## What changed
- Replaced tasks/notes handlers, DTOs, and routes with unified `/journals/entries` CRUD endpoints.
- Added a `JournalEntryDTO` with nullable task fields (`status`, `scheduled_at`, `deadline_at`) and updated calendar responses to return entries.
- Consolidated SQL queries for journal entries and updated sync logic, plus regenerated Swagger docs.

## Why it changed
- Notes and tasks now share a single entries surface, reducing duplicated code and aligning with unified journal entry storage.

## New conventions/decisions
- Use `status` presence to indicate tasks; `null` status and timestamps indicate non-task entries.
- `type` query param accepts `task`, `misc`, or `note` (alias for `misc`) on `/journals/entries`.

## Follow-ups / TODOs
- None.
