# Notes ingest and endpoints

- What changed: ingest now captures non-task bullet entries as notes (status NULL, type "misc") while tasks keep status populated; tasks list queries now exclude NULL status rows; added GET /notes endpoint and SQL query for notes; tasks schema adds raw_line and type fields and makes status nullable.
- Why it changed: need to retain non-task journal entries and expose them separately from tasks.
- New conventions/decisions: notes are stored in the tasks table with status NULL and type "misc"; tasks list endpoints filter on status IS NOT NULL.
- Follow-ups / TODOs: ensure DB is rebuilt (drop/recreate tasks table) to pick up schema changes.
