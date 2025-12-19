# Fix track image path expansion

- What changed
  - Expanded `~` paths when saving and serving track images.
- Why it changed
  - Track folders stored with `~` were producing image paths that could not be located on disk.
- New conventions/decisions
  - Image paths are expanded to absolute paths when saved or served.
- Follow-ups / TODOs
  - Consider normalizing stored folder paths to absolute paths during migration or startup.
