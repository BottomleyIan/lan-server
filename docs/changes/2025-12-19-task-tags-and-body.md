# Add task tags and document body usage

- What changed
  - Added a `tags` column to tasks in the migration and threaded it through sqlc queries.
  - Updated task DTOs, mappers, and handlers to accept and return `tags`.
- Why it changed
  - Tasks need tag metadata, and the body field is intended to store markdown content.
- New conventions/decisions
  - Task tags are stored as a JSON array in the `tasks.tags` text column.
- Follow-ups / TODOs
  - Regenerate Swagger docs when you want the API docs updated.
