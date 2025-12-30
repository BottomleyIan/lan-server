package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"bottomley.ian/musicserver/internal/db"
)

type createNoteRequest struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Body        *string  `json:"body,omitempty"`
}

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

// CreateNote godoc
// @Summary Append a note entry for today
// @Tags notes
// @Accept json
// @Produce json
// @Param request body createNoteRequest true "Note payload"
// @Success 204
// @Router /notes [post]
func (h *Handlers) CreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	description := strings.TrimSpace(body.Description)
	if description == "" {
		http.Error(w, "description required", http.StatusBadRequest)
		return
	}

	entry := renderJournalEntry(body.Tags, description, body.Body)
	if err := h.appendToTodayJournal(r.Context(), entry); err != nil {
		if errors.Is(err, errJournalsFolderNotFound) {
			http.Error(w, "journals folder not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
