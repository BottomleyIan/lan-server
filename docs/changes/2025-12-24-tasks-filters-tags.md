# Tasks tags and filtering

- What changed: added tags storage for tasks, parsed logseq tags during journal sync, exposed tags in the /tasks response, and added query filters for statuses/status, tags, year, and month; updated SQL queries, generated sqlc code, and regenerated Swagger docs.
- Why it changed: /tasks needs to return tags like /journal and support filtering by status, tag, year, and month.
- New conventions/decisions: status filters accept the API status values (e.g., TODO, IN-PROGRESS) and expand to the underlying Logseq status set.
- Follow-ups / TODOs: none.
