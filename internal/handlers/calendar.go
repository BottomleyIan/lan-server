package handlers

import (
	"errors"
	"net/http"

	"bottomley.ian/musicserver/internal/db"
)

// GetCalendarDay godoc
// @Summary Get calendar day view
// @Tags calendar
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month (1-12)"
// @Param day query int true "Day (1-31)"
// @Success 200 {object} DayViewDTO
// @Router /calendar [get]
func (h *Handlers) GetCalendarDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	year, ok := parseRequiredIntQueryParam(w, r, "year", 0, 9999)
	if !ok {
		return
	}
	month, ok := parseRequiredIntQueryParam(w, r, "month", 1, 12)
	if !ok {
		return
	}
	day, ok := parseRequiredIntQueryParam(w, r, "day", 1, 31)
	if !ok {
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

	dateFilter := buildDateFilter(&year, &month, &day)
	var dateParam interface{}
	if dateFilter != nil {
		dateParam = *dateFilter
	}
	rows, err := h.App.Queries.ListJournalEntries(r.Context(), db.ListJournalEntriesParams{
		Column1: dateParam,
		Column2: nil,
		Column3: nil,
		Column4: nil,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, DayViewDTO{
		Year:    year,
		Month:   month,
		Day:     day,
		Entries: journalEntriesDTOFromDB(rows),
	})
}
