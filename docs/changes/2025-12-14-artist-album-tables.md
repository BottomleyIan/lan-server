# Schema updates for artist/album metadata

## What changed
- Added `artists` and `albums` tables with soft-delete timestamps, unique constraints, and `updated_at` triggers.
- Extended `tracks` with `artist_id`, `album_id`, `genre`, and `year` columns plus foreign keys and supporting indexes.
- Expanded `sqlc` overrides and `dbtypes` with a `NullInt64` alias so new nullable columns generate concrete types instead of `interface{}`.
- Added `.gitignore` rules for SQLite artifacts and removed the checked-in `data.sqlite` files from version control.
- Added GET/PUT/DELETE handlers for artists and albums (DTO-based), refreshed track DTO/swagger, and regenerated Swagger docs.
- Added list endpoints for artists and albums, plus corresponding sqlc queries and Swagger updates.

## Why it changed
- Normalize artist/album data to avoid duplication and support metadata-aware queries; capture genre/year tags alongside tracks.

## New conventions/decisions
- Artists are unique by name; albums are unique per artist/title.
- Track `artist_id`/`album_id` are nullable for files without tags; `genre`/`year` default to NULL.
- Nullable integer columns should use `dbtypes.NullInt64` via sqlc overrides to avoid interface{} in generated code.

## Follow-ups / TODOs
- Delete the existing SQLite file before the next run so the create-only migration applies the new schema.
- Update sqlc models/queries and scanner logic to populate artist/album/genre/year.
- Decide if/when to prune unused artist/album rows when no tracks reference them.
