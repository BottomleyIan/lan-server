# Journal entries table rename

- What changed: renamed the underlying tasks table to journal_entries and updated queries, sqlc overrides, and mappers to use the new table/model name.
- Why it changed: make the storage name reflect that it contains both tasks and notes.
- New conventions/decisions: tasks and notes continue to share the journal_entries table; API names remain /tasks and /notes.
- Follow-ups / TODOs: drop/recreate the tasks/journal_entries table so the renamed schema is applied.
