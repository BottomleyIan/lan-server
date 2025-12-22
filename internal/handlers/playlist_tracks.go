package handlers

import (
	"context"
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

	tx, err := h.App.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	queries := h.App.Queries.WithTx(tx)

	trackRow, err := queries.GetTrackWithJoins(r.Context(), body.TrackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			http.Error(w, "track not found", http.StatusNotFound)
		} else {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	existing, err := queries.GetPlaylistTrack(r.Context(), db.GetPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    body.TrackID,
	})
	exists := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	count, err := queries.CountPlaylistTracks(r.Context(), playlistID)
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var desired int64
	if body.Position != nil {
		desired = *body.Position
	} else {
		if exists {
			desired = count - 1
		} else {
			desired = count
		}
	}
	if desired < 0 {
		_ = tx.Rollback()
		http.Error(w, "position must be >= 0", http.StatusBadRequest)
		return
	}

	var row db.PlaylistTrack
	if exists {
		maxPos := count - 1
		if desired > maxPos {
			desired = maxPos
		}
		if desired != existing.Position {
			if desired > existing.Position {
				if err := queries.ShiftPlaylistTrackPositionsDownRange(r.Context(), db.ShiftPlaylistTrackPositionsDownRangeParams{
					PlaylistID: playlistID,
					Position:   existing.Position,
					Position_2: desired,
				}); err != nil {
					_ = tx.Rollback()
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
			} else {
				if err := queries.ShiftPlaylistTrackPositionsUpRange(r.Context(), db.ShiftPlaylistTrackPositionsUpRangeParams{
					PlaylistID: playlistID,
					Position:   desired,
					Position_2: existing.Position,
				}); err != nil {
					_ = tx.Rollback()
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
			}
			row, err = queries.UpdatePlaylistTrackPosition(r.Context(), db.UpdatePlaylistTrackPositionParams{
				Position:   desired,
				PlaylistID: playlistID,
				TrackID:    body.TrackID,
			})
			if err != nil {
				_ = tx.Rollback()
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		} else {
			row = existing
		}
	} else {
		if desired > count {
			desired = count
		}
		if err := queries.ShiftPlaylistTrackPositionsUpFrom(r.Context(), db.ShiftPlaylistTrackPositionsUpFromParams{
			PlaylistID: playlistID,
			Position:   desired,
		}); err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		row, err = queries.AddPlaylistTrack(r.Context(), db.AddPlaylistTrackParams{
			PlaylistID: playlistID,
			TrackID:    body.TrackID,
			Position:   desired,
		})
		if err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	if err := normalizePlaylistPositions(r.Context(), queries, playlistID); err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	row, err = queries.GetPlaylistTrack(r.Context(), db.GetPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    body.TrackID,
	})
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
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

	tx, err := h.App.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	queries := h.App.Queries.WithTx(tx)

	trackRow, err := queries.GetTrackWithJoins(r.Context(), body.TrackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			http.Error(w, "track not found", http.StatusNotFound)
		} else {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	existing, err := queries.GetPlaylistTrack(r.Context(), db.GetPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    body.TrackID,
	})
	exists := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	count, err := queries.CountPlaylistTracks(r.Context(), playlistID)
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var row db.PlaylistTrack
	if exists {
		desired := count - 1
		if desired != existing.Position {
			if err := queries.ShiftPlaylistTrackPositionsDownRange(r.Context(), db.ShiftPlaylistTrackPositionsDownRangeParams{
				PlaylistID: playlistID,
				Position:   existing.Position,
				Position_2: desired,
			}); err != nil {
				_ = tx.Rollback()
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			row, err = queries.UpdatePlaylistTrackPosition(r.Context(), db.UpdatePlaylistTrackPositionParams{
				Position:   desired,
				PlaylistID: playlistID,
				TrackID:    body.TrackID,
			})
			if err != nil {
				_ = tx.Rollback()
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		} else {
			row = existing
		}
	} else {
		row, err = queries.AddPlaylistTrack(r.Context(), db.AddPlaylistTrackParams{
			PlaylistID: playlistID,
			TrackID:    body.TrackID,
			Position:   count,
		})
		if err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	if err := normalizePlaylistPositions(r.Context(), queries, playlistID); err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	row, err = queries.GetPlaylistTrack(r.Context(), db.GetPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    body.TrackID,
	})
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
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

	tx, err := h.App.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	queries := h.App.Queries.WithTx(tx)

	if _, err := queries.GetPlaylistTrack(r.Context(), db.GetPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    trackID,
	}); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "playlist track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	affected, err := queries.DeletePlaylistTrack(r.Context(), db.DeletePlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    trackID,
	})
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		_ = tx.Rollback()
		http.Error(w, "playlist track not found", http.StatusNotFound)
		return
	}

	if err := normalizePlaylistPositions(r.Context(), queries, playlistID); err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
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

	tx, err := h.App.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	queries := h.App.Queries.WithTx(tx)

	existing, err := queries.GetPlaylistTrack(r.Context(), db.GetPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    trackID,
	})
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "playlist track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	count, err := queries.CountPlaylistTracks(r.Context(), playlistID)
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	maxPos := count - 1
	desired := body.Position
	if desired > maxPos {
		desired = maxPos
	}

	row := existing
	if desired != existing.Position {
		if desired > existing.Position {
			if err := queries.ShiftPlaylistTrackPositionsDownRange(r.Context(), db.ShiftPlaylistTrackPositionsDownRangeParams{
				PlaylistID: playlistID,
				Position:   existing.Position,
				Position_2: desired,
			}); err != nil {
				_ = tx.Rollback()
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		} else {
			if err := queries.ShiftPlaylistTrackPositionsUpRange(r.Context(), db.ShiftPlaylistTrackPositionsUpRangeParams{
				PlaylistID: playlistID,
				Position:   desired,
				Position_2: existing.Position,
			}); err != nil {
				_ = tx.Rollback()
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}
		row, err = queries.UpdatePlaylistTrackPosition(r.Context(), db.UpdatePlaylistTrackPositionParams{
			Position:   desired,
			PlaylistID: playlistID,
			TrackID:    trackID,
		})
		if err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	if err := normalizePlaylistPositions(r.Context(), queries, playlistID); err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	row, err = queries.GetPlaylistTrack(r.Context(), db.GetPlaylistTrackParams{
		PlaylistID: playlistID,
		TrackID:    trackID,
	})
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
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

func normalizePlaylistPositions(ctx context.Context, queries *db.Queries, playlistID int64) error {
	ids, err := queries.ListPlaylistTrackIDs(ctx, playlistID)
	if err != nil {
		return err
	}
	for idx, trackID := range ids {
		if err := queries.UpdatePlaylistTrackPositionNoReturn(ctx, db.UpdatePlaylistTrackPositionNoReturnParams{
			Position:   int64(idx),
			PlaylistID: playlistID,
			TrackID:    trackID,
		}); err != nil {
			return err
		}
	}
	return nil
}
