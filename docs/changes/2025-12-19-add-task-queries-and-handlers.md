# Add task queries and handlers

- What changed
  - Added sqlc queries and generated db code for tasks, statuses, and transitions.
  - Added task DTOs, mappers, and HTTP handlers (including transitions and status list).
  - Wired new task routes under `/api/tasks`.
- Why it changed
  - Provide an initial task API and persistence layer for the expanded server scope.
- New conventions/decisions
  - Status changes are recorded via `/tasks/{id}/transitions`, which also updates the task status.
- Follow-ups / TODOs
  - Regenerate Swagger docs after reviewing the new endpoints.
