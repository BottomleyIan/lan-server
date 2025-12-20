# Add track rating endpoint

- What changed
  - Added a PATCH endpoint to update a track's rating.
  - Enabled CORS for PATCH requests.
- Why it changed
  - Provide a dedicated endpoint for updating ratings without other track metadata.
- New conventions/decisions
  - Rating updates use `PATCH /tracks/{id}/rating` with a 1-5 value (or null to clear).
- Follow-ups / TODOs
  - Regenerate Swagger docs when you want the API docs updated.
