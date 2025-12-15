-- Next position for playlist
-- name: NextPlaylistPosition :one
SELECT COALESCE(MAX(position), 0) + 1
FROM playlist_tracks
WHERE playlist_id = ?
  AND deleted_at IS NULL;

-- Add playlist track (upsert by playlist+track)
-- name: AddPlaylistTrack :one
INSERT INTO playlist_tracks (playlist_id, track_id, position)
VALUES (?, ?, ?)
ON CONFLICT(playlist_id, track_id) DO UPDATE SET
  position = excluded.position,
  deleted_at = NULL
RETURNING *;

-- List playlist tracks with track metadata
-- name: ListPlaylistTracks :many
SELECT
  pt.id AS playlist_track_id,
  pt.playlist_id,
  pt.track_id,
  pt.position,
  pt.deleted_at AS playlist_track_deleted_at,
  pt.created_at AS playlist_track_created_at,
  pt.updated_at AS playlist_track_updated_at,

  t.id AS track_row_id,
  t.folder_id,
  t.artist_id,
  t.album_id,
  t.rel_path,
  t.filename,
  t.ext,
  t.genre,
  t.year,
  t.rating,
  t.image_path,
  t.size_bytes,
  t.last_modified,
  t.last_seen_at,
  t.deleted_at AS track_deleted_at,
  t.created_at AS track_created_at,
  t.updated_at AS track_updated_at
FROM playlist_tracks pt
JOIN tracks t ON t.id = pt.track_id
WHERE pt.playlist_id = ?
  AND pt.deleted_at IS NULL
  AND t.deleted_at IS NULL
ORDER BY pt.position;

-- Update playlist track position
-- name: UpdatePlaylistTrackPosition :one
UPDATE playlist_tracks
SET position = ?
WHERE id = ?
  AND playlist_id = ?
  AND deleted_at IS NULL
RETURNING *;
