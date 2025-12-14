package handlers

import (
	"net/http"
)

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
