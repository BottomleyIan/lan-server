# Add tasks tables and transitions

- What changed
  - Added task tables for statuses, tasks, and task transitions in a new migration.
  - Seeded the initial task status set.
- Why it changed
  - Provide a reusable task model and history tracking as the server expands beyond music.
- New conventions/decisions
  - Task statuses are stored in a lookup table and referenced by code (text).
  - Task history is recorded in `task_transitions` with `status_code`, `reason`, and `changed_at`.
- Follow-ups / TODOs
  - Add sqlc queries and API handlers for tasks and transitions.
