# Track list artist filter

- What changed
  - Added optional `artistId` query parameter to `GET /tracks` to filter by artist, matching the existing album filter behavior.
- Why it changed
  - Allow clients to fetch tracks for a specific artist directly.
- New conventions/decisions
  - `albumId` and `artistId` cannot be provided together.
- Follow-ups / TODOs
  - Consider adding combined filtering if needed.
