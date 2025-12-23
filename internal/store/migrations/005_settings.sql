-- ---------- settings ----------
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TRIGGER IF NOT EXISTS settings_set_updated_at
AFTER UPDATE ON settings
FOR EACH ROW
BEGIN
  UPDATE settings
  SET updated_at = CURRENT_TIMESTAMP
  WHERE key = OLD.key;
END;
