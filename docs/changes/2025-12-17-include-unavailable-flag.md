# Optional include_unavailable for albums and tracks

## What changed
- Album list endpoint now accepts `include_unavailable` (default false) and filters albums to those with tracks in available folders unless explicitly included.
- Track list (`/tracks`) and album tracks (`/albums/{id}/tracks`) accept `include_unavailable` to include tracks from unavailable roots; defaults remain available-only.
- Track queries now take an `include_unavailable` flag instead of hardcoded availability filters; SQL and handlers updated accordingly. Swagger/sqlc regenerated.

## Why it changed
- Needed a way to include items from unmounted/unavailable volumes for admin/debug scenarios while keeping default responses scoped to available folders.

## Follow-ups / TODOs
- Consider surfacing availability metadata on album/track DTOs to make UI filtering clearer.
