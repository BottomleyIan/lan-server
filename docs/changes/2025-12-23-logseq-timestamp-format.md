# Logseq timestamp formatting

- What changed
  - Added ISO date/datetime parsing for task scheduled/deadline inputs and format them as Logseq timestamps when writing to journals.
- Why it changed
  - Accept `YYYY-MM-DD` or RFC3339 timestamps while writing Logseq-compatible `YYYY-MM-DD ddd HH:MM` or `YYYY-MM-DD ddd`.
- New conventions/decisions
  - RFC3339 inputs are converted to local time before formatting.
- Follow-ups / TODOs
  - Decide if timestamp formatting should preserve the original timezone instead of local time.
