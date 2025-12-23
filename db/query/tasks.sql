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
  scheduled_at,
  deadline_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListTasks :many
SELECT *
FROM tasks
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
