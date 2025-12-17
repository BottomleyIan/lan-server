-- Create playlist
-- name: CreatePlaylist :one
INSERT INTO playlists (name) VALUES (?)
RETURNING *;

-- List playlists (excluding deleted)
-- name: ListPlaylists :many
SELECT *
FROM playlists
WHERE deleted_at IS NULL
ORDER BY name;

-- Get playlist by ID (excluding deleted)
-- name: GetPlaylistByID :one
SELECT *
FROM playlists
WHERE id = ?
  AND deleted_at IS NULL;

-- Soft delete playlist
-- name: SoftDeletePlaylist :execrows
UPDATE playlists
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL;

-- Update playlist name
-- name: UpdatePlaylist :one
UPDATE playlists
SET name = ?
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;
