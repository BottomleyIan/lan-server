# Task status mapping

- What changed
  - Mapped Logseq task statuses to frontend status values before returning JSON.
- Why it changed
  - Frontend expects a reduced set of statuses: TODO, IN-PROGRESS, DONE, CANCELLED.
- New conventions/decisions
  - LATER/NOW/TODO -> TODO; DOING/IN-PROGRESS/WAITING -> IN-PROGRESS; DONE/CANCELLED preserved.
- Follow-ups / TODOs
  - Confirm whether WAITING should map to IN-PROGRESS or TODO.
