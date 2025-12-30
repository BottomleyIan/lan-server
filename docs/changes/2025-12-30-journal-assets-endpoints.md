# Journal assets endpoints

- What changed: added GET/POST /journals/assets for serving and uploading Logseq asset files; introduced JournalAssetDTO and response path formatting; wired routes and Swagger docs.
- Why it changed: Logseq images are stored in the journals assets folder and need API access.
- New conventions/decisions: assets are addressed by a path query param (accepts ../assets/...) and stored under the journals assets folder; POST accepts multipart file uploads.
- Follow-ups / TODOs: none.
