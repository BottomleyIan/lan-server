-- List settings
-- name: ListSettings :many
SELECT *
FROM settings
ORDER BY key;

-- Get setting by key
-- name: GetSetting :one
SELECT *
FROM settings
WHERE key = ?;

-- Create setting
-- name: CreateSetting :one
INSERT INTO settings (key, value)
VALUES (?, ?)
RETURNING *;

-- Update setting
-- name: UpdateSetting :one
UPDATE settings
SET value = ?
WHERE key = ?
RETURNING *;

-- Delete setting
-- name: DeleteSetting :execrows
DELETE FROM settings
WHERE key = ?;
