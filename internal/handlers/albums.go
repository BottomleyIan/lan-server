package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"bottomley.ian/musicserver/internal/db"

	"github.com/go-chi/chi/v5"
)

type updateAlbumRequest struct {
	ArtistID int64  `json:"artist_id"`
	Title    string `json:"title"`
}

// ListAlbums godoc
// @Summary List albums
// @Tags albums
// @Produce json
// @Success 200 {array} AlbumDTO
// @Router /albums [get]
func (h *Handlers) ListAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rows, err := h.App.Queries.ListAlbums(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, albumsDTOFromDB(rows))
}

// GetAlbum godoc
// @Summary Get album
// @Tags albums
// @Produce json
// @Param id path int true "Album ID"
// @Success 200 {object} AlbumDTO
// @Router /albums/{id} [get]
func (h *Handlers) GetAlbum(w http.ResponseWriter, r *http.Request) {
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

	album, err := h.App.Queries.GetAlbumByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "album not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, albumDTOFromDB(album))
}

// UpdateAlbum godoc
// @Summary Update album
// @Tags albums
// @Accept json
// @Produce json
// @Param id path int true "Album ID"
// @Param request body updateAlbumRequest true "Album update payload"
// @Success 200 {object} AlbumDTO
// @Router /albums/{id} [put]
func (h *Handlers) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body updateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.ArtistID == 0 || body.Title == "" {
		http.Error(w, "artist_id and title are required", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.UpdateAlbum(r.Context(), db.UpdateAlbumParams{
		ArtistID: body.ArtistID,
		Title:    body.Title,
		ID:       id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "album not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, albumDTOFromDB(row))
}

// DeleteAlbum godoc
// @Summary Delete album
// @Description Soft-delete an album
// @Tags albums
// @Param id path int true "Album ID"
// @Success 204
// @Router /albums/{id} [delete]
func (h *Handlers) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	_, err = h.App.Queries.SoftDeleteAlbum(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "album not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
