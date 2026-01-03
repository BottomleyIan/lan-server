# Journal entry position updates

## What changed
- Updated journal entry PUT to target entry position rather than hash, with raw body payload.
- Added a status update endpoint that changes task status by position.
- Removed If-Match requirement from entry updates.
- Regenerated Swagger docs.

## Why it changed
- Simplify updates for single-user workflows and avoid hash mismatch issues.

## New conventions/decisions
- Entry updates are addressed by `{position}` within the journal day.
- Status updates use `PUT /journals/entries/{year}/{month}/{day}/{position}/{status}`.

## Follow-ups / TODOs
- None.
