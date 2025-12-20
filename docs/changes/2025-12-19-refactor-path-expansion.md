# Refactor path expansion helpers

- What changed
  - Centralized `~` and relative path expansion in `internal/services/fs`.
  - Updated scanner and handlers to use the shared helper and removed duplicate implementations.
- Why it changed
  - Avoid multiple copies of path expansion logic across the codebase.
- New conventions/decisions
  - Use `fs.ExpandUserPath` for `~` expansion and `fs.ExpandPath` for full normalization.
- Follow-ups / TODOs
  - None.
