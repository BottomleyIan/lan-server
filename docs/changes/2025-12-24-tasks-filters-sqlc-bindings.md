# Fix task filters SQL bindings

- What changed: switched the tasks list query to positional bindings so SQLite receives bound JSON values instead of raw sqlc.narg() calls; updated generated sqlc code and handler param names accordingly.
- Why it changed: /tasks returned 500 due to sqlite errors from unresolved sqlc.narg() calls inside json_each.
- New conventions/decisions: avoid sqlc.narg inside json_each subqueries; use positional params for JSON list filters.
- Follow-ups / TODOs: none.
