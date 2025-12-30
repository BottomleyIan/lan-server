package handlers

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

const maxJournalAssetBytes = 20 << 20

// GetJournalAsset godoc
// @Summary Get journal asset
// @Tags journals
// @Produce application/octet-stream
// @Param path query string true "Asset path or filename"
// @Success 200
// @Router /journals/assets [get]
func (h *Handlers) GetJournalAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	raw := strings.TrimSpace(r.URL.Query().Get("path"))
	relPath, ok := normalizeJournalAssetPath(raw)
	if !ok {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	assetsFolder, err := h.journalAssetsFolder(r.Context())
	if err != nil {
		if errors.Is(err, errJournalsFolderNotFound) {
			http.Error(w, "journals folder not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fullPath := filepath.Join(assetsFolder, filepath.FromSlash(relPath))
	data, err := h.App.FS.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "asset not found", http.StatusNotFound)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(relPath))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// UploadJournalAsset godoc
// @Summary Upload journal asset
// @Tags journals
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Asset file"
// @Param filename query string false "Override filename"
// @Success 201 {object} JournalAssetDTO
// @Router /journals/assets [post]
func (h *Handlers) UploadJournalAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxJournalAssetBytes)
	if err := r.ParseMultipartForm(maxJournalAssetBytes); err != nil {
		http.Error(w, "invalid upload", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := strings.TrimSpace(r.URL.Query().Get("filename"))
	if filename == "" {
		filename = header.Filename
	}
	relPath, ok := normalizeJournalAssetPath(filename)
	if !ok {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	assetsFolder, err := h.journalAssetsFolder(r.Context())
	if err != nil {
		if errors.Is(err, errJournalsFolderNotFound) {
			http.Error(w, "journals folder not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := h.App.FS.MkdirAll(assetsFolder, 0o755); err != nil {
		http.Error(w, "unable to save asset", http.StatusInternalServerError)
		return
	}

	fullPath := filepath.Join(assetsFolder, filepath.FromSlash(relPath))
	writer, err := h.App.FS.Create(fullPath)
	if err != nil {
		http.Error(w, "unable to save asset", http.StatusInternalServerError)
		return
	}
	defer writer.Close()

	written, err := io.Copy(writer, file)
	if err != nil {
		http.Error(w, "unable to save asset", http.StatusInternalServerError)
		return
	}

	if written > maxJournalAssetBytes {
		http.Error(w, "asset too large", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, JournalAssetDTO{
		Path:     buildJournalAssetResponsePath(relPath),
		Filename: filepath.Base(relPath),
		Size:     written,
	})
}

func (h *Handlers) journalAssetsFolder(ctx context.Context) (string, error) {
	folder, ok, err := h.journalsFolder(ctx)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errJournalsFolderNotFound
	}
	return filepath.Join(folder, "assets"), nil
}

func normalizeJournalAssetPath(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	for _, prefix := range []string{"../assets/", "./assets/", "assets/"} {
		if strings.HasPrefix(normalized, prefix) {
			normalized = strings.TrimPrefix(normalized, prefix)
			break
		}
	}
	normalized = strings.TrimPrefix(normalized, "./")
	cleaned := path.Clean(normalized)
	if cleaned == "." || cleaned == "/" || strings.HasPrefix(cleaned, "..") {
		return "", false
	}
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "" {
		return "", false
	}
	return cleaned, true
}

func buildJournalAssetResponsePath(relPath string) string {
	return path.Join("..", "assets", relPath)
}
