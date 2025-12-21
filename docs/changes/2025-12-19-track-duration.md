# Add track duration support

- What changed
  - Added `duration_seconds` to tracks schema, sqlc overrides, and DTOs.
  - Extracted duration via ffmpeg-go during scanning and store in `tracks.duration_seconds`.
  - Enforced ffmpeg/ffprobe presence at startup (panic if missing).
- Why it changed
  - Provide track length in the API for UI and sorting.
- New conventions/decisions
  - Duration is stored as integer seconds derived from ffprobe output.
- Follow-ups / TODOs
  - Regenerate Swagger docs when you want the API docs updated.
