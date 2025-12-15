# TODO

## High
- [ ] Fix picture handling in metadata: extract/normalize covers, save thumbnails, and store references from tracks to cached images (or decide final storage approach).
- [ ] Add image endpoints for tracks/albums: serve track image (fallback to album image) and album image (fallback to first track image).

## Medium

## Low
- [ ] Revisit embeddings later (consider FTS5 first; hang any embeddings off stable `track_id`).
- [ ] Add curl-friendly responses (detect curl User-Agent and render concise text output for terminal users).
