-- Get a single album by ID (excluding soft-deleted)
-- name: GetAlbumByID :one
SELECT *
FROM albums
WHERE id = ?
  AND deleted_at IS NULL;

-- Get a single album by ID with artist info
-- name: GetAlbumWithArtist :one
SELECT
  sqlc.embed(a),
  sqlc.embed(ar)
FROM albums a
LEFT JOIN artists ar ON ar.id = a.artist_id
WHERE a.id = ?
  AND a.deleted_at IS NULL;

-- Update album image path (only if currently NULL)
-- name: UpdateAlbumImagePath :one
UPDATE albums
SET image_path = COALESCE(?, image_path)
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;

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
  AND (?1 IS NULL OR LOWER(title) LIKE (LOWER(?1) || '%'))
ORDER BY title;

-- List albums with artist info
-- name: ListAlbumsWithArtist :many
SELECT
  sqlc.embed(a),
  sqlc.embed(ar)
FROM albums a
LEFT JOIN artists ar ON ar.id = a.artist_id
WHERE a.deleted_at IS NULL
  AND (?1 IS NULL OR LOWER(a.title) LIKE (LOWER(?1) || '%'))
ORDER BY a.title;

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
