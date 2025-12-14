package scanner

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"bottomley.ian/musicserver/internal/db"
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
		log.Printf("name: %s, type: %s, ext: %s", d.Name(), d.Type(), ext)
		utp := db.UpsertTrackParams{
			FolderID:     folderID,
			RelPath:      rel,
			Filename:     d.Name(),
			Ext:          strings.TrimPrefix(ext, "."),
			SizeBytes:    sizeBytes,
			LastModified: lastModified,
		}
		_, err = s.Q.UpsertTrack(ctx, utp)

		return err
	})
}
