package scanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"bottomley.ian/musicserver/internal/db"
	dbtypes "bottomley.ian/musicserver/internal/dbtypes"
	myfs "bottomley.ian/musicserver/internal/services/fs"
)

type Scanner struct {
	Q  *db.Queries
	FS myfs.FS
}

func New(q *db.Queries, fs myfs.FS) *Scanner {
	return &Scanner{
		Q:  q,
		FS: fs,
	}
}

var audioExt = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".flac": true,
	".aac":  true,
	".ogg":  true,
	".m4a":  true,
}

var ErrFolderUnavailable = errors.New("folder unavailable")

func isMusic(entry fs.DirEntry) (ext string, ok bool) {
	if entry.IsDir() {
		return "", false
	}
	ext = strings.ToLower(filepath.Ext(entry.Name()))
	return ext, audioExt[ext]
}

func (s *Scanner) ScanFolder(ctx context.Context, folderID int64) error {
	folder, err := s.Q.GetFolderByID(ctx, folderID)
	if err != nil {
		return err
	}
	root, err := myfs.ExpandPath(folder.Path)
	if err != nil {
		return err
	}
	info, statErr := s.FS.Stat(root)
	if statErr != nil {
		return fmt.Errorf("%w: %v", ErrFolderUnavailable, statErr)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %s is not a directory", ErrFolderUnavailable, root)
	}
	log.Printf("%s", root)
	return s.FS.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {

		if walkErr != nil {
			return walkErr
		}

		// Respect cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ext, ok := isMusic(d)
		if !ok {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		sizeBytes := info.Size()
		lastModified := info.ModTime().Unix()
		baseTitle := strings.TrimSuffix(d.Name(), ext)

		utp := db.UpsertTrackParams{
			FolderID:     folderID,
			RelPath:      rel,
			Title:        baseTitle,
			Filename:     d.Name(),
			Ext:          strings.TrimPrefix(ext, "."),
			SizeBytes:    sizeBytes,
			LastModified: lastModified,
		}
		track, err := s.Q.UpsertTrack(ctx, utp)
		if err != nil {
			return err
		}

		metadata, err := s.ReadMetadata(path)
		if err != nil {
			return err
		}
		artistID, albumID, albumRow, err := s.upsertArtistAlbum(ctx, metadata.Artist, metadata.Album)
		if err != nil {
			return err
		}

		var genre dbtypes.NullString
		var year dbtypes.NullInt64
		title := baseTitle
		if g := strings.TrimSpace(metadata.Genre); g != "" {
			genre = dbtypes.NullString{String: g, Valid: true}
		}
		if metadata.Year > 0 {
			year = dbtypes.NullInt64{Int64: int64(metadata.Year), Valid: true}
		}
		if t := strings.TrimSpace(metadata.Title); t != "" {
			title = t
		}
		var durationSeconds dbtypes.NullInt64
		if metadata.DurationSeconds != nil {
			durationSeconds = dbtypes.NullInt64{Int64: *metadata.DurationSeconds, Valid: true}
		}

		_, err = s.Q.UpdateTrackMetadata(ctx, db.UpdateTrackMetadataParams{
			ArtistID:        artistID,
			AlbumID:         albumID,
			Title:           title,
			Genre:           genre,
			Year:            year,
			ImagePath:       dbtypes.NullString{},
			DurationSeconds: durationSeconds,
			ID:              track.ID,
		})
		if err != nil {
			return err
		}

		if metadata.Picture != nil && albumID.Valid {
			savedPath, err := s.saveTrackImage(metadata.Picture, albumID.Int64, track.ID)
			if err != nil {
				log.Printf("warn: failed to save image for track %d: %v", track.ID, err)
			} else {
				_, err := s.Q.UpdateTrackImagePath(ctx, db.UpdateTrackImagePathParams{
					ImagePath: dbtypes.NullString{String: savedPath, Valid: true},
					ID:        track.ID,
				})
				if err != nil {
					log.Printf("warn: failed to set track image path %d: %v", track.ID, err)
				}
			}
		}

		if albumRow != nil && !albumRow.ImagePath.Valid {
			dirPath := filepath.Dir(path)
			candidate, err := s.findAlbumImageFile(dirPath)
			if err != nil {
				log.Printf("warn: failed to search album image in %s: %v", dirPath, err)
			} else if candidate != "" {
				saved, err := s.saveAlbumImage(candidate, albumRow.ID)
				if err != nil {
					log.Printf("warn: failed to save album image %s: %v", candidate, err)
				} else {
					_, err = s.Q.UpdateAlbumImagePath(ctx, db.UpdateAlbumImagePathParams{
						ImagePath: dbtypes.NullString{String: saved, Valid: true},
						ID:        albumRow.ID,
					})
					if err != nil {
						log.Printf("warn: failed to set album image path %d: %v", albumRow.ID, err)
					}
				}
			}
		}

		return err
	})
}

func (s *Scanner) upsertArtistAlbum(ctx context.Context, artistName, albumTitle string) (dbtypes.NullInt64, dbtypes.NullInt64, *db.Album, error) {
	var artistID dbtypes.NullInt64
	var albumID dbtypes.NullInt64
	var albumRow *db.Album

	if name := strings.TrimSpace(artistName); name != "" {
		artist, err := s.Q.UpsertArtist(ctx, name)
		if err != nil {
			return artistID, albumID, nil, err
		}
		artistID = dbtypes.NullInt64{Int64: artist.ID, Valid: true}
	}

	if title := strings.TrimSpace(albumTitle); title != "" && artistID.Valid {
		album, err := s.Q.UpsertAlbum(ctx, db.UpsertAlbumParams{
			ArtistID: artistID.Int64,
			Title:    title,
		})
		if err != nil {
			return artistID, albumID, nil, err
		}
		albumID = dbtypes.NullInt64{Int64: album.ID, Valid: true}
		albumRow = &album
	}

	return artistID, albumID, albumRow, nil
}

func (s *Scanner) saveTrackImage(pic *Picture, albumID, trackID int64) (string, error) {
	if pic == nil || len(pic.Data) == 0 {
		return "", fmt.Errorf("empty picture")
	}
	base := filepath.Join("tmp", "covers", fmt.Sprintf("%d", albumID))
	if err := s.FS.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	ext := ".img"
	switch strings.ToLower(pic.MIME) {
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	}
	filename := fmt.Sprintf("%d%s", trackID, ext)
	fullPath := filepath.Join(base, filename)
	if err := s.FS.WriteFile(fullPath, pic.Data, 0o644); err != nil {
		return "", err
	}
	rel := filepath.ToSlash(filepath.Join(fmt.Sprintf("%d", albumID), filename))
	return rel, nil
}

func (s *Scanner) findAlbumImageFile(dir string) (string, error) {
	entries, err := s.FS.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), "._") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		switch ext {
		case ".jpg", ".jpeg", ".png":
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", nil
}

func (s *Scanner) saveAlbumImage(srcPath string, albumID int64) (string, error) {
	f, err := s.FS.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := s.FS.MkdirAll(filepath.Join("tmp", "covers", "albums"), 0o755); err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == "" {
		ext = ".img"
	}

	destRel := filepath.ToSlash(filepath.Join("albums", fmt.Sprintf("%d%s", albumID, ext)))
	destFull := filepath.Join("tmp", "covers", destRel)

	out, err := s.FS.Create(destFull)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, f); err != nil {
		return "", err
	}
	return destRel, nil
}
