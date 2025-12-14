-- name: CreateFolder :one
INSERT INTO folders (path)
VALUES (?)
RETURNING id, path, deleted_at, created_at, updated_at;

-- name: GetFolderByID :one
SELECT id, path, deleted_at, created_at, updated_at
FROM folders
WHERE id = ?;

-- name: ListFolders :many
SELECT id, path, deleted_at, created_at, updated_at
FROM folders
WHERE deleted_at IS NULL
ORDER BY path;

-- name: SoftDeleteFolder :exec
UPDATE folders
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL;

-- name: RestoreFolder :exec
UPDATE folders
SET deleted_at = NULL
WHERE id = ? AND deleted_at IS NOT NULL;
