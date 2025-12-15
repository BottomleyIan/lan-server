# Schema updates for artist/album metadata

## What changed
- Added `artists` and `albums` tables with soft-delete timestamps, unique constraints, and `updated_at` triggers.
- Extended `tracks` with `artist_id`, `album_id`, `genre`, and `year` columns plus foreign keys and supporting indexes.
- Expanded `sqlc` overrides and `dbtypes` with a `NullInt64` alias so new nullable columns generate concrete types instead of `interface{}`.
- Added `.gitignore` rules for SQLite artifacts and removed the checked-in `data.sqlite` files from version control.
- Added GET/PUT/DELETE handlers for artists and albums (DTO-based), refreshed track DTO/swagger, and regenerated Swagger docs.
- Added list endpoints for artists and albums, plus corresponding sqlc queries and Swagger updates.
- Added `rating` (1–5) to tracks, track DTO/mappers, track GET/PUT endpoints to update rating, and Swagger regeneration.
- Added artist/album upsert queries and scanner integration to upsert metadata from tags (artist/album/genre/year) onto tracks after file scans.
- Added track playback/download endpoints that stream files from disk using folder+rel_path, with Swagger docs regenerated.
- Updated playback/download Swagger responses to mark binary content so Swagger “Try it” works for streaming files.
- Added project-level `TODO.md` and linked it from `README.md` for quick task tracking.
- Updated TODOs: added high-priority cover art handling/thumbnail task; moved embeddings note into low-priority checkbox; removed embeddings note from README.
- Added low-priority TODO to serve curl-friendly terminal responses when detecting curl User-Agent.
- Added playlists support: new migration/tables (`playlists`, `playlist_tracks`), sqlc queries, DTO/mappers, routes for list/get/post/put, playlist-track management (list/add/update position), and Swagger regeneration. Delete existing SQLite DB before restart to apply create-only migrations.
- Added cover image plumbing: `image_path` columns for tracks/albums (moved into base migration), sqlc overrides/models/DTOs, and scanner now saves embedded cover art to `tmp/covers/{album_id}/{track_id}.*` and sets track/album image paths. Delete SQLite to pick up the migration.

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
