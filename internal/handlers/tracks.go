package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"bottomley.ian/musicserver/internal/db"
	dbtypes "bottomley.ian/musicserver/internal/dbtypes"

	"github.com/go-chi/chi/v5"
)

type updateTrackRequest struct {
	Rating *int64 `json:"rating,omitempty"`
}

// ListTracks godoc
// @Summary List tracks
// @Description List all non-deleted tracks
// @Tags tracks
// @Produce json
// @Success 200 {array} TrackDTO
// @Router /tracks [get]
func (h *Handlers) ListTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tracks, err := h.App.Queries.ListPlayableTracks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, tracksDTOFromDB(tracks))
}

// GetTrack godoc
// @Summary Get track
// @Tags tracks
// @Produce json
// @Param id path int true "Track ID"
// @Success 200 {object} TrackDTO
// @Router /tracks/{id} [get]
func (h *Handlers) GetTrack(w http.ResponseWriter, r *http.Request) {
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

	row, err := h.App.Queries.GetTrackByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, trackDTOFromDB(row))
}

// UpdateTrack godoc
// @Summary Update track
// @Description Update track metadata (rating 1-5 or clear)
// @Tags tracks
// @Accept json
// @Produce json
// @Param id path int true "Track ID"
// @Param request body updateTrackRequest true "Track update payload"
// @Success 200 {object} TrackDTO
// @Router /tracks/{id} [put]
func (h *Handlers) UpdateTrack(w http.ResponseWriter, r *http.Request) {
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

	var body updateTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	var rating dbtypes.NullInt64
	if body.Rating != nil {
		if *body.Rating < 1 || *body.Rating > 5 {
			http.Error(w, "rating must be 1-5", http.StatusBadRequest)
			return
		}
		rating = dbtypes.NullInt64{Int64: *body.Rating, Valid: true}
	}

	row, err := h.App.Queries.UpdateTrackRating(r.Context(), db.UpdateTrackRatingParams{
		Rating: rating,
		ID:     id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, trackDTOFromDB(row))
}
