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

type updateAlbumRequest struct {
	ArtistID int64  `json:"artist_id"`
	Title    string `json:"title"`
}

// ListAlbums godoc
// @Summary List albums
// @Tags albums
// @Produce json
// @Param startswith query string false "Prefix filter on title"
// @Success 200 {array} AlbumDTO
// @Router /albums [get]
func (h *Handlers) ListAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	prefix := strings.TrimSpace(r.URL.Query().Get("startswith"))
	var startsWith sql.NullString
	if prefix != "" {
		startsWith = sql.NullString{String: prefix, Valid: true}
	}

	rows, err := h.App.Queries.ListAlbumsWithArtist(r.Context(), startsWith)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, albumsDTOFromRows(rows))
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

	album, err := h.App.Queries.GetAlbumWithArtist(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "album not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, albumDTOFromParts(album.Album, album.Artist))
}

// ListAlbumTracks godoc
// @Summary List tracks for an album
// @Tags albums
// @Produce json
// @Param id path int true "Album ID"
// @Param expand query string false "Comma-separated expansions (album,artist)" Enums(album,artist) example(album,artist)
// @Param startswith query string false "Prefix filter on filename"
// @Success 200 {array} TrackDTO
// @Router /albums/{id}/tracks [get]
func (h *Handlers) ListAlbumTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	opts, err := parseTrackListOptions(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// ensure album exists
	if _, err := h.App.Queries.GetAlbumByID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "album not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	tracks, err := h.listTracksShared(r.Context(), &id, opts)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, filterTracks(tracks, opts))
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

	_, err = h.App.Queries.UpdateAlbum(r.Context(), db.UpdateAlbumParams{
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

	updated, err := h.App.Queries.GetAlbumWithArtist(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, albumDTOFromParts(updated.Album, updated.Artist))
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
