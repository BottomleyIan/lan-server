-- ---------- journal_entries ----------
-- name: CreateJournalEntry :one
INSERT INTO journal_entries (
  year,
  month,
  day,
  journal_date,
  position,
  title,
  raw_line,
  hash,
  body,
  status,
  tags,
  property_keys,
  type,
  scheduled_at,
  deadline_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListJournalEntries :many
SELECT *
FROM journal_entries
WHERE (
    ?1 IS NULL
    OR journal_date LIKE ?1
    OR scheduled_at LIKE ?1
    OR deadline_at LIKE ?1
  )
  AND (
    ?2 IS NULL
    OR type = ?2
  )
  AND (
    ?3 IS NULL
    OR status IN (SELECT value FROM json_each(?3))
  )
  AND (
    ?4 IS NULL
    OR EXISTS (
      SELECT 1
      FROM json_each(journal_entries.tags)
      WHERE LOWER(value) IN (SELECT LOWER(value) FROM json_each(?4))
    )
  )
ORDER BY year DESC, month DESC, day DESC, position ASC;

-- name: ListJournalEntryTags :many
SELECT tags
FROM journal_entries;

-- name: ListJournalEntryPropertyKeys :many
SELECT property_keys
FROM journal_entries;

-- name: ListJournalEntryBodiesByPropertyKey :many
SELECT body
FROM journal_entries
WHERE EXISTS (
  SELECT 1
  FROM json_each(journal_entries.property_keys)
  WHERE LOWER(value) = LOWER(?)
);

-- name: GetJournalEntryByDateHash :one
SELECT *
FROM journal_entries
WHERE year = ?
  AND month = ?
  AND day = ?
  AND hash = ?
LIMIT 1;

-- name: GetJournalEntryByDatePosition :one
SELECT *
FROM journal_entries
WHERE year = ?
  AND month = ?
  AND day = ?
  AND position = ?
LIMIT 1;

-- name: DeleteJournalEntriesByDate :exec
DELETE FROM journal_entries
WHERE year = ?
  AND month = ?
  AND day = ?;

-- name: DeleteJournalEntriesByMonth :exec
DELETE FROM journal_entries
WHERE year = ?
  AND month = ?;
