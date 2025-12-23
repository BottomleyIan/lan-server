# Logseq tasks ingestion

- What changed
  - Replaced tasks schema to store Logseq-derived fields and removed task statuses/transitions tables.
  - Updated tasks SQL queries and handlers to list tasks only; task creation/editing endpoints were removed.
  - Added Logseq task parsing from journals ("- STATUS description" blocks with DEADLINE/SCHEDULED lines) during journal scans.
- Why it changed
  - Tasks are derived from Logseq journal markdown and should stay in sync with the filesystem source of truth.
- New conventions/decisions
  - Task records are rebuilt per journal file when the file changes; unchanged journals only refresh `last_checked_at`.
  - Supported statuses are: LATER, NOW, DONE, TODO, DOING, CANCELLED, IN-PROGRESS, WAITING.
- Follow-ups / TODOs
  - Decide how to handle deleted journal files (currently tasks/journals are not purged on missing files unless refreshed by month).
