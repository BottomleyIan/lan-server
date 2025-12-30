# Journal entry update/delete

- What changed: added update and delete endpoints for tasks and notes by date+hash; updates require an If-Match header and accept a raw block payload; shared journal file editing and re-sync logic.
- Why it changed: enable mutation of notes and tasks by hash without relying on numeric IDs.
- New conventions/decisions: updates require If-Match equal to the hash path param; payload is the full raw block text.
- Follow-ups / TODOs: ensure schema rebuild to include hash column.
