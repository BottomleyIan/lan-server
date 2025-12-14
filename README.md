# Music Server (Local Network)

This project is a **local-network Go web server** for indexing and serving music files.
It exposes a JSON API (documented via Swagger/OpenAPI) and is intended to be paired with a React (or similar) frontend.

The system is designed to:

* Run on macOS (initially), LAN-only
* Index music files from user-defined root folders
* Store metadata in SQLite for fast searching
* Remain functional even when external volumes are temporarily unavailable

---

## Tech Stack

### Backend

* **Go**
* **net/http** + **chi** (routing)
* **SQLite** (single-file DB)
* **sqlc** (type-safe SQL access)
* **modernc.org/sqlite** (pure Go SQLite driver)
* **Swagger (swaggo)** for API documentation

### Frontend (planned)

* React or similar SPA
* Consumes JSON API only (no server-side rendering)

---

## Project Structure

```
cmd/server/           # main entry point
internal/
  app/                # App-wide dependency container
  handlers/           # HTTP handlers (Swagger-annotated)
  db/                 # sqlc-generated DB code
  store/              # migrations and DB helpers
docs/                 # Swagger-generated files
```

---

## Running the Server

```bash
go run ./cmd/server
```

Server runs on:

```
http://localhost:8080
```

Swagger UI:

```
http://localhost:8080/swagger/index.html
```

---

## Database

### SQLite

* Single file (default: `./data.sqlite`)
* WAL mode enabled
* Foreign keys enabled per connection

### Current Tables

#### `folders`

Represents **root scan locations** only (not every directory).

Example:

* `/Users/ianbottomley/Music`
* `/Volumes/SSD/Music`

Fields:

* `id`
* `path`
* `deleted_at`
* `created_at`
* `updated_at`

---

## Planned Database Expansion

### Folder Availability & Status

Folders may be temporarily unavailable (e.g. external volumes not mounted).
The system **retains indexed data** even when roots are unavailable.

Planned additional fields for `folders`:

* `available` (INTEGER 0/1)
* `last_seen_at`
* `last_scan_at`
* `last_scan_status` (`ok | error | skipped_unavailable`)
* `last_scan_error`

---

### Tracks / Files (planned)

Each music file will:

* Belong to exactly one root folder
* Be stored with a **relative path** from that root

Planned tables:

* `tracks`
* `track_metadata`

This allows:

* Fast search
* Stable identity across root moves
* Rebuilding the index if needed

---

## Design Principles

* **Filesystem is the source of truth**
* **Database is an index/cache**
* Indexed data is never deleted just because a volume is unavailable
* Availability is derived and cached, not assumed permanent
* No filesystem watchers required initially; explicit scans are used

---

## Development Notes

* Handlers are methods on a `Handlers` struct
* Shared dependencies live in `internal/app.App`
* Swagger annotations live directly above handler methods
* `chi` is used for routing and middleware
* Air is used for live reload (Swagger generation excluded from watch)

---

## Next Steps (Planned)

* Add `tracks` + metadata tables
* Implement folder scan logic
* Add search endpoints
* Build React UI
* Optional: background scan jobs

---
