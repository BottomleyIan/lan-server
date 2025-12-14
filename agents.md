# AGENTS.md — Working Agreement for This Repo

This file describes how to work on this repository (for automated agents and humans) without losing architectural intent.

---

## What This Project Is

A **local network** Go server that indexes a music library on disk into SQLite and exposes a JSON API.
The filesystem remains the source of truth; SQLite is a fast index/cache.

---

## Core Design Rules

### Filesystem vs Database

* Filesystem = truth
* DB = index/cache
* DB must remain rebuildable from disk

### Folder Semantics

* `folders` table stores **scan roots only**
* Do **not** store every subdirectory as a row
* Tracks are linked to roots and stored via `rel_path` (relative path)

### Availability & Unmounted Volumes

* Roots may not be mounted (external SSD/NAS)
* Keep indexed data even if a root is unavailable
* Default queries should exclude unavailable roots (`folders.available = 1`)
* Provide optional “include unavailable” behavior for admin/debug

### Soft Delete

* Use `deleted_at` for tracks and roots
* Missing files are detected via scanning and soft-deleted (not hard deleted)

---

## Package Responsibilities

### `internal/app`

Holds shared dependencies (currently DB + sqlc Queries).
Do not put HTTP or scan logic here.

### `internal/services/fs`

Filesystem abstraction:

* interface returning `io/fs` types
* OS implementation uses `os.Stat` + `filepath.WalkDir`

### `internal/services/scanner`

Scan logic:

* no HTTP awareness
* accepts `context.Context`
* uses FS interface + sqlc Queries
* respects cancellation via `ctx.Done()`

### `internal/handlers`

HTTP layer:

* thin wrappers
* convert DB types to API DTOs
* Swagger annotations live here
* never return sqlc structs directly in Swagger responses (Swagger can’t parse `sql.Null*`)

### `internal/dbtypes`

Contains type aliases to `database/sql` null types:

* `NullTime`, `NullString`
  Used because sqlc import inference with SQLite/null types can be flaky. sqlc should reference `dbtypes.*` rather than `sql.*`.

---

## Swagger Rules

* Subtree route for `net/http` mux: use `/swagger/` (not `*`)
* Swagger annotations must be in doc comments directly above **named** handlers
* API types must be DTOs (no `sql.NullTime` / `sql.NullString` fields)
* Regenerate docs after changing annotations:

  * `swag init -g main.go -d cmd/server,internal/handlers`

---

## Database & Migrations Rules

* Migrations are embedded and applied at startup
* Current approach is create-only (safe to run repeatedly)
* If introducing multiple migrations later, consider a `schema_migrations` table to avoid reapplying large scripts

---

## Prompt / Change Summary Requirement (IMPORTANT)

Whenever an agent is used to assist development:

1. Create a Markdown summary of the prompt/decisions and what changed.
2. Include:

   * What changed
   * Why it changed
   * Any new conventions/decisions
   * Follow-ups / TODOs
3. Commit this summary alongside the code changes.

Suggested location:

```
docs/changes/YYYY-MM-DD-short-description.md
```

---

## What NOT To Do

* Don’t make scans a GET endpoint (use POST to trigger, GET to observe status)
* Don’t store absolute file paths as track identity (use `folder_id + rel_path`)
* Don’t update per-track “availability” when a volume unmounts (folder-level availability)
* Don’t leak DB-layer null types into API responses

---

## Implementation Defaults

* Use `context.Context` as the first arg to any function that does I/O or DB calls
* Prefer explicit errors and typed sentinel errors for mapping to HTTP responses
* Keep handler logic thin; keep scan logic in scanner service

