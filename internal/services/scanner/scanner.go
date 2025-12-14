package scanner

import (
	"context"
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
			ArtistID: artistID,
			AlbumID:  albumID,
			Genre:    genre,
			Year:     year,
			ID:       track.ID,
		})

		return err
	})
}
