-- Get a single artist by ID (excluding soft-deleted)
-- name: GetArtistByID :one
SELECT *
FROM artists
WHERE id = ?
  AND deleted_at IS NULL;

-- List artists
-- name: ListArtists :many
SELECT *
FROM artists
WHERE deleted_at IS NULL
ORDER BY name;

-- Update artist name
-- name: UpdateArtist :one
UPDATE artists
SET name = ?
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;

-- Soft delete artist
-- name: SoftDeleteArtist :one
UPDATE artists
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;
