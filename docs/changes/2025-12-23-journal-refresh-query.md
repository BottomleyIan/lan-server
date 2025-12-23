# Journal refresh query

- What changed
  - Added `refresh=true` query support for `/journals/{year}/{month}` to control filesystem refresh.
- Why it changed
  - Allow clients to avoid rescanning disk unless explicitly requested.
- New conventions/decisions
  - When `refresh` is omitted or false, the endpoint returns cached DB results only.
- Follow-ups / TODOs
  - Consider documenting default behavior in the API README.
