-- Get a single album by ID (excluding soft-deleted)
-- name: GetAlbumByID :one
SELECT *
FROM albums
WHERE id = ?
  AND deleted_at IS NULL;

-- Upsert album by artist/title (revives soft-deleted)
-- name: UpsertAlbum :one
INSERT INTO albums (artist_id, title)
VALUES (?, ?)
ON CONFLICT(artist_id, title) DO UPDATE SET
  artist_id = excluded.artist_id,
  title = excluded.title,
  deleted_at = NULL
RETURNING *;

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
