package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

	tracks, err := h.App.Queries.ListPlayableTracksWithJoins(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, tracksDTOFromPlayableRows(tracks))
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

	row, err := h.App.Queries.GetTrackWithJoins(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "track not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, trackDTOFromJoinedRow(row))
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

	_, err = h.App.Queries.UpdateTrackRating(r.Context(), db.UpdateTrackRatingParams{
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

	updated, err := h.App.Queries.GetTrackWithJoins(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, trackDTOFromJoinedRow(updated))
}

// StreamTrack godoc
// @Summary Stream a track
// @Tags tracks
// @Produce application/octet-stream
// @Param id path int true "Track ID"
// @Success 200 {file} file
// @Router /tracks/{id}/play [get]
func (h *Handlers) StreamTrack(w http.ResponseWriter, r *http.Request) {
	h.serveTrackFile(w, r, "inline")
}

// DownloadTrack godoc
// @Summary Download a track
// @Tags tracks
// @Produce application/octet-stream
// @Param id path int true "Track ID"
// @Success 200 {file} file
// @Router /tracks/{id}/download [get]
func (h *Handlers) DownloadTrack(w http.ResponseWriter, r *http.Request) {
	h.serveTrackFile(w, r, "attachment")
}

func (h *Handlers) serveTrackFile(w http.ResponseWriter, r *http.Request, disposition string) {
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

	pathParts, err := h.App.Queries.GetPlayableTrackPathPartsByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "track not found or unavailable", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	basePath, err := expandPath(pathParts.FolderPath)
	if err != nil {
		http.Error(w, "invalid folder path", http.StatusInternalServerError)
		return
	}

	absPath := filepath.Clean(filepath.Join(basePath, pathParts.RelPath))
	log.Printf("serve track id=%d path=%s", id, absPath)
	f, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || info.IsDir() {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	ctype := mime.TypeByExtension(filepath.Ext(info.Name()))
	if ctype == "" {
		ctype = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ctype)
	if disposition != "" {
		w.Header().Set("Content-Disposition", disposition+"; filename=\""+info.Name()+"\"")
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

// expandPath mirrors scanner's expansion to handle ~ and relative paths.
func expandPath(p string) (string, error) {
	p = strings.TrimSpace(p)

	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	}

	p = filepath.Clean(p)
	if !filepath.IsAbs(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		p = abs
	}
	return p, nil
}
