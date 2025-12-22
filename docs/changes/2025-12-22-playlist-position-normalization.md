# Playlist position normalization

- What changed
  - Added playlist track position normalization after add, enqueue, update, and delete operations.
  - Insert/move operations now shift positions to preserve relative order before normalization.
  - Enqueue now uses the same position logic as add/move to keep a contiguous index.
- Why it changed
  - Playlist track positions were drifting into duplicates and gaps after mutations; positions are expected to represent array indices.
- New conventions/decisions
  - Playlist track positions are kept contiguous and zero-based after every mutation.
- Follow-ups / TODOs
  - Consider adding a maintenance endpoint or migration to normalize existing playlists in bulk.
