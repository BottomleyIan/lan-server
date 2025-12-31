# Normalize tags to lowercase

- What changed: journal and journal_entry tag lists are normalized to lowercase before storing in the database.
- Why it changed: tag filtering was case-sensitive, causing missed matches.
- New conventions/decisions: tags are stored lowercase; UI should treat tags as case-insensitive.
- Follow-ups / TODOs: consider lowercasing tag filter inputs as well for consistency.
