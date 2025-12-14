-- name: CreateFolder :one
INSERT INTO folders (path)
VALUES (?)
RETURNING *;

-- name: ListFolders :many
SELECT *
FROM folders
WHERE deleted_at IS NULL
ORDER BY path;

-- name: GetFolderByID :one
SELECT *
FROM folders
WHERE id = ?;

-- name: SoftDeleteFolder :one
UPDATE folders
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL
RETURNING *;

-- name: SetFolderAvailability :exec
UPDATE folders
SET
  available = ?,
  last_seen_at = CASE WHEN ? = 1 THEN CURRENT_TIMESTAMP ELSE last_seen_at END
WHERE id = ?;

-- Start a scan and capture a scan "start time" in DB timestamp format.
-- Use the returned last_scan_at value as the scan_start_time passed to MarkMissingTracksForFolder.
-- name: StartFolderScan :one
UPDATE folders
SET
  last_scan_at = CURRENT_TIMESTAMP,
  last_scan_status = 'running',
  last_scan_error = NULL
WHERE id = ?
RETURNING last_scan_at;

-- name: FinishFolderScanOK :exec
UPDATE folders
SET
  last_scan_status = 'ok',
  last_scan_error = NULL,
  last_seen_at = CURRENT_TIMESTAMP,
  available = 1
WHERE id = ?;

-- name: FinishFolderScanUnavailable :exec
UPDATE folders
SET
  last_scan_status = 'skipped_unavailable',
  last_scan_error = ?,
  available = 0
WHERE id = ?;

-- name: FinishFolderScanError :exec
UPDATE folders
SET
  last_scan_status = 'error',
  last_scan_error = ?,
  last_seen_at = CURRENT_TIMESTAMP,
  available = 1
WHERE id = ?;

