-- ---------- tasks ----------
-- name: CreateTask :one
INSERT INTO tasks (
  year,
  month,
  day,
  position,
  title,
  body,
  status,
  tags,
  scheduled_at,
  deadline_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListTasks :many
SELECT *
FROM tasks
WHERE (?1 IS NULL OR year = ?1)
  AND (?2 IS NULL OR month = ?2)
  AND (
    ?3 IS NULL
    OR status IN (SELECT value FROM json_each(?3))
  )
  AND (
    ?4 IS NULL
    OR EXISTS (
      SELECT 1
      FROM json_each(tasks.tags)
      WHERE value IN (SELECT value FROM json_each(?4))
    )
  )
ORDER BY year DESC, month DESC, day DESC, position ASC;

-- name: DeleteTasksByDate :exec
DELETE FROM tasks
WHERE year = ?
  AND month = ?
  AND day = ?;

-- name: DeleteTasksByMonth :exec
DELETE FROM tasks
WHERE year = ?
  AND month = ?;
