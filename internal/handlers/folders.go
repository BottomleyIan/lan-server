package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

)

// ListFolders godoc
// @Summary List folders
// @Description List all non-deleted folders
// @Tags folders
// @Produce json
// @Success 200 {array} db.Folder
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
	writeJSON(w, folders)
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
// @Success 200 {object} db.Folder
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
	writeJSON(w, row)
}

// GetFolder godoc
// @Summary Get folder
// @Tags folders
// @Produce json
// @Param id path int true "Folder ID"
// @Success 200 {object} db.Folder
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

	writeJSON(w, folder)
}

