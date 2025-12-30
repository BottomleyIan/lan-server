# Serialize journal sync

- What changed: added a mutex around journal sync to avoid concurrent task/note list requests conflicting during import.
- Why it changed: concurrent syncs can hit SQLite locking and cause 500s.
- New conventions/decisions: journal sync runs serialized via a handler-level mutex.
- Follow-ups / TODOs: none.
