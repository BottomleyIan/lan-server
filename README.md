# Music Server (Local Network)

A local-network Go server for indexing a music library on disk and exposing it via a JSON API (with Swagger docs). Designed for macOS initially, with a React frontend planned.

This project indexes filesystem content into SQLite for fast browsing/searching, while treating the filesystem as the source of truth.

See `TODO.md` for pending tasks.

---

## Current Status

Implemented:

* Go server skeleton (`cmd/server`)
* SQLite single-file DB (modernc driver)
* Migrations applied on startup (**create-only**, no destructive migration)
* sqlc-generated query layer
* Swagger docs (swaggo + http-swagger)
* HTTP handlers moved to `internal/handlers`
* DTO layer in handlers (no `sql.NullTime` leakage to Swagger/JSON)
* Scan service foundation:

  * `internal/services/fs` provides an FS abstraction (OS-backed now)
  * `internal/services/scanner` walks roots, filters audio files, captures size + mtime, and upserts tracks

Planned next:

* Finish scan bookkeeping (Start/Finish scan status + mark missing tracks)
* Add chi router + middleware (if not already switched)
* Add endpoints for folders/tracks/search + scan trigger/status
* Add metadata extraction and FTS search later

---

## Tech Stack

Backend:

* Go
* `net/http` (router currently plain mux; chi recommended next)
* SQLite (WAL mode)
* `modernc.org/sqlite` driver
* `sqlc` for typed SQL
* Swagger via `swaggo/swag` + `http-swagger`

Frontend:

* React (planned)

---

## Project Layout

```
cmd/server/                  # main entrypoint
internal/
  app/                       # dependency container (DB + Queries)
  db/                        # sqlc-generated code
  dbtypes/                   # aliases for NullTime/NullString (sqlc import workaround)
  handlers/                  # HTTP handlers + DTOs + mappers
  services/
    fs/                      # FS interface + OS implementation (WalkDir/Stat)
    scanner/                 # scan logic (filesystem -> DB upserts)
  store/                     # migrations embed + ApplyMigrations
docs/                        # swagger generated (not watched by air)
db/query/                    # sqlc query files
```

---

## Database

### Philosophy

* Filesystem is the **source of truth**
* SQLite DB is an **index/cache** for fast queries
* Roots may be unavailable (unmounted volumes); keep indexed data but exclude unavailable by default

### Tables

#### `folders` (scan roots)

Root paths only (e.g. `/Users/.../Music`, `/Volumes/SSD/Music`).
Includes cached availability + scan status:

* `available` (0/1)
* `last_seen_at`
* `last_scan_at`
* `last_scan_status` (`running|ok|error|skipped_unavailable`)
* `last_scan_error`

#### `tracks` (indexed files)

Tracks belong to a root folder and store file facts:

* `folder_id`
* `rel_path` (relative to folder root; unique per folder)
* `filename`, `ext` (lowercase, no dot)
* `size_bytes`
* `last_modified` (unix seconds)
* `last_seen_at` (used to mark missing files after a scan)
* `deleted_at` (soft delete)

---

## Migrations

Migrations are embedded and applied at startup. Current approach is **create-only** migrations (safe to run repeatedly).
At this stage the project uses **one init migration file**.

---

## Swagger

Swagger UI is served at:

* `http://localhost:8080/swagger/index.html`

Important notes:

* `http.ServeMux` subtree routing requires `/swagger/` (not `/swagger/*`)
* Swagger annotations must sit directly above named handler methods/functions
* API returns DTOs (not sqlc structs) to avoid Swagger issues with `sql.NullTime`

---

## Running

### Install tools

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### Generate sqlc

```bash
sqlc generate
```

### Generate swagger

From repo root (typical):

```bash
rm -rf docs
swag init -g main.go -d cmd/server,internal/handlers
```

### Run server

```bash
go run ./cmd/server
```

---

## Scanning (current implementation direction)

Scanner walks a folder root and upserts tracks:

* Uses `services/fs` interface for testability and future non-OS implementations
* Uses `filepath.Rel(root, path)` to store `rel_path`
* Uses `DirEntry.Info()` to capture `size_bytes` and `mtime`

Next scan milestones:

* Use `StartFolderScan` to obtain a scan start timestamp
* After scan: mark missing tracks via `last_seen_at < scan_start`
* Set folder scan status via Finish OK / Unavailable / Error

---

## Notes for Future Work

* Prefer chi for routing once endpoints expand (path params, middleware)
* Consider FTS5 for text search before embeddings
