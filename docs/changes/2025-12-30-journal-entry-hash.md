# Journal entry hash

- What changed: added a hash column to journal_entries and compute it from the full raw block (first line + body lines); exposed hash on task and note DTOs; updated ingestion and schema.
- Why it changed: enable update/delete operations by stable content hash for tasks and notes.
- New conventions/decisions: hash uses FNV-1a 64-bit over the exact block text as stored in the file.
- Follow-ups / TODOs: add endpoints for update/delete by hash.
