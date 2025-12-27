-- Get journal by date
-- name: GetJournalByDate :one
SELECT *
FROM journals
WHERE year = ?
  AND month = ?
  AND day = ?;

-- Update journal last checked timestamp
-- name: UpdateJournalLastChecked :exec
UPDATE journals
SET last_checked_at = CURRENT_TIMESTAMP
WHERE year = ?
  AND month = ?
  AND day = ?;

-- Delete journals for a month
-- name: DeleteJournalsByMonth :exec
DELETE FROM journals
WHERE year = ?
  AND month = ?;

-- List journals with optional filters
-- name: ListJournalsFiltered :many
SELECT *
FROM journals
WHERE (?1 IS NULL OR year = ?1)
  AND (?2 IS NULL OR month = ?2)
  AND (?3 IS NULL OR day = ?3)
  AND (
    ?4 IS NULL
    OR EXISTS (
      SELECT 1
      FROM json_each(journals.tags)
      WHERE value = ?4
    )
  )
ORDER BY year DESC, month DESC, day DESC;

-- List journals for a month
-- name: ListJournalsByMonth :many
SELECT *
FROM journals
WHERE year = ?
  AND month = ?
ORDER BY day;

-- List all journal tags (raw JSON per journal)
-- name: ListJournalTags :many
SELECT tags
FROM journals;

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
