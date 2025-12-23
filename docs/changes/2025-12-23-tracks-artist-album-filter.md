# Track list album+artist filtering

- What changed
  - Allowed `albumId` and `artistId` to be provided together on `GET /tracks` and added combined filter queries.
  - Updated shared track listing helper to handle album+artist filtering.
- Why it changed
  - Support fetching tracks for a specific artist within an album.
- New conventions/decisions
  - If the album/artist combination has no matches, the endpoint returns an empty list.
- Follow-ups / TODOs
  - Consider adding combined filtering support elsewhere if needed.
