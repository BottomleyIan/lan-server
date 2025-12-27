# Journal list endpoint

- What changed: added GET /journal with optional query filters for year, month, day, and tag; added a filtered journals query and regenerated sqlc and Swagger docs.
- Why it changed: provide a single endpoint to query journals by tag/date without path parameters.
- New conventions/decisions: /journal uses query parameters and returns JournalDTO arrays.
- Follow-ups / TODOs: none.
