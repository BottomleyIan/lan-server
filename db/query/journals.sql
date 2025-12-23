-- Get journal by date
-- name: GetJournalByDate :one
SELECT *
FROM journals
WHERE year = ?
  AND month = ?
  AND day = ?;

-- List journals for a month
-- name: ListJournalsByMonth :many
SELECT *
FROM journals
WHERE year = ?
  AND month = ?
ORDER BY day;

-- Upsert journal metadata
-- name: UpsertJournal :one
INSERT INTO journals (year, month, day, size_bytes, hash, tags, last_checked_at)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(year, month, day) DO UPDATE SET
  size_bytes = excluded.size_bytes,
  hash = excluded.hash,
  tags = excluded.tags,
  last_checked_at = CURRENT_TIMESTAMP
RETURNING *;
