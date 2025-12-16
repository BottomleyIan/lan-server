# 2025-12-16 startswith filters for lists

## What changed
- Added `startswith` query param to list endpoints for artists, albums, and tracks (including album tracks) to do prefix filtering.
- Track list endpoints honor `expand` as before and now accept a prefix filter on filename; shared logic handles joins vs base queries.
- Added prefix-friendly indexes via migration `003_prefix_indexes.sql` on artist name, album title, and track filename.
- Swagger updated for the new query params.

## Why it changed
- Needed fast “type-ahead” style filtering on list endpoints without full-text search.

## New conventions / decisions
- Prefix filtering uses `startswith` and leverages simple `LIKE 'prefix%'` queries with supporting indexes.
- Invalid `expand` values still return 400; `startswith` is optional.

## Follow-ups / TODOs
- Consider adding broader search or case-insensitive normalization later if needed.
