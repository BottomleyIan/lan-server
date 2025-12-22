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

type addPlaylistTrackRequest struct {
	TrackID  int64  `json:"track_id"`
	Position *int64 `json:"position,omitempty"`
}

type enqueuePlaylistTrackRequest struct {
	TrackID int64 `json:"track_id"`
}

type updatePlaylistTrackRequest struct {
	Position int64 `json:"position"`
}

// ClearPlaylist godoc
// @Summary Remove all tracks from a playlist
// @Tags playlists
// @Param id path int true "Playlist ID"
// @Success 204
// @Router /playlists/{id}/clear [post]
func (h *Handlers) ClearPlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playlistID, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	if _, err := h.App.Queries.GetPlaylistByID(r.Context(), playlistID); err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	if err := h.App.Queries.ClearPlaylistTracks(r.Context(), playlistID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListPlaylistTracks godoc
// @Summary List tracks in a playlist
// @Tags playlists
// @Produce json
// @Param id path int true "Playlist ID"
// @Success 200 {array} PlaylistTrackDTO
// @Router /playlists/{id}/tracks [get]
func (h *Handlers) ListPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playlistID, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	// ensure playlist exists
	if _, err := h.App.Queries.GetPlaylistByID(r.Context(), playlistID); err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	rows, err := h.App.Queries.ListPlaylistTracks(r.Context(), playlistID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, playlistTracksDTOFromRows(rows))
}

// AddPlaylistTrack godoc
// @Summary Add a track to a playlist
// @Tags playlists
// @Accept json
// @Produce json
// @Param id path int true "Playlist ID"
// @Param request body addPlaylistTrackRequest true "Track to add"
// @Success 200 {object} PlaylistTrackDTO
// @Router /playlists/{id}/tracks [post]
func (h *Handlers) AddPlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playlistID, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	var body addPlaylistTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.TrackID == 0 {
		http.Error(w, "track_id required", http.StatusBadRequest)
		return
	}

	if _, err := h.App.Queries.GetPlaylistByID(r.Context(), playlistID); err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}
	trackRow, err := h.App.Queries.GetTrackWithJoins(r.Context(), body.TrackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "track not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	position := body.Position
	if position == nil {
		next, err := h.App.Queries.NextPlaylistPosition(r.Context(), playlistID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		position = &next
	}

	row, err := h.App.Queries.AddPlaylistTrack(r.Context(), db.AddPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    body.TrackID,
		Position:   *position,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trackDTO := trackDTOFromJoinedRow(trackRow)
	writeJSON(w, playlistTrackDTOFromPT(row, &trackDTO))
}

// EnqueuePlaylistTrack godoc
// @Summary Enqueue a track at the end of a playlist
// @Tags playlists
// @Accept json
// @Produce json
// @Param id path int true "Playlist ID"
// @Param request body enqueuePlaylistTrackRequest true "Track to enqueue"
// @Success 200 {object} PlaylistTrackDTO
// @Router /playlists/{id}/enqueue [post]
func (h *Handlers) EnqueuePlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playlistID, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	var body enqueuePlaylistTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.TrackID == 0 {
		http.Error(w, "track_id required", http.StatusBadRequest)
		return
	}

	if _, err := h.App.Queries.GetPlaylistByID(r.Context(), playlistID); err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}
	trackRow, err := h.App.Queries.GetTrackWithJoins(r.Context(), body.TrackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "track not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	next, err := h.App.Queries.NextPlaylistPosition(r.Context(), playlistID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	row, err := h.App.Queries.AddPlaylistTrack(r.Context(), db.AddPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    body.TrackID,
		Position:   next,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trackDTO := trackDTOFromJoinedRow(trackRow)
	writeJSON(w, playlistTrackDTOFromPT(row, &trackDTO))
}

// DeletePlaylistTrack godoc
// @Summary Delete a track from a playlist
// @Tags playlists
// @Param id path int true "Playlist ID"
// @Param track_id path int true "Track ID"
// @Success 204
// @Router /playlists/{id}/tracks/{track_id} [delete]
func (h *Handlers) DeletePlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playlistID, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}
	trackID, ok := parseIDParam(w, r, "track_id")
	if !ok {
		return
	}

	if _, err := h.App.Queries.GetPlaylistByID(r.Context(), playlistID); err != nil {
		http.Error(w, "playlist not found", http.StatusNotFound)
		return
	}

	affected, err := h.App.Queries.DeletePlaylistTrack(r.Context(), db.DeletePlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    trackID,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		http.Error(w, "playlist track not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdatePlaylistTrack godoc
// @Summary Update playlist track position
// @Tags playlists
// @Accept json
// @Produce json
// @Param id path int true "Playlist ID"
// @Param track_id path int true "Track ID"
// @Param request body updatePlaylistTrackRequest true "New position"
// @Success 200 {object} PlaylistTrackDTO
// @Router /playlists/{id}/tracks/{track_id} [put]
func (h *Handlers) UpdatePlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playlistID, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}
	trackID, ok := parseIDParam(w, r, "track_id")
	if !ok {
		return
	}

	var body updatePlaylistTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Position < 0 {
		http.Error(w, "position must be >= 0", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.UpdatePlaylistTrackPosition(r.Context(), db.UpdatePlaylistTrackPositionParams{
		Position:   body.Position,
		PlaylistID: playlistID,
		TrackID:    trackID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "playlist track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trackRow, err := h.App.Queries.GetTrackWithJoins(r.Context(), row.TrackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// track missing but playlist track exists; return without track details
			writeJSON(w, playlistTrackDTOFromPT(row, nil))
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trackDTO := trackDTOFromJoinedRow(trackRow)
	writeJSON(w, playlistTrackDTOFromPT(row, &trackDTO))
}

func parseIDParam(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	idStr := chi.URLParam(r, name)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid "+name, http.StatusBadRequest)
		return 0, false
	}
	return id, true
}
