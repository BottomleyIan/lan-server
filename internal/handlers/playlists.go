package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bottomley.ian/musicserver/internal/db"

	"github.com/go-chi/chi/v5"
)

type createPlaylistRequest struct {
	Name string `json:"name"`
}

// ListPlaylists godoc
// @Summary List playlists
// @Tags playlists
// @Produce json
// @Success 200 {array} PlaylistDTO
// @Router /playlists [get]
func (h *Handlers) ListPlaylists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rows, err := h.App.Queries.ListPlaylists(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, playlistsDTOFromDB(rows))
}

// CreatePlaylist godoc
// @Summary Create playlist
// @Tags playlists
// @Accept json
// @Produce json
// @Param request body createPlaylistRequest true "Playlist to create"
// @Success 200 {object} PlaylistDTO
// @Router /playlists [post]
func (h *Handlers) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body createPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.CreatePlaylist(r.Context(), body.Name)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, playlistDTOFromDB(row))
}

// GetPlaylist godoc
// @Summary Get playlist
// @Tags playlists
// @Produce json
// @Param id path int true "Playlist ID"
// @Success 200 {object} PlaylistDTO
// @Router /playlists/{id} [get]
func (h *Handlers) GetPlaylist(w http.ResponseWriter, r *http.Request) {
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

	row, err := h.App.Queries.GetPlaylistByID(r.Context(), id)
	if err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	writeJSON(w, playlistDTOFromDB(row))
}

// UpdatePlaylist godoc
// @Summary Update playlist
// @Tags playlists
// @Accept json
// @Produce json
// @Param id path int true "Playlist ID"
// @Param request body createPlaylistRequest true "Playlist update payload"
// @Success 200 {object} PlaylistDTO
// @Router /playlists/{id} [put]
func (h *Handlers) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
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

	var body createPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.UpdatePlaylist(r.Context(), db.UpdatePlaylistParams{
		Name: body.Name,
		ID:   id,
	})
	if err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	writeJSON(w, playlistDTOFromDB(row))
}

// DeletePlaylist godoc
// @Summary Delete playlist
// @Tags playlists
// @Param id path int true "Playlist ID"
// @Success 204
// @Router /playlists/{id} [delete]
func (h *Handlers) DeletePlaylist(w http.ResponseWriter, r *http.Request) {
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
	if id <= 0 || id == 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// soft delete playlist
	affected, err := h.App.Queries.SoftDeletePlaylist(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	// clear any tracks; ignore failure, log if needed later
	_ = h.App.Queries.ClearPlaylistTracks(r.Context(), id)

	w.WriteHeader(http.StatusNoContent)
}
