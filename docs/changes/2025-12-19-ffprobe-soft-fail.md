# Soften ffprobe failures during scan

- What changed
  - ffprobe duration errors now log a warning and scanning continues.
- Why it changed
  - Avoid aborting scans when ffprobe fails for a single file.
- New conventions/decisions
  - ffprobe failure is non-fatal per-file; missing ffmpeg/ffprobe still panics at startup.
- Follow-ups / TODOs
  - None.
