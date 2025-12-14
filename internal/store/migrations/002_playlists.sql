-- ---------- playlists ----------
CREATE TABLE IF NOT EXISTS playlists (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,

  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TRIGGER IF NOT EXISTS playlists_set_updated_at
AFTER UPDATE ON playlists
FOR EACH ROW
BEGIN
  UPDATE playlists
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;

-- ---------- playlist_tracks ----------
CREATE TABLE IF NOT EXISTS playlist_tracks (
  id INTEGER PRIMARY KEY,
  playlist_id INTEGER NOT NULL,
  track_id INTEGER NOT NULL,
  position INTEGER NOT NULL,

  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  FOREIGN KEY(playlist_id) REFERENCES playlists(id),
  FOREIGN KEY(track_id) REFERENCES tracks(id),
  UNIQUE(playlist_id, track_id)
);

CREATE INDEX IF NOT EXISTS idx_playlist_tracks_playlist ON playlist_tracks(playlist_id);
CREATE INDEX IF NOT EXISTS idx_playlist_tracks_position ON playlist_tracks(playlist_id, position);
CREATE INDEX IF NOT EXISTS idx_playlist_tracks_track ON playlist_tracks(track_id);

CREATE TRIGGER IF NOT EXISTS playlist_tracks_set_updated_at
AFTER UPDATE ON playlist_tracks
FOR EACH ROW
BEGIN
  UPDATE playlist_tracks
  SET updated_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;
END;
