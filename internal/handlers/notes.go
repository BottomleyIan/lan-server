package handlers

import (
	"errors"
	"net/http"
	"strings"

	"bottomley.ian/musicserver/internal/db"
)

// ListNotes godoc
// @Summary List notes
// @Tags notes
// @Produce json
// @Param year query int false "Filter by year"
// @Param month query int false "Filter by month (1-12)"
// @Param day query int false "Filter by day (1-31)"
// @Param tag query string false "Filter by tag"
// @Success 200 {array} NoteDTO
// @Router /notes [get]
func (h *Handlers) ListNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := h.syncJournalsFromDisk(r.Context()); err != nil {
		if errors.Is(err, errJournalsFolderNotFound) {
			http.Error(w, "journals folder not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	year, ok := parseOptionalIntQueryParam(w, r, "year", 0, 9999)
	if !ok {
		return
	}
	month, ok := parseOptionalIntQueryParam(w, r, "month", 1, 12)
	if !ok {
		return
	}
	day, ok := parseOptionalIntQueryParam(w, r, "day", 1, 31)
	if !ok {
		return
	}

	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	var tagParam interface{}
	if tag != "" {
		tagParam = tag
	}

	rows, err := h.App.Queries.ListNotes(r.Context(), db.ListNotesParams{
		Column1: year,
		Column2: month,
		Column3: day,
		Column4: tagParam,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, notesDTOFromDB(rows))
}
