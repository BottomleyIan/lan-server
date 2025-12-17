package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"bottomley.ian/musicserver/internal/services/scanner"
)

// ListFolders godoc
// @Summary List folders
// @Description List all non-deleted folders
// @Tags folders
// @Produce json
// @Success 200 {array} handlers.FolderDTO
// @Router /folders [get]
func (h *Handlers) ListFolders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	folders, err := h.App.Queries.ListFolders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, foldersDTOFromDB(folders))
}

type createFolderRequest struct {
	Path string `json:"path"`
}

// CreateFolder godoc
// @Summary Create folder
// @Description Create a folder by path
// @Tags folders
// @Accept json
// @Produce json
// @Param request body handlers.createFolderRequest true "Folder to create"
// @Success 200 {object} handlers.FolderDTO
// @Router /folders [post]
func (h *Handlers) CreateFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body createFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Path == "" {
		http.Error(w, "path required", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.CreateFolder(r.Context(), body.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, folderDTOFromDB(row))
}

// GetFolder godoc
// @Summary Get folder
// @Tags folders
// @Produce json
// @Param id path int true "Folder ID"
// @Success 200 {object} handlers.FolderDTO
// @Router /folders/{id} [get]
func (h *Handlers) GetFolder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}

	folder, err := h.App.Queries.GetFolderByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	writeJSON(w, folderDTOFromDB(folder))
}

// DeleteFolder godoc
// @Summary Delete folder
// @Description Soft-delete a folder root
// @Tags folders
// @Param id path int true "Folder ID"
// @Success 204
// @Router /folders/{id} [delete]
func (h *Handlers) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	_, err = h.App.Queries.SoftDeleteFolder(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "folder not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ScanFolder godoc
// @Summary Trigger folder scan
// @Description Start a scan of a folder root and update indexed tracks
// @Tags folders
// @Param id path int true "Folder ID"
// @Success 202 {object} ScanDTO
// @Router /folders/{id}/scan [post]
func (h *Handlers) ScanFolder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	f, err := h.App.Queries.GetFolderByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "folder not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if f.DeletedAt.Valid {
		http.Error(w, "folder not found", http.StatusNotFound)
		return
	}
	startedAt, err := h.App.Queries.StartFolderScan(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	err = h.Scanner.ScanFolder(ctx, id)
	if err != nil {
		var finishErr error
		switch {
		case errors.Is(err, scanner.ErrFolderUnavailable):
			finishErr = h.App.Queries.FinishFolderScanUnavailable(ctx, err.Error(), id)
		case errors.Is(err, context.Canceled):
			finishErr = h.App.Queries.FinishFolderScanError(ctx, "scan canceled", id)
		default:
			finishErr = h.App.Queries.FinishFolderScanError(ctx, err.Error(), id)
		}
		if finishErr != nil {
			log.Printf("failed to record scan failure for folder %d: %v", id, finishErr)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.App.Queries.FinishFolderScanOK(ctx, id); err != nil {
		log.Printf("failed to record scan success for folder %d: %v", id, err)
		http.Error(w, "failed to record scan result", http.StatusInternalServerError)
		return
	}

	dto := ScanDTO{
		FolderID:  id,
		Status:    "ok",
		StartedAt: startedAt.Time,
	}

	w.WriteHeader(http.StatusAccepted)
	writeJSON(w, dto)
}

// GetScanStatus godoc
// @Summary Get scan status
// @Description Get the most recent scan status for a folder
// @Tags folders
// @Param id path int true "Folder ID"
// @Success 200 {object} ScanDTO
// @Router /folders/{id}/scan [get]
func (h *Handlers) ScanStatus(w http.ResponseWriter, r *http.Request) {
	// implementation later
}
