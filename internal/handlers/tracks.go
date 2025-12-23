package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bottomley.ian/musicserver/internal/db"
	dbtypes "bottomley.ian/musicserver/internal/dbtypes"
	myfs "bottomley.ian/musicserver/internal/services/fs"

	"github.com/go-chi/chi/v5"
)

type updateTrackRequest struct {
	Rating *int64 `json:"rating,omitempty"`
}

type updateTrackRatingRequest struct {
	Rating *int64 `json:"rating,omitempty"`
}

type trackListOptions struct {
	includeAlbum       bool
	includeArtist      bool
	startsWith         *string
	includeUnavailable bool
}

var (
	allowedExpand = map[string]struct{}{
		"album":  {},
		"artist": {},
	}
	allowedExpandList = []string{"album", "artist"}
)

// ListTracks godoc
// @Summary List tracks
// @Description List all non-deleted tracks
// @Tags tracks
// @Produce json
// @Param albumId query int false "Filter by album ID"
// @Param artistId query int false "Filter by artist ID"
// @Param expand query string false "Comma-separated expansions (album,artist); defaults to none" Enums(album,artist) example(album,artist)
// @Param startswith query string false "Prefix filter on filename"
// @Param include_unavailable query bool false "Include tracks from unavailable folders (default: false)"
// @Success 200 {array} TrackDTO
// @Router /tracks [get]
func (h *Handlers) ListTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	opts, err := parseTrackListOptions(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	albumIDStr := r.URL.Query().Get("albumId")
	var albumID *int64
	if albumIDStr != "" {
		id, err := strconv.ParseInt(albumIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid albumId", http.StatusBadRequest)
			return
		}
		albumID = &id
	}

	artistIDStr := r.URL.Query().Get("artistId")
	var artistID *int64
	if artistIDStr != "" {
		id, err := strconv.ParseInt(artistIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid artistId", http.StatusBadRequest)
			return
		}
		artistID = &id
	}

	tracks, err := h.listTracksShared(r.Context(), albumID, artistID, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, filterTracks(tracks, opts))
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

// UpdateTrackRating godoc
// @Summary Update track rating
// @Description Update track rating (1-5) or clear it
// @Tags tracks
// @Accept json
// @Produce json
// @Param id path int true "Track ID"
// @Param request body updateTrackRatingRequest true "Track rating payload"
// @Success 200 {object} TrackDTO
// @Router /tracks/{id}/rating [patch]
func (h *Handlers) UpdateTrackRating(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body updateTrackRatingRequest
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

func (h *Handlers) listTracksShared(ctx context.Context, albumID *int64, artistID *int64, opts trackListOptions) ([]TrackDTO, error) {
	prefix := nullStringFromPrefix(opts.startsWith)
	includeUnavailable := int64(0)
	if opts.includeUnavailable {
		includeUnavailable = 1
	}

	if albumID != nil && artistID != nil {
		if opts.includeAlbum || opts.includeArtist {
			rows, err := h.App.Queries.ListPlayableTracksForAlbumArtist(ctx, db.ListPlayableTracksForAlbumArtistParams{
				Column1:  includeUnavailable,
				AlbumID:  dbtypes.NullInt64{Int64: *albumID, Valid: true},
				ArtistID: dbtypes.NullInt64{Int64: *artistID, Valid: true},
				Column4:  prefix,
			})
			if err != nil {
				return nil, err
			}
			return tracksDTOFromAlbumArtistRows(rows), nil
		}

		rows, err := h.App.Queries.ListPlayableTracksForAlbumArtistBase(ctx, db.ListPlayableTracksForAlbumArtistBaseParams{
			Column1:  includeUnavailable,
			AlbumID:  dbtypes.NullInt64{Int64: *albumID, Valid: true},
			ArtistID: dbtypes.NullInt64{Int64: *artistID, Valid: true},
			Column4:  prefix,
		})
		if err != nil {
			return nil, err
		}
		return tracksDTOFromBase(rows), nil
	}

	if albumID != nil {
		if opts.includeAlbum || opts.includeArtist {
			rows, err := h.App.Queries.ListPlayableTracksForAlbum(ctx, db.ListPlayableTracksForAlbumParams{
				Column1: includeUnavailable,
				AlbumID: dbtypes.NullInt64{Int64: *albumID, Valid: true},
				Column3: prefix,
			})
			if err != nil {
				return nil, err
			}
			return tracksDTOFromAlbumRows(rows), nil
		}

		rows, err := h.App.Queries.ListPlayableTracksForAlbumBase(ctx, db.ListPlayableTracksForAlbumBaseParams{
			Column1: includeUnavailable,
			AlbumID: dbtypes.NullInt64{Int64: *albumID, Valid: true},
			Column3: prefix,
		})
		if err != nil {
			return nil, err
		}
		return tracksDTOFromBase(rows), nil
	}

	if artistID != nil {
		if opts.includeAlbum || opts.includeArtist {
			rows, err := h.App.Queries.ListPlayableTracksForArtist(ctx, db.ListPlayableTracksForArtistParams{
				Column1:  includeUnavailable,
				ArtistID: dbtypes.NullInt64{Int64: *artistID, Valid: true},
				Column3:  prefix,
			})
			if err != nil {
				return nil, err
			}
			return tracksDTOFromArtistRows(rows), nil
		}

		rows, err := h.App.Queries.ListPlayableTracksForArtistBase(ctx, db.ListPlayableTracksForArtistBaseParams{
			Column1:  includeUnavailable,
			ArtistID: dbtypes.NullInt64{Int64: *artistID, Valid: true},
			Column3:  prefix,
		})
		if err != nil {
			return nil, err
		}
		return tracksDTOFromBase(rows), nil
	}

	if opts.includeAlbum || opts.includeArtist {
		rows, err := h.App.Queries.ListPlayableTracksWithJoins(ctx, db.ListPlayableTracksWithJoinsParams{
			Prefix:             prefix,
			IncludeUnavailable: includeUnavailable,
		})
		if err != nil {
			return nil, err
		}
		return tracksDTOFromPlayableRows(rows), nil
	}

	rows, err := h.App.Queries.ListPlayableTracks(ctx, db.ListPlayableTracksParams{
		Prefix:             prefix,
		IncludeUnavailable: includeUnavailable,
	})
	if err != nil {
		return nil, err
	}
	return tracksDTOFromBase(rows), nil
}

func parseTrackListOptions(r *http.Request) (trackListOptions, error) {
	opts := trackListOptions{
		includeAlbum:  false,
		includeArtist: false,
	}

	expandRaw := r.URL.Query().Get("expand")
	startsWith := strings.TrimSpace(r.URL.Query().Get("startswith"))
	includeUnavailableRaw := strings.TrimSpace(r.URL.Query().Get("include_unavailable"))
	if includeUnavailableRaw == "" {
		// Accept camelCase as a fallback.
		includeUnavailableRaw = strings.TrimSpace(r.URL.Query().Get("includeUnavailable"))
	}

	expandSet := parseCSVSet(expandRaw)
	if expandRaw != "" {
		if invalid := diffSet(expandSet, allowedExpand); len(invalid) > 0 {
			return trackListOptions{}, fmt.Errorf("invalid expand value(s): %s; allowed: %s", strings.Join(invalid, ", "), strings.Join(allowedExpandList, ", "))
		}
		opts.includeAlbum = contains(expandSet, "album")
		opts.includeArtist = contains(expandSet, "artist")
	}

	if startsWith != "" {
		opts.startsWith = &startsWith
	}

	if includeUnavailableRaw != "" {
		parsed, err := strconv.ParseBool(includeUnavailableRaw)
		if err != nil {
			return trackListOptions{}, fmt.Errorf("invalid include_unavailable: %v", err)
		}
		opts.includeUnavailable = parsed
	}

	return opts, nil
}

func filterTracks(tracks []TrackDTO, opts trackListOptions) []TrackDTO {
	for i := range tracks {
		if !opts.includeAlbum {
			tracks[i].Album = nil
		}
		if !opts.includeArtist {
			tracks[i].Artist = nil
		}
	}
	return tracks
}

func nullStringFromPrefix(prefix *string) sql.NullString {
	if prefix == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *prefix, Valid: true}
}

func parseCSVSet(input string) map[string]struct{} {
	out := make(map[string]struct{})
	if input == "" {
		return out
	}
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out[trimmed] = struct{}{}
		}
	}
	return out
}

func diffSet(values, allowed map[string]struct{}) []string {
	var invalid []string
	for val := range values {
		if _, ok := allowed[val]; !ok {
			invalid = append(invalid, val)
		}
	}
	return invalid
}

func contains(set map[string]struct{}, key string) bool {
	_, ok := set[key]
	return ok
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

	basePath, err := myfs.ExpandPath(pathParts.FolderPath)
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
