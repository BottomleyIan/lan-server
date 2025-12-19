-- ---------- task_statuses ----------
CREATE TABLE IF NOT EXISTS task_statuses (
  code TEXT PRIMARY KEY,
  label TEXT NOT NULL
);

INSERT OR IGNORE INTO task_statuses (code, label) VALUES
  ('BACKLOG', 'Backlog'),
  ('TODO', 'Todo'),
  ('IN_PROGRESS', 'In Progress'),
  ('BLOCKED', 'Blocked'),
  ('WAITING', 'Waiting'),
  ('SCHEDULED', 'Scheduled'),
  ('REVIEW', 'Review'),
  ('REVISION', 'Revision'),
  ('TESTING', 'Testing'),
  ('APPROVED', 'Approved'),
  ('READY', 'Ready'),
  ('COMPLETED', 'Completed'),
  ('CANCELLED', 'Cancelled'),
  ('ARCHIVED', 'Archived');

-- ---------- tasks ----------
CREATE TABLE IF NOT EXISTS tasks (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NULL,
  tags TEXT NULL,
  status_code TEXT NOT NULL DEFAULT 'BACKLOG',

  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  FOREIGN KEY(status_code) REFERENCES task_statuses(code)
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status_code);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks(deleted_at);

CREATE TRIGGER IF NOT EXISTS tasks_set_updated_at
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
  UPDATE tasks
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;

-- ---------- task_transitions ----------
CREATE TABLE IF NOT EXISTS task_transitions (
  id INTEGER PRIMARY KEY,
  task_id INTEGER NOT NULL,
  status_code TEXT NOT NULL,
  reason TEXT NULL,
  changed_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  FOREIGN KEY(task_id) REFERENCES tasks(id),
  FOREIGN KEY(status_code) REFERENCES task_statuses(code)
);

CREATE INDEX IF NOT EXISTS idx_task_transitions_task_id ON task_transitions(task_id);
CREATE INDEX IF NOT EXISTS idx_task_transitions_status ON task_transitions(status_code);
CREATE INDEX IF NOT EXISTS idx_task_transitions_changed_at ON task_transitions(changed_at);
