-- folders table
CREATE TABLE IF NOT EXISTS folders (
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL UNIQUE,
  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

-- keep updated_at current on updates
CREATE TRIGGER IF NOT EXISTS folders_set_updated_at
AFTER UPDATE ON folders
FOR EACH ROW
BEGIN
  UPDATE folders
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;

-- recommended pragmas (you can also set these from Go)
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
