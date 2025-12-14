PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

-- ---------- folders (scan roots) ----------
CREATE TABLE IF NOT EXISTS folders (
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL UNIQUE,

  -- soft delete for roots
  deleted_at DATETIME NULL,

  -- availability / status (cached runtime state)
  available INTEGER NOT NULL DEFAULT 1 CHECK (available IN (0,1)),
  last_seen_at DATETIME NULL,

  -- scan status for the root
  last_scan_at DATETIME NULL,
  last_scan_status TEXT NULL,    -- "ok" | "error" | "skipped_unavailable" | "running"
  last_scan_error TEXT NULL,     -- error message for UI/debug

  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE INDEX IF NOT EXISTS idx_folders_available ON folders(available);

-- keep updated_at current on updates
CREATE TRIGGER IF NOT EXISTS folders_set_updated_at
AFTER UPDATE ON folders
FOR EACH ROW
BEGIN
  UPDATE folders
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;

-- ---------- artists ----------
CREATE TABLE IF NOT EXISTS artists (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,

  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE INDEX IF NOT EXISTS idx_artists_deleted_at ON artists(deleted_at);

CREATE TRIGGER IF NOT EXISTS artists_set_updated_at
AFTER UPDATE ON artists
FOR EACH ROW
BEGIN
  UPDATE artists
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;

-- ---------- albums ----------
CREATE TABLE IF NOT EXISTS albums (
  id INTEGER PRIMARY KEY,
  artist_id INTEGER NOT NULL,
  title TEXT NOT NULL,

  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  FOREIGN KEY(artist_id) REFERENCES artists(id),
  UNIQUE(artist_id, title)
);

CREATE INDEX IF NOT EXISTS idx_albums_artist_id ON albums(artist_id);
CREATE INDEX IF NOT EXISTS idx_albums_deleted_at ON albums(deleted_at);

CREATE TRIGGER IF NOT EXISTS albums_set_updated_at
AFTER UPDATE ON albums
FOR EACH ROW
BEGIN
  UPDATE albums
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;

-- ---------- tracks (indexed files) ----------
CREATE TABLE IF NOT EXISTS tracks (
  id INTEGER PRIMARY KEY,
  folder_id INTEGER NOT NULL,
  artist_id INTEGER NULL,
  album_id INTEGER NULL,

  rel_path TEXT NOT NULL,           -- relative to folders.path
  filename TEXT NOT NULL,
  ext TEXT NOT NULL,                -- lowercase, no dot (enforce in code)

  genre TEXT NULL,
  year INTEGER NULL,

  size_bytes INTEGER NOT NULL,
  last_modified INTEGER NOT NULL,   -- unix seconds since epoch

  -- last time this file was seen during scanning
  last_seen_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  FOREIGN KEY(folder_id) REFERENCES folders(id),
  FOREIGN KEY(artist_id) REFERENCES artists(id),
  FOREIGN KEY(album_id) REFERENCES albums(id),
  UNIQUE(folder_id, rel_path)
);

CREATE INDEX IF NOT EXISTS idx_tracks_folder_id ON tracks(folder_id);
CREATE INDEX IF NOT EXISTS idx_tracks_artist_id ON tracks(artist_id);
CREATE INDEX IF NOT EXISTS idx_tracks_album_id ON tracks(album_id);
CREATE INDEX IF NOT EXISTS idx_tracks_ext ON tracks(ext);
CREATE INDEX IF NOT EXISTS idx_tracks_deleted_at ON tracks(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tracks_folder_seen ON tracks(folder_id, last_seen_at);

-- keep updated_at current on updates
CREATE TRIGGER IF NOT EXISTS tracks_set_updated_at
AFTER UPDATE ON tracks
FOR EACH ROW
BEGIN
  UPDATE tracks
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;
