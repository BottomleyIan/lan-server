package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"bottomley.ian/musicserver/internal/db"
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

type updateTrackImageRequest struct {
	URL string `json:"url"`
}

// UpdateTrackImage godoc
// @Summary Update track image from URL
// @Tags images
// @Accept json
// @Produce json
// @Param id path int true "Track ID"
// @Param request body updateTrackImageRequest true "Image URL payload"
// @Success 200 {object} TrackDTO
// @Router /tracks/{id}/image [post]
func (h *Handlers) UpdateTrackImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body updateTrackImageRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	imageURL := strings.TrimSpace(body.URL)
	if imageURL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	parsed, err := url.Parse(imageURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	track, err := h.App.Queries.GetTrackByID(r.Context(), id)
	if err != nil {
		http.Error(w, "track not found", http.StatusNotFound)
		return
	}

	folder, err := h.App.Queries.GetFolderByID(r.Context(), track.FolderID)
	if err != nil {
		http.Error(w, "folder not found", http.StatusNotFound)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, imageURL, nil)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "unable to fetch image", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "unable to fetch image", http.StatusBadGateway)
		return
	}

	const maxImageBytes = 10 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageBytes+1))
	if err != nil {
		http.Error(w, "unable to read image", http.StatusBadGateway)
		return
	}
	if len(data) > maxImageBytes {
		http.Error(w, "image too large", http.StatusBadRequest)
		return
	}

	ext, err := imageExtension(parsed.Path, resp.Header.Get("Content-Type"), data)
	if err != nil {
		http.Error(w, "unsupported image type", http.StatusBadRequest)
		return
	}

	destPath, err := trackImagePath(folder.Path, track.RelPath, track.Filename, ext)
	if err != nil {
		http.Error(w, "invalid track path", http.StatusBadRequest)
		return
	}

	if err := h.App.FS.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		http.Error(w, "unable to prepare destination", http.StatusInternalServerError)
		return
	}
	if err := h.App.FS.WriteFile(destPath, data, 0o644); err != nil {
		http.Error(w, "unable to save image", http.StatusInternalServerError)
		return
	}

	row, err := h.App.Queries.UpdateTrackImagePath(r.Context(), db.UpdateTrackImagePathParams{
		ImagePath: dbtypes.NullString{String: destPath, Valid: true},
		ID:        track.ID,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, trackDTOFromDB(row))
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
	expanded, err := myfs.ExpandUserPath(rel)
	if err == nil && filepath.IsAbs(expanded) {
		return expanded
	}
	return filepath.Join("tmp", "covers", rel)
}

func imageExtension(urlPath string, contentType string, data []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(urlPath))
	if ext != "" {
		ext = normalizeImageExtension(ext)
	}
	if ext == "" {
		ct := strings.ToLower(strings.TrimSpace(contentType))
		if ct == "" && len(data) > 0 {
			ct = strings.ToLower(http.DetectContentType(data))
		}
		ext = extensionFromContentType(ct)
	}
	if ext == "" {
		return "", fmt.Errorf("unsupported image type")
	}
	return ext, nil
}

func normalizeImageExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpeg":
		return ".jpg"
	case ".jpg", ".png", ".gif", ".webp":
		return ext
	default:
		return ""
	}
}

func extensionFromContentType(ct string) string {
	if ct == "" {
		return ""
	}
	switch strings.Split(ct, ";")[0] {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func trackImagePath(folderPath, relPath, filename, ext string) (string, error) {
	if folderPath == "" || filename == "" {
		return "", fmt.Errorf("missing track path")
	}
	expanded, err := myfs.ExpandUserPath(folderPath)
	if err != nil {
		return "", err
	}
	folderPath = expanded
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" {
		return "", fmt.Errorf("missing track name")
	}
	relDir := filepath.Dir(filepath.FromSlash(relPath))
	if relDir == "." {
		relDir = ""
	}
	destDir := filepath.Join(folderPath, relDir)
	return filepath.Join(destDir, base+ext), nil
}
