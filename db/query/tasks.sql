-- ---------- journal_entries ----------
-- name: CreateTask :one
INSERT INTO journal_entries (
  year,
  month,
  day,
  position,
  title,
  raw_line,
  hash,
  body,
  status,
  tags,
  type,
  scheduled_at,
  deadline_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListTasks :many
SELECT *
FROM journal_entries
WHERE status IS NOT NULL
  AND (
    ?1 IS NULL
    OR year = ?1
    OR substr(scheduled_at, 1, 4) = printf('%04d', ?1)
    OR substr(deadline_at, 1, 4) = printf('%04d', ?1)
  )
  AND (
    ?2 IS NULL
    OR month = ?2
    OR substr(scheduled_at, 6, 2) = printf('%02d', ?2)
    OR substr(deadline_at, 6, 2) = printf('%02d', ?2)
  )
  AND (
    ?5 IS NULL
    OR day = ?5
    OR substr(scheduled_at, 9, 2) = printf('%02d', ?5)
    OR substr(deadline_at, 9, 2) = printf('%02d', ?5)
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
      WHERE value IN (SELECT value FROM json_each(?4))
    )
  )
ORDER BY year DESC, month DESC, day DESC, position ASC;

-- name: GetJournalEntryByDateHash :one
SELECT *
FROM journal_entries
WHERE year = ?
  AND month = ?
  AND day = ?
  AND hash = ?
LIMIT 1;

-- name: ListNotes :many
SELECT *
FROM journal_entries
WHERE status IS NULL
  AND (?1 IS NULL OR year = ?1)
  AND (?2 IS NULL OR month = ?2)
  AND (?3 IS NULL OR day = ?3)
  AND (
    ?4 IS NULL
    OR EXISTS (
      SELECT 1
      FROM json_each(journal_entries.tags)
      WHERE value = ?4
    )
  )
ORDER BY year DESC, month DESC, day DESC, position ASC;

-- name: DeleteTasksByDate :exec
DELETE FROM journal_entries
WHERE year = ?
  AND month = ?
  AND day = ?;

-- name: DeleteTasksByMonth :exec
DELETE FROM journal_entries
WHERE year = ?
  AND month = ?;
