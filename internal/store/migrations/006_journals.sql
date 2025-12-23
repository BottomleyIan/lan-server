-- ---------- journals ----------
DROP TABLE IF EXISTS journals;

CREATE TABLE IF NOT EXISTS journals (
  year INTEGER NOT NULL,
  month INTEGER NOT NULL,
  day INTEGER NOT NULL,
  size_bytes INTEGER NOT NULL,
  hash TEXT NOT NULL,
  tags TEXT NOT NULL DEFAULT '[]',
  last_checked_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  PRIMARY KEY(year, month, day)
);

CREATE INDEX IF NOT EXISTS idx_journals_year_month ON journals(year, month);

CREATE TRIGGER IF NOT EXISTS journals_set_updated_at
AFTER UPDATE ON journals
FOR EACH ROW
BEGIN
  UPDATE journals
  SET updated_at = CURRENT_TIMESTAMP
  WHERE year = OLD.year
    AND month = OLD.month
    AND day = OLD.day;
END;
