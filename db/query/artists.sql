-- Get a single artist by ID (excluding soft-deleted)
-- name: GetArtistByID :one
SELECT *
FROM artists
WHERE id = ?
  AND deleted_at IS NULL;

-- Upsert artist by name (revives soft-deleted)
-- name: UpsertArtist :one
INSERT INTO artists (name)
VALUES (?)
ON CONFLICT(name) DO UPDATE SET
  name = excluded.name,
  deleted_at = NULL
RETURNING *;

-- List artists (optional prefix filter)
-- name: ListArtists :many
SELECT *
FROM artists
WHERE deleted_at IS NULL
  AND (?1 IS NULL OR name LIKE (?1 || '%'))
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
