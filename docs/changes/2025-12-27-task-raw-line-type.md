# Task raw line and type fields

- What changed: added raw_line and type columns to the tasks table schema; updated task ingestion to capture the full first-line text and set type to "task"; regenerated sqlc code.
- Why it changed: need to persist the original task line from the journal file and store a task type.
- New conventions/decisions: tasks.type currently set to "task" for all ingested Logseq tasks.
- Follow-ups / TODOs: none.
