DROP TABLE IF EXISTS task_transitions;
DROP TABLE IF EXISTS task_statuses;
DROP TABLE IF EXISTS tasks;

-- ---------- tasks ----------
CREATE TABLE IF NOT EXISTS tasks (
  id INTEGER PRIMARY KEY,
  year INTEGER NOT NULL,
  month INTEGER NOT NULL,
  day INTEGER NOT NULL,
  position INTEGER NOT NULL,
  title TEXT NOT NULL,
  body TEXT NULL,
  status TEXT NOT NULL,
  scheduled_at TEXT NULL,
  deadline_at TEXT NULL,

  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE INDEX IF NOT EXISTS idx_tasks_date ON tasks(year, month, day);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

CREATE TRIGGER IF NOT EXISTS tasks_set_updated_at
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
  UPDATE tasks
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;
