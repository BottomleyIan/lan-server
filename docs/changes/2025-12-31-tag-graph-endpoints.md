# Tag graph endpoints

- What changed: added /journals/tags/graph and /journals/tags/graph/{tag} endpoints to compute tag co-occurrence graphs from journal tag lists; logs a timing line for each computation.
- Why it changed: provide a simple tag network for UI visualization.
- New conventions/decisions: tag graph counts co-occurrences within a single journal entry and normalizes tags to lower-case.
- Follow-ups / TODOs: consider caching if logs show slow runtimes.
