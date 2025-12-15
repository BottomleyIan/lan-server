package handlers

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	dbtypes "bottomley.ian/musicserver/internal/dbtypes"
	myfs "bottomley.ian/musicserver/internal/services/fs"

	"github.com/go-chi/chi/v5"
)

// GetTrackImage godoc
// @Summary Get track image
// @Tags images
// @Produce image/jpeg
// @Produce image/png
// @Produce application/octet-stream
// @Param id path int true "Track ID"
// @Success 200
// @Router /tracks/{id}/image [get]
func (h *Handlers) GetTrackImage(w http.ResponseWriter, r *http.Request) {
	h.serveImageWithFallback(w, r, true)
}

// GetAlbumImage godoc
// @Summary Get album image
// @Tags images
// @Produce image/jpeg
// @Produce image/png
// @Produce application/octet-stream
// @Param id path int true "Album ID"
// @Success 200
// @Router /albums/{id}/image [get]
func (h *Handlers) GetAlbumImage(w http.ResponseWriter, r *http.Request) {
	h.serveImageWithFallback(w, r, false)
}

func (h *Handlers) serveImageWithFallback(w http.ResponseWriter, r *http.Request, isTrack bool) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var imagePath string
	if isTrack {
		imagePath, err = h.resolveTrackImagePath(r.Context(), id)
	} else {
		imagePath, err = h.resolveAlbumImagePath(r.Context(), id)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	data, ctype, err := readImageFile(h.App.FS, imagePath)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	if ctype == "" {
		ctype = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ctype)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func readImageFile(fs myfs.FS, path string) ([]byte, string, error) {
	b, err := fs.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	ext := strings.ToLower(filepath.Ext(path))
	ctype := mime.TypeByExtension(ext)
	return b, ctype, nil
}

func (h *Handlers) resolveTrackImagePath(ctx context.Context, id int64) (string, error) {
	track, err := h.App.Queries.GetTrackByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("track not found")
	}

	if track.ImagePath.Valid {
		return coverFullPath(track.ImagePath.String), nil
	}

	if track.AlbumID.Valid {
		album, err := h.App.Queries.GetAlbumByID(ctx, track.AlbumID.Int64)
		if err == nil && album.ImagePath.Valid {
			return coverFullPath(album.ImagePath.String), nil
		}
		if err == nil {
			img, err := h.App.Queries.GetFirstTrackImageForAlbum(ctx, dbtypes.NullInt64{Int64: album.ID, Valid: true})
			if err == nil && img.Valid {
				return coverFullPath(img.String), nil
			}
		}
	}

	return "", fmt.Errorf("image not found")
}

func (h *Handlers) resolveAlbumImagePath(ctx context.Context, id int64) (string, error) {
	album, err := h.App.Queries.GetAlbumByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("album not found")
	}

	if album.ImagePath.Valid {
		return coverFullPath(album.ImagePath.String), nil
	}

	img, err := h.App.Queries.GetFirstTrackImageForAlbum(ctx, dbtypes.NullInt64{Int64: album.ID, Valid: true})
	if err == nil && img.Valid {
		return coverFullPath(img.String), nil
	}

	return "", fmt.Errorf("image not found")
}

func coverFullPath(rel string) string {
	if rel == "" {
		return ""
	}
	return filepath.Join("tmp", "covers", rel)
}
