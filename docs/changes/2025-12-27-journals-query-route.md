# Journals query route

- What changed: moved the query-based journals list endpoint to GET /journals (from /journal) to match existing routing conventions; updated Swagger docs.
- Why it changed: align the route name with the other journals endpoints.
- New conventions/decisions: /journals now supports query filtering via year, month, day, and tag.
- Follow-ups / TODOs: none.
