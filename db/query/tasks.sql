-- ---------- task statuses ----------
-- name: ListTaskStatuses :many
SELECT *
FROM task_statuses
ORDER BY code;

-- ---------- tasks ----------
-- name: CreateTask :one
INSERT INTO tasks (title, body, tags, status_code)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListTasks :many
SELECT *
FROM tasks
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetTaskByID :one
SELECT *
FROM tasks
WHERE id = ?
  AND deleted_at IS NULL;

-- name: UpdateTask :one
UPDATE tasks
SET
  title = ?,
  body = ?,
  tags = ?
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status_code = ?
WHERE id = ?
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteTask :execrows
UPDATE tasks
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL;

-- ---------- task transitions ----------
-- name: AddTaskTransition :one
INSERT INTO task_transitions (task_id, status_code, reason)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListTaskTransitions :many
SELECT *
FROM task_transitions
WHERE task_id = ?
ORDER BY changed_at DESC;
