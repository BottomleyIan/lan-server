-- Next position for playlist
-- name: NextPlaylistPosition :one
SELECT COALESCE(MAX(position), -1) + 1
FROM playlist_tracks
WHERE playlist_id = ?
  AND deleted_at IS NULL;

-- Get playlist track by playlist+track
-- name: GetPlaylistTrack :one
SELECT *
FROM playlist_tracks
WHERE playlist_id = ?
  AND track_id = ?
  AND deleted_at IS NULL;

-- Count playlist tracks
-- name: CountPlaylistTracks :one
SELECT COUNT(*)
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

-- Clear all tracks from playlist
-- name: ClearPlaylistTracks :exec
UPDATE playlist_tracks
SET deleted_at = CURRENT_TIMESTAMP
WHERE playlist_id = ?
  AND deleted_at IS NULL;

-- Shift playlist track positions up from a position (inclusive)
-- name: ShiftPlaylistTrackPositionsUpFrom :exec
UPDATE playlist_tracks
SET position = position + 1
WHERE playlist_id = ?
  AND deleted_at IS NULL
  AND position >= ?;

-- Shift playlist track positions down within a range (exclusive/inclusive)
-- name: ShiftPlaylistTrackPositionsDownRange :exec
UPDATE playlist_tracks
SET position = position - 1
WHERE playlist_id = ?
  AND deleted_at IS NULL
  AND position > ?
  AND position <= ?;

-- Shift playlist track positions up within a range (inclusive/exclusive)
-- name: ShiftPlaylistTrackPositionsUpRange :exec
UPDATE playlist_tracks
SET position = position + 1
WHERE playlist_id = ?
  AND deleted_at IS NULL
  AND position >= ?
  AND position < ?;

-- List playlist track IDs in position order
-- name: ListPlaylistTrackIDs :many
SELECT track_id
FROM playlist_tracks
WHERE playlist_id = ?
  AND deleted_at IS NULL
ORDER BY position, id;

-- Update playlist track position without returning row
-- name: UpdatePlaylistTrackPositionNoReturn :exec
UPDATE playlist_tracks
SET position = ?
WHERE playlist_id = ?
  AND track_id = ?
  AND deleted_at IS NULL;

-- List playlist tracks with track metadata
-- name: ListPlaylistTracks :many
SELECT
  sqlc.embed(pt),
  sqlc.embed(t),
  sqlc.embed(ar),
  sqlc.embed(al),
  sqlc.embed(al_ar)
FROM playlist_tracks pt
JOIN tracks t ON t.id = pt.track_id
LEFT JOIN artists ar ON ar.id = t.artist_id
LEFT JOIN albums al ON al.id = t.album_id
LEFT JOIN artists al_ar ON al_ar.id = al.artist_id
WHERE pt.playlist_id = ?
  AND pt.deleted_at IS NULL
  AND t.deleted_at IS NULL
ORDER BY pt.position;

-- Delete a track from playlist
-- name: DeletePlaylistTrack :execrows
UPDATE playlist_tracks
SET deleted_at = CURRENT_TIMESTAMP
WHERE playlist_id = ?
  AND track_id = ?
  AND deleted_at IS NULL;

-- Update playlist track position
-- name: UpdatePlaylistTrackPosition :one
UPDATE playlist_tracks
SET position = ?
WHERE playlist_id = ?
  AND track_id = ?
  AND deleted_at IS NULL
RETURNING *;
