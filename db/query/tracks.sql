-- Upsert a track discovered during scan.
-- Always bumps last_seen_at to CURRENT_TIMESTAMP.
-- name: UpsertTrack :one
INSERT INTO tracks (
  folder_id, rel_path, filename, ext, size_bytes, last_modified, last_seen_at, deleted_at
) VALUES (
  ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, NULL
)
ON CONFLICT(folder_id, rel_path) DO UPDATE SET
  filename      = excluded.filename,
  ext           = excluded.ext,
  size_bytes    = excluded.size_bytes,
  last_modified = excluded.last_modified,
  last_seen_at  = CURRENT_TIMESTAMP,
  deleted_at    = NULL
RETURNING *;

-- List tracks for a folder (excluding deleted)
-- name: ListTracksForFolder :many
SELECT *
FROM tracks
WHERE folder_id = ? AND deleted_at IS NULL
ORDER BY rel_path;

-- Get a single track by ID (excluding deleted)
-- name: GetTrackByID :one
SELECT *
FROM tracks
WHERE id = ?
  AND deleted_at IS NULL;

-- Mark tracks missing if not seen during this scan pass.
-- Pass scan_start_time from StartFolderScan (folders.last_scan_at returned value).
-- name: MarkMissingTracksForFolder :exec
UPDATE tracks
SET deleted_at = CURRENT_TIMESTAMP
WHERE folder_id = ?
  AND deleted_at IS NULL
  AND last_seen_at < ?;

-- Default: list all playable tracks (roots currently available)
-- name: ListPlayableTracks :many
SELECT t.*
FROM tracks t
JOIN folders f ON f.id = t.folder_id
WHERE t.deleted_at IS NULL
  AND f.deleted_at IS NULL
  AND f.available = 1
ORDER BY t.rel_path;

-- Include unavailable roots too (for admin/debug UI)
-- name: ListAllIndexedTracks :many
SELECT t.*
FROM tracks t
JOIN folders f ON f.id = t.folder_id
WHERE t.deleted_at IS NULL
  AND f.deleted_at IS NULL
ORDER BY t.rel_path;

-- Optional: get absolute path pieces for playback (folder path + rel path)
-- name: GetPlayableTrackPathPartsByID :one
SELECT
  t.id,
  f.path AS folder_path,
  t.rel_path
FROM tracks t
JOIN folders f ON f.id = t.folder_id
WHERE t.id = ?
  AND t.deleted_at IS NULL
  AND f.deleted_at IS NULL
  AND f.available = 1;

-- Update track rating (nullable)
-- name: UpdateTrackRating :one
UPDATE tracks
SET rating = ?
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;

-- Update track metadata from tags
-- name: UpdateTrackMetadata :one
UPDATE tracks
SET artist_id = ?, album_id = ?, genre = ?, year = ?, image_path = COALESCE(?, image_path)
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;

-- Update track image path (only set when provided)
-- name: UpdateTrackImagePath :one
UPDATE tracks
SET image_path = ?
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;
