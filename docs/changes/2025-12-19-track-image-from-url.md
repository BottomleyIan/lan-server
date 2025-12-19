# Add track image update endpoint

- What changed
  - Added a POST endpoint to set a track image from a URL and save it alongside the track file.
  - Adjusted image path resolution to accept absolute paths.
- Why it changed
  - Allow manual track artwork updates by downloading and storing images per-track.
- New conventions/decisions
  - Track image files are saved in the track's folder with the track filename base and the image extension.
- Follow-ups / TODOs
  - Regenerate Swagger docs when you want the API docs updated.
