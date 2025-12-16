# 2025-12-16 track title support

## What changed
- Added a `title` column to tracks (defaulting to filename on scan) and wired it through sqlc models/queries.
- Scanner now persists tag titles when present, with a filename fallback to keep responses non-empty.
- Track API DTOs expose `title`; Swagger regenerated to reflect the new field.

## Why it changed
- Track listings lacked a dedicated title field even though audio metadata provides it, making responses less useful.

## New conventions / decisions
- Track titles are required (defaulted from filename) and updated from tags when available.

## Follow-ups / TODOs
- Consider a future migration to backfill existing libraries after dropping/recreating the DB as planned.
