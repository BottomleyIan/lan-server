# Journal entry title request field

## What changed
- Renamed journal entry create/update request field from `description` to `title`.
- Updated handlers and regenerated Swagger docs.

## Why it changed
- Align request payload naming with the response field (`title`).

## New conventions/decisions
- Requests now use `title` for journal entry text.

## Follow-ups / TODOs
- None.
