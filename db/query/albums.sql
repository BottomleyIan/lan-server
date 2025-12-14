-- Get a single album by ID (excluding soft-deleted)
-- name: GetAlbumByID :one
SELECT *
FROM albums
WHERE id = ?
  AND deleted_at IS NULL;

-- List albums
-- name: ListAlbums :many
SELECT *
FROM albums
WHERE deleted_at IS NULL
ORDER BY title;

-- Update album title/artist
-- name: UpdateAlbum :one
UPDATE albums
SET artist_id = ?, title = ?
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;

-- Soft delete album
-- name: SoftDeleteAlbum :one
UPDATE albums
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;
