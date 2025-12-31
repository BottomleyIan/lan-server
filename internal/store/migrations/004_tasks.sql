DROP TABLE IF EXISTS task_transitions;
DROP TABLE IF EXISTS task_statuses;
DROP TABLE IF EXISTS journal_entries;

-- ---------- journal_entries ----------
CREATE TABLE IF NOT EXISTS journal_entries (
  id INTEGER PRIMARY KEY,
  year INTEGER NOT NULL,
  month INTEGER NOT NULL,
  day INTEGER NOT NULL,
  journal_date TEXT NOT NULL,
  position INTEGER NOT NULL,
  title TEXT NOT NULL,
  raw_line TEXT NOT NULL,
  hash TEXT NOT NULL,
  body TEXT NULL,
  status TEXT NULL,
  tags TEXT NOT NULL DEFAULT '[]',
  type TEXT NOT NULL,
  scheduled_at TEXT NULL,
  deadline_at TEXT NULL,

  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE INDEX IF NOT EXISTS idx_journal_entries_date ON journal_entries(year, month, day);
CREATE INDEX IF NOT EXISTS idx_journal_entries_status ON journal_entries(status);

CREATE TRIGGER IF NOT EXISTS journal_entries_set_updated_at
AFTER UPDATE ON journal_entries
FOR EACH ROW
BEGIN
  UPDATE journal_entries
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;
