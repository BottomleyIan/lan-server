-- Indexes to speed up prefix queries
CREATE INDEX IF NOT EXISTS idx_artists_name_prefix ON artists(name);
CREATE INDEX IF NOT EXISTS idx_albums_title_prefix ON albums(title);
CREATE INDEX IF NOT EXISTS idx_tracks_filename_prefix ON tracks(filename);
