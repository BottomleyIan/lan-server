# Playlist track update alignment

- What changed
  - Updated playlist track position updates to key on `playlist_id + track_id` instead of playlist track row ID.
  - Allowed position `0` when updating playlist track positions.
  - Swagger docs now describe the `track_id` param as a track ID.
- Why it changed
  - Aligns PUT and DELETE semantics on the same identifier and supports zero-based ordering.
- New conventions/decisions
  - Playlist track position updates are keyed by track ID within the playlist.
  - Position values can be zero-based.
- Follow-ups / TODOs
  - Consider whether `NextPlaylistPosition` should default to `0` for empty playlists if full zero-based indexing is desired.
