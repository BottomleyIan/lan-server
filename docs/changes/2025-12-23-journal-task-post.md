# Journal and task creation endpoints

- What changed
  - Added `POST /journals` to append a journal entry to today's journal file.
  - Added `POST /tasks` to append a Logseq task entry (status, tags, scheduled/deadline, body) to today's journal file.
  - New entries trigger a journal/task resync for the current day.
- Why it changed
  - Allow creating journal entries and tasks directly through the API while keeping the filesystem as source of truth.
- New conventions/decisions
  - Entries are appended to `YYYY_MM_DD.md` and body lines are indented by two spaces.
  - Tags are formatted as `[[tag]]` sequences with no separators.
- Follow-ups / TODOs
  - Consider validating scheduled/deadline formats more strictly if needed.
