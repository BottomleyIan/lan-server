package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
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
	root, err := expandPath(folder.Path)
	if err != nil {
		return err
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

		utp := db.UpsertTrackParams{
			FolderID:     folderID,
			RelPath:      rel,
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
		var artistID dbtypes.NullInt64
		var albumID dbtypes.NullInt64
		var genre dbtypes.NullString
		var year dbtypes.NullInt64

		if name := strings.TrimSpace(metadata.Artist); name != "" {
			artist, err := s.Q.UpsertArtist(ctx, name)
			if err != nil {
				return err
			}
			artistID = dbtypes.NullInt64{Int64: artist.ID, Valid: true}
		}

		if title := strings.TrimSpace(metadata.Album); title != "" && artistID.Valid {
			album, err := s.Q.UpsertAlbum(ctx, db.UpsertAlbumParams{
				ArtistID: artistID.Int64,
				Title:    title,
			})
			if err != nil {
				return err
			}
			albumID = dbtypes.NullInt64{Int64: album.ID, Valid: true}
		}

		if g := strings.TrimSpace(metadata.Genre); g != "" {
			genre = dbtypes.NullString{String: g, Valid: true}
		}
		if metadata.Year > 0 {
			year = dbtypes.NullInt64{Int64: int64(metadata.Year), Valid: true}
		}

		_, err = s.Q.UpdateTrackMetadata(ctx, db.UpdateTrackMetadataParams{
			ArtistID:  artistID,
			AlbumID:   albumID,
			Genre:     genre,
			Year:      year,
			ImagePath: dbtypes.NullString{},
			ID:        track.ID,
		})
		if err != nil {
			return err
		}

		if metadata.Picture != nil && albumID.Valid {
			savedPath, err := saveTrackImage(metadata.Picture, albumID.Int64, track.ID)
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
				_, err = s.Q.UpdateAlbumImagePath(ctx, db.UpdateAlbumImagePathParams{
					ImagePath: dbtypes.NullString{String: savedPath, Valid: true},
					ID:        albumID.Int64,
				})
				if err != nil {
					log.Printf("warn: failed to set album image path %d: %v", albumID.Int64, err)
				}
			}
		}

		return err
	})
}

func saveTrackImage(pic *Picture, albumID, trackID int64) (string, error) {
	if pic == nil || len(pic.Data) == 0 {
		return "", fmt.Errorf("empty picture")
	}
	base := filepath.Join("tmp", "covers", fmt.Sprintf("%d", albumID))
	if err := os.MkdirAll(base, 0o755); err != nil {
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
	if err := os.WriteFile(fullPath, pic.Data, 0o644); err != nil {
		return "", err
	}
	rel := filepath.ToSlash(filepath.Join(fmt.Sprintf("%d", albumID), filename))
	return rel, nil
}
