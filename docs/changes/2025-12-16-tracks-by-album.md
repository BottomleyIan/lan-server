# 2025-12-16 tracks by album endpoints

## What changed
- Added GET `/albums/{id}/tracks` and support for `albumId` query param on `/tracks` to return tracks for a specific album, including nested artist/album summaries.
- Introduced sqlc queries joining tracks with artist/album for album-scoped lists and updated mappers/routes/Swagger docs accordingly.

## Why it changed
- Needed album-focused track retrieval for album detail pages and optional filtered track browsing via query params.

## New conventions / decisions
- Album track endpoints use the playable (available roots only) dataset with nested artist/album summaries.

## Follow-ups / TODOs
- Consider pagination/filters (artistId, playlistId, search) for broader track browsing later.
