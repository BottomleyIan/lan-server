# Journals Swagger operation IDs

- What changed
  - Added explicit Swagger operation IDs for journals endpoints to avoid duplicate names in clients.
- Why it changed
  - swagger-typescript-api reported a method name collision for the journals routes.
- New conventions/decisions
  - Journals endpoints use explicit `@ID` annotations for stable client generation.
- Follow-ups / TODOs
  - None.
