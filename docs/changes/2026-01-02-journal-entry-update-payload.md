# Journal entry update payload

## What changed
- Updated journal entry PUT payload to match the POST payload fields (description/tags/body/status/scheduled/deadline).
- Removed the requirement to send raw Logseq blocks in update requests.
- Regenerated Swagger docs.

## Why it changed
- Make entry updates simpler and consistent with create requests.

## New conventions/decisions
- Update requests now render entries from structured fields instead of raw text.

## Follow-ups / TODOs
- None.
