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
WHERE (sqlc.narg('year') IS NULL OR year = sqlc.narg('year'))
  AND (sqlc.narg('month') IS NULL OR month = sqlc.narg('month'))
  AND (
    sqlc.narg('statuses') IS NULL
    OR status IN (SELECT value FROM json_each(sqlc.narg('statuses')))
  )
  AND (
    sqlc.narg('tags') IS NULL
    OR EXISTS (
      SELECT 1
      FROM json_each(tasks.tags)
      WHERE value IN (SELECT value FROM json_each(sqlc.narg('tags')))
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
