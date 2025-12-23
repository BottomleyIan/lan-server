# Settings CRUD API

- What changed
  - Added settings table migration and sqlc queries.
  - Added CRUD handlers for settings plus a hard-coded settings keys endpoint.
  - Wired settings routes and regenerated Swagger docs.
- Why it changed
  - Provide a simple place to store user-configurable values like journal folder and theme.
- New conventions/decisions
  - Settings keys are validated against a hard-coded list returned by `/settings/keys`.
- Follow-ups / TODOs
  - Expand the settings key list as new configuration needs arise.
