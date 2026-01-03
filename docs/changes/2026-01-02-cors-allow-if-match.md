# CORS allow If-Match

## What changed
- Added `If-Match` and `If-None-Match` to CORS allowed headers and exposed `ETag`.

## Why it changed
- Fix browser preflight failures when updating journal entries with `If-Match`.

## New conventions/decisions
- None.

## Follow-ups / TODOs
- None.
