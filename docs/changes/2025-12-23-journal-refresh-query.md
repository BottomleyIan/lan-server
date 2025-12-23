# Journal refresh query

- What changed
  - Added `refresh=true` query support for `/journals/{year}/{month}` to drop existing rows for that month before reloading.
- Why it changed
  - Enable a dev-only reset of journal metadata while keeping the default scan/compare behavior.
- New conventions/decisions
  - Default behavior always scans and compares size/hash; `refresh=true` forces a month reset first.
- Follow-ups / TODOs
  - Consider documenting the dev-only nature of the refresh flag in the API README.
