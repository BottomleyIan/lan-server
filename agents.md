# Agent Instructions for This Repository

This file exists to ensure **continuity, correctness, and efficiency** when working with automated agents or future contributors.

---

## High-Level Intent

This project is a **local music indexing and API server**, not a file-serving daemon and not a streaming-first system.

Key goals:

* Deterministic behavior
* Clear separation of concerns
* SQLite as an index, not a primary data store
* Designed to work even when media volumes are intermittently unavailable

---

## Core Architectural Rules

### 1. Filesystem vs Database

* Filesystem is the **source of truth**
* Database is a **search/index layer**
* Indexed data is preserved even if files are temporarily unavailable

### 2. Folder Semantics

* `folders` table contains **root scan locations only**
* Subdirectories are **not** stored as separate rows
* Files are linked to roots using relative paths

### 3. Availability

* Volume availability is **runtime state**
* Stored fields represent *last known* state
* Do not delete or invalidate tracks when a root is unavailable

---

## Code Structure Rules

* Shared dependencies go in `internal/app.App`
* HTTP handlers live in `internal/handlers`
* Handlers are methods, not anonymous functions
* Routing uses `chi`
* Swagger annotations must live directly above handler methods
* Database access must go through sqlc-generated code

---

## Database Rules

* All schema changes must be migrations
* All queries must be written in SQL and compiled via sqlc
* Avoid ORMs
* Prefer explicit fields over generic JSON blobs

---

## Swagger / API Documentation

* Swagger is documentation, not validation
* Swagger must reflect real behavior
* `@Router` paths must match actual routes
* Use DTOs if exposing DB types becomes awkward

---

## Agent Workflow Requirement (IMPORTANT)

**Any time an agent is used to assist development:**

1. The prompt or discussion **must be summarized** in a Markdown file
2. The summary must include:

   * What was changed
   * Why the change was made
   * Any architectural decisions
3. This Markdown file should be committed **alongside the code changes**

This ensures:

* Architectural intent is preserved
* Changes are auditable
* Future agents can reason correctly without re-discovery

Suggested location:

```
docs/changes/YYYY-MM-DD-description.md
```

---

## What Not To Do

* Do not auto-index every directory as a DB row
* Do not tie DB identity to absolute paths
* Do not assume volumes are always mounted
* Do not introduce heavy frameworks
* Do not hide logic inside middleware magic

---

## Philosophy

Favor:

* Explicitness over cleverness
* Rebuildability over brittleness
* Simplicity over premature optimization

If unsure, **document the decision**.

---
