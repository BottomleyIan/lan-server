# Journal tags endpoint

- What changed: added a /journals/tags endpoint that returns all distinct tags derived from the journals table, with optional startswith filtering.
- Why it changed: clients need a tag list sourced from the journal index without scanning the filesystem.
- New conventions/decisions: startswith is the standard prefix filter query param for tag listings.
- Follow-ups / TODOs: none.
