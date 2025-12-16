package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"bottomley.ian/musicserver/internal/db"

	"github.com/go-chi/chi/v5"
)

type updateArtistRequest struct {
	Name string `json:"name"`
}

// ListArtists godoc
// @Summary List artists
// @Tags artists
// @Produce json
// @Param startswith query string false "Prefix filter on name"
// @Success 200 {array} ArtistDTO
// @Router /artists [get]
func (h *Handlers) ListArtists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	prefix := strings.TrimSpace(r.URL.Query().Get("startswith"))
	var startsWith sql.NullString
	if prefix != "" {
		startsWith = sql.NullString{String: prefix, Valid: true}
	}

	rows, err := h.App.Queries.ListArtists(r.Context(), startsWith)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, artistsDTOFromDB(rows))
}

// GetArtist godoc
// @Summary Get artist
// @Tags artists
// @Produce json
// @Param id path int true "Artist ID"
// @Success 200 {object} ArtistDTO
// @Router /artists/{id} [get]
func (h *Handlers) GetArtist(w http.ResponseWriter, r *http.Request) {
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

	artist, err := h.App.Queries.GetArtistByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "artist not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, artistDTOFromDB(artist))
}

// UpdateArtist godoc
// @Summary Update artist
// @Tags artists
// @Accept json
// @Produce json
// @Param id path int true "Artist ID"
// @Param request body updateArtistRequest true "Artist update payload"
// @Success 200 {object} ArtistDTO
// @Router /artists/{id} [put]
func (h *Handlers) UpdateArtist(w http.ResponseWriter, r *http.Request) {
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

	var body updateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.UpdateArtist(r.Context(), db.UpdateArtistParams{
		Name: body.Name,
		ID:   id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "artist not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, artistDTOFromDB(row))
}

// DeleteArtist godoc
// @Summary Delete artist
// @Description Soft-delete an artist
// @Tags artists
// @Param id path int true "Artist ID"
// @Success 204
// @Router /artists/{id} [delete]
func (h *Handlers) DeleteArtist(w http.ResponseWriter, r *http.Request) {
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

	_, err = h.App.Queries.SoftDeleteArtist(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "artist not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
