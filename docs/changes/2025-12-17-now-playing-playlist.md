# Ensure default Now Playing playlist and playlist maintenance endpoints

## What changed
- Added migration to seed a default `Now Playing` playlist at ID 1, with safeguards for existing duplicates.
- Introduced playlist delete endpoint with a guard against deleting the reserved playlist (IDs <= 0 or 1 return unauthorized).
- Added endpoints to clear all tracks from a playlist and to delete a single track from a playlist.
- Added enqueue endpoint to append a track to the end of a playlist using the next position.
- Updated routing, swagger comments, and sqlc-generated code to support the new playlist operations.

## Why it changed
- Product requirement to always have a `Now Playing` playlist and provide playlist maintenance APIs.

## New conventions/decisions
- Playlist ID 1 (`Now Playing`) is reserved and cannot be deleted.
- Playlist clear and per-track delete operations soft-delete playlist_tracks rows instead of hard deletes.

## Follow-ups / TODOs
- Consider adding authorization/ownership controls around playlist mutation endpoints if multi-user support is needed.
