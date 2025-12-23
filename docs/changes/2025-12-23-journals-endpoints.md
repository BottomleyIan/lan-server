# Journals endpoints and metadata

- What changed
  - Added journals table migration and sqlc queries for upserting and listing journal metadata.
  - Added `/journals/{year}/{month}` and `/journals/{year}/{month}/{day}` endpoints backed by the Journals Folder setting.
  - Journal scans now compare size/hash against stored metadata to decide whether to upsert or only refresh `last_checked_at`.
  - Updated the Journals Folder setting definition and regenerated Swagger docs.
- Why it changed
  - Provide month/day journal APIs with metadata tracking (size, hash, tags) for logseq journals.
- New conventions/decisions
  - Journals are identified by `year/month/day` derived from `YYYY_MM_DD.md` filenames and stored with zero-padded parsing.
  - Tags are stored as a JSON array string on each journal row.
- Follow-ups / TODOs
  - Consider a dedicated tags table if we need efficient tag searches across journals.
