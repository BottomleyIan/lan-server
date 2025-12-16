# 2025-12-16 track expand defaults

## What changed
- Default `expand` behavior for `GET /tracks` and `GET /albums/{id}/tracks` now omits nested album/artist data unless explicitly requested.
- Swagger parameter descriptions now note the default, and docs were regenerated.

## Why it changed
- The endpoints were returning album/artist objects even without an `expand` query, which was unexpected for callers wanting explicit opt-in.

## New conventions / decisions
- `expand=album` and/or `expand=artist` must be provided to include those nested objects in track list responses; invalid values still return 400s.

## Follow-ups / TODOs
- Make sure any clients that relied on implicit album/artist payloads are adjusted.
- Add automated tests around `expand` defaults to prevent regressions.
