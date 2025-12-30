# Notes sync on list

- What changed: notes listing now triggers the same journal import/hash check logic as tasks; shared the sync logic between /tasks and /notes.
- Why it changed: notes should reflect the latest journal file contents before listing.
- New conventions/decisions: use shared syncJournalsFromDisk for endpoints that depend on journal imports.
- Follow-ups / TODOs: none.
