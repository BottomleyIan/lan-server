-- If a playlist named "Now Playing" already exists with a different ID, rename it to avoid UNIQUE conflicts.
UPDATE playlists
SET name = name || ' (duplicate Now Playing)'
WHERE name = 'Now Playing' AND id <> 1;

-- Ensure ID 1 exists.
INSERT INTO playlists (id, name)
VALUES (1, 'Now Playing')
ON CONFLICT(id) DO NOTHING;

-- Normalize name at ID 1.
UPDATE playlists
SET name = 'Now Playing'
WHERE id = 1;
