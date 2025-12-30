package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"bottomley.ian/musicserver/internal/db"
)

// ListTasks godoc
// @Summary List tasks
// @Tags tasks
// @Produce json
// @Param statuses query []string false "Statuses filter (comma-separated or repeated)"
// @Param status query []string false "Status filter (comma-separated or repeated)"
// @Param tags query []string false "Tags filter (comma-separated or repeated)"
// @Param year query int false "Filter by year"
// @Param month query int false "Filter by month (1-12)"
// @Param day query int false "Filter by day (1-31)"
// @Success 200 {array} TaskDTO
// @Router /tasks [get]
func (h *Handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
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

	filters, ok := parseTasksFilters(w, r)
	if !ok {
		return
	}

	var yearParam interface{}
	if filters.Year != nil {
		yearParam = *filters.Year
	}
	var monthParam interface{}
	if filters.Month != nil {
		monthParam = *filters.Month
	}
	var dayParam interface{}
	if filters.Day != nil {
		dayParam = *filters.Day
	}
	var statusesParam interface{}
	if len(filters.Statuses) > 0 {
		data, err := json.Marshal(filters.Statuses)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		statusesParam = string(data)
	}
	var tagsParam interface{}
	if len(filters.Tags) > 0 {
		data, err := json.Marshal(filters.Tags)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		tagsParam = string(data)
	}

	rows, err := h.App.Queries.ListTasks(r.Context(), db.ListTasksParams{
		Column1: yearParam,
		Column2: monthParam,
		Column3: statusesParam,
		Column4: tagsParam,
		Column5: dayParam,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, tasksDTOFromDB(rows))
}

type taskListFilters struct {
	Year     *int64
	Month    *int64
	Day      *int64
	Statuses []string
	Tags     []string
}

func parseTasksFilters(w http.ResponseWriter, r *http.Request) (taskListFilters, bool) {
	var filters taskListFilters

	year, ok := parseOptionalIntParam(w, r, "year", 0, 9999)
	if !ok {
		return filters, false
	}
	month, ok := parseOptionalIntParam(w, r, "month", 1, 12)
	if !ok {
		return filters, false
	}
	day, ok := parseOptionalIntParam(w, r, "day", 1, 31)
	if !ok {
		return filters, false
	}
	filters.Year = year
	filters.Month = month
	filters.Day = day

	statuses := append(parseQueryList(r, "statuses"), parseQueryList(r, "status")...)
	filters.Statuses = normalizeStatuses(statuses)
	filters.Tags = normalizeTags(parseQueryList(r, "tags"))

	return filters, true
}

func parseOptionalIntParam(w http.ResponseWriter, r *http.Request, key string, min, max int) (*int64, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		http.Error(w, "invalid "+key, http.StatusBadRequest)
		return nil, false
	}
	out := int64(value)
	return &out, true
}

func parseQueryList(r *http.Request, key string) []string {
	values := r.URL.Query()[key]
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			out = append(out, trimmed)
		}
	}
	return out
}

func normalizeStatuses(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized := strings.ToUpper(trimmed)
		for _, status := range expandTaskStatusFilter(normalized) {
			if _, ok := seen[status]; ok {
				continue
			}
			seen[status] = struct{}{}
			out = append(out, status)
		}
	}
	return out
}

func normalizeTags(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func expandTaskStatusFilter(status string) []string {
	switch status {
	case "IN-PROGRESS":
		return []string{"IN-PROGRESS", "DOING", "WAITING"}
	case "TODO":
		return []string{"TODO", "NOW", "LATER"}
	case "DONE":
		return []string{"DONE"}
	case "CANCELLED":
		return []string{"CANCELLED"}
	default:
		return []string{status}
	}
}
