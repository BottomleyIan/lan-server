# 2025-12-16 track list expand support

## What changed
- Added `expand` (album, artist) to `GET /tracks` and `GET /albums/{id}/tracks`, controlling whether nested album/artist objects are populated.
- Queries now skip joins when expansions aren’t requested, keeping default behavior intact; Swagger updated for the new param.

## Why it changed
- Needed a way to opt into/omit nested album/artist data without adding field-level filtering.

## New conventions / decisions
- Default responses still include album and artist; `expand` can restrict to only those requested (invalid values return 400).
- Album-scoped track lists reuse the same shared logic and expansion rules as the main tracks list.

## Follow-ups / TODOs
- Consider adding pagination/search filters later when tracks endpoint becomes a search endpoint.
