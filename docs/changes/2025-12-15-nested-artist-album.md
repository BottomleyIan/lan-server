# 2025-12-15 nested artist/album data in responses

## What changed
- Added nested artist info to album responses and nested artist/album summaries to track responses (including playlist tracks) to make related metadata explicit objects.
- Extended SQL queries to join artists/albums, updated handler mappings to include summaries, and regenerated Swagger docs.

## Why it changed
- Requested clearer responses where artist/album details accompany IDs so consumers can show names without extra calls.

## New conventions / decisions
- Track DTOs include optional `artist` and `album` summaries; Album DTOs include an optional `artist` summary.
- Playlist track payloads now surface the nested track summaries with artist/album data when available.

## Follow-ups / TODOs
- Consider limiting fields returned in summaries if payload size becomes an issue.
