package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bottomley.ian/musicserver/internal/db"

	"github.com/go-chi/chi/v5"
)

type createJournalEntryRequest struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Body        *string  `json:"body,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Deadline    *string  `json:"deadline,omitempty"`
	Scheduled   *string  `json:"scheduled,omitempty"`
}

type journalEntryListFilters struct {
	Year     *int64
	Month    *int64
	Day      *int64
	Type     *string
	Statuses []string
	Tags     []string
}

// ListJournalEntries godoc
// @Summary List journal entries
// @Tags journals
// @Produce json
// @Param year query int false "Filter by year"
// @Param month query int false "Filter by month (1-12)"
// @Param day query int false "Filter by day (1-31)"
// @Param type query string false "Entry type (task|misc|note)"
// @Param statuses query []string false "Statuses filter (comma-separated or repeated)"
// @Param status query []string false "Status filter (comma-separated or repeated)"
// @Param tags query []string false "Tags filter (comma-separated or repeated)"
// @Param tag query []string false "Tags filter (comma-separated or repeated)"
// @Success 200 {array} JournalEntryDTO
// @Router /journals/entries [get]
func (h *Handlers) ListJournalEntries(w http.ResponseWriter, r *http.Request) {
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

	filters, ok := parseJournalEntryFilters(w, r)
	if !ok {
		return
	}

	var yearParam interface{}
	dateFilter := buildDateFilter(filters.Year, filters.Month, filters.Day)
	if dateFilter != nil {
		yearParam = *dateFilter
	}
	var typeParam interface{}
	if filters.Type != nil {
		typeParam = *filters.Type
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

	rows, err := h.App.Queries.ListJournalEntries(r.Context(), db.ListJournalEntriesParams{
		Column1: yearParam,
		Column2: typeParam,
		Column3: statusesParam,
		Column4: tagsParam,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, journalEntriesDTOFromDB(rows))
}

// CreateJournalEntry godoc
// @Summary Append a journal entry for today
// @Tags journals
// @Accept json
// @Produce json
// @ID createJournalEntry
// @Param request body createJournalEntryRequest true "Journal entry payload"
// @Success 204
// @Router /journals/entries [post]
func (h *Handlers) CreateJournalEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body createJournalEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	description := strings.TrimSpace(body.Description)
	if description == "" {
		http.Error(w, "description required", http.StatusBadRequest)
		return
	}

	entry, ok := renderEntryFromRequest(w, body, description)
	if !ok {
		return
	}

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

// UpdateJournalEntryByPosition godoc
// @Summary Update journal entry by position
// @Tags journals
// @Accept json
// @Produce json
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param day path int true "Day"
// @Param position path int true "Entry position"
// @Param request body updateJournalEntryRequest true "Updated entry payload"
// @Success 200 {object} JournalEntryDTO
// @Router /journals/entries/{year}/{month}/{day}/{position} [put]
func (h *Handlers) UpdateJournalEntryByPosition(w http.ResponseWriter, r *http.Request) {
	h.updateJournalEntryByPosition(w, r)
}

// DeleteJournalEntryByHash godoc
// @Summary Delete journal entry by hash
// @Tags journals
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param day path int true "Day"
// @Param hash path string true "Entry hash"
// @Success 204
// @Router /journals/entries/{year}/{month}/{day}/{hash} [delete]
func (h *Handlers) DeleteJournalEntryByHash(w http.ResponseWriter, r *http.Request) {
	h.deleteJournalEntryByHash(w, r)
}

// UpdateJournalEntryStatus godoc
// @Summary Update journal entry status by position
// @Tags journals
// @Produce json
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param day path int true "Day"
// @Param position path int true "Entry position"
// @Param status path string true "Task status"
// @Success 200 {object} JournalEntryDTO
// @Router /journals/entries/{year}/{month}/{day}/{position}/{status} [put]
func (h *Handlers) UpdateJournalEntryStatus(w http.ResponseWriter, r *http.Request) {
	h.updateJournalEntryStatus(w, r)
}

func renderEntryFromRequest(w http.ResponseWriter, body createJournalEntryRequest, description string) (string, bool) {
	if body.Status == nil {
		if hasTimestamp(body.Deadline) || hasTimestamp(body.Scheduled) {
			http.Error(w, "status required for scheduled/deadline", http.StatusBadRequest)
			return "", false
		}
		return renderJournalEntry(body.Tags, description, body.Body), true
	}

	status := strings.TrimSpace(*body.Status)
	if status == "" {
		http.Error(w, "status required", http.StatusBadRequest)
		return "", false
	}
	if _, ok := logseqTaskStatusSet[status]; !ok {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return "", false
	}

	return renderTaskEntry(status, body.Tags, description, body.Deadline, body.Scheduled, body.Body), true
}

func hasTimestamp(value *string) bool {
	if value == nil {
		return false
	}
	return strings.TrimSpace(*value) != ""
}

func parseJournalEntryFilters(w http.ResponseWriter, r *http.Request) (journalEntryListFilters, bool) {
	var filters journalEntryListFilters

	year, ok := parseOptionalIntQueryParam(w, r, "year", 0, 9999)
	if !ok {
		return filters, false
	}
	month, ok := parseOptionalIntQueryParam(w, r, "month", 1, 12)
	if !ok {
		return filters, false
	}
	day, ok := parseOptionalIntQueryParam(w, r, "day", 1, 31)
	if !ok {
		return filters, false
	}
	filters.Year = year
	filters.Month = month
	filters.Day = day

	entryType := strings.TrimSpace(r.URL.Query().Get("type"))
	if entryType != "" {
		normalized, ok := normalizeEntryType(entryType)
		if !ok {
			http.Error(w, "invalid type", http.StatusBadRequest)
			return filters, false
		}
		filters.Type = &normalized
	}

	statuses := append(parseQueryList(r, "statuses"), parseQueryList(r, "status")...)
	filters.Statuses = normalizeStatuses(statuses)

	tags := append(parseQueryList(r, "tags"), parseQueryList(r, "tag")...)
	filters.Tags = normalizeTags(tags)

	return filters, true
}

func normalizeEntryType(value string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "task":
		return "task", true
	case "note":
		return "misc", true
	case "misc":
		return "misc", true
	default:
		return "", false
	}
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

func buildDateFilter(year, month, day *int64) *string {
	if year == nil && month == nil && day == nil {
		return nil
	}

	yearPart := "%"
	if year != nil {
		yearPart = fmt.Sprintf("%04d", *year)
	}
	monthPart := "%"
	if month != nil {
		monthPart = fmt.Sprintf("%02d", *month)
	}
	dayPart := "%"
	if day != nil {
		dayPart = fmt.Sprintf("%02d", *day)
	}

	pattern := yearPart + "-" + monthPart + "-" + dayPart
	if !strings.HasSuffix(pattern, "%") {
		pattern += "%"
	}
	return &pattern
}

func (h *Handlers) updateJournalEntryByPosition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	year, month, day, ok := parseYearMonthDayParams(w, r)
	log.Printf("year:%d, month:%d, day: %d", year, month, day)
	if !ok {
		return
	}

	positionStr := strings.TrimSpace(chi.URLParam(r, "position"))
	if positionStr == "" {
		http.Error(w, "position required", http.StatusBadRequest)
		return
	}
	position, err := strconv.Atoi(positionStr)
	if err != nil || position < 0 {
		http.Error(w, "invalid position", http.StatusBadRequest)
		return
	}

	var body updateJournalEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Raw) == "" {
		http.Error(w, "raw required", http.StatusBadRequest)
		return
	}

	//newEntry, ok := parseLogseqEntryBlock(body.Raw)
	//if !ok {
	//	http.Error(w, "invalid entry", http.StatusBadRequest)
	//	return
	//}

	h.journalSyncMu.Lock()
	defer h.journalSyncMu.Unlock()

	folder, ok, err := h.journalsFolder(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if !ok {
		http.Error(w, "journals folder not found", http.StatusNotFound)
		return
	}

	filename := journalFilename(year, month, day)
	fullPath := filepath.Join(folder, filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "journal not found", http.StatusNotFound)
			return
		}
		writeInternalError(w, err)
		return
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	entries := parseLogseqEntries(content)
	start, end, ok := findEntryRangeByPosition(entries, position)
	if !ok {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	newLines := strings.Split(strings.TrimRight(body.Raw, "\n"), "\n")
	updatedLines := append(append([]string{}, lines[:start]...), append(newLines, lines[end+1:]...)...)
	newContent := strings.Join(updatedLines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	if err := h.App.FS.WriteFile(fullPath, []byte(newContent), 0o644); err != nil {
		writeInternalError(w, err)
		return
	}

	entryRow, err := h.syncJournalDayFromContent(r.Context(), year, month, day, []byte(newContent), position)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}
		writeInternalError(w, err)
		return
	}

	writeJSON(w, journalEntryDTOFromDB(entryRow))
}

func (h *Handlers) deleteJournalEntryByHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	year, month, day, ok := parseYearMonthDayParams(w, r)
	if !ok {
		return
	}

	hash := strings.TrimSpace(chi.URLParam(r, "hash"))
	if hash == "" {
		http.Error(w, "hash required", http.StatusBadRequest)
		return
	}

	h.journalSyncMu.Lock()
	defer h.journalSyncMu.Unlock()

	folder, ok, err := h.journalsFolder(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "journals folder not found", http.StatusNotFound)
		return
	}

	filename := journalFilename(year, month, day)
	fullPath := filepath.Join(folder, filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "journal not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	entries := parseLogseqEntries(content)
	start, end, ok := findEntryRange(entries, hash)
	if !ok {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	updatedLines := append([]string{}, lines[:start]...)
	updatedLines = append(updatedLines, lines[end+1:]...)
	newContent := strings.Join(updatedLines, "\n")
	if newContent != "" && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	if err := h.App.FS.WriteFile(fullPath, []byte(newContent), 0o644); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := h.syncJournalDayAfterEdit(r.Context(), year, month, day, []byte(newContent)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) updateJournalEntryStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	year, month, day, ok := parseYearMonthDayParams(w, r)
	log.Printf("year:%d, month:%d, day: %d", year, month, day)
	if !ok {
		return
	}

	positionStr := strings.TrimSpace(chi.URLParam(r, "position"))
	log.Printf("pos:%s", positionStr)
	if positionStr == "" {
		http.Error(w, "position required", http.StatusBadRequest)
		return
	}
	position, err := strconv.Atoi(positionStr)
	if err != nil || position < 0 {
		http.Error(w, "invalid position", http.StatusBadRequest)
		return
	}

	status := strings.TrimSpace(chi.URLParam(r, "status"))
	if status == "" {
		http.Error(w, "status required", http.StatusBadRequest)
		return
	}
	if _, ok := logseqTaskStatusSet[status]; !ok {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	h.journalSyncMu.Lock()
	defer h.journalSyncMu.Unlock()

	folder, ok, err := h.journalsFolder(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if !ok {
		http.Error(w, "journals folder not found", http.StatusNotFound)
		return
	}

	filename := journalFilename(year, month, day)
	fullPath := filepath.Join(folder, filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "journal not found", http.StatusNotFound)
			return
		}
		writeInternalError(w, err)
		return
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	entries := parseLogseqEntries(content)
	start, _, ok := findEntryRangeByPosition(entries, position)
	log.Printf("start: %d", start)
	if !ok {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}
	entry := entries[position]
	if entry.Type != "task" {
		http.Error(w, "task entry required", http.StatusBadRequest)
		return
	}

	tagText := formatLogseqTags(entry.Tags)
	log.Printf("tagText: %s", tagText)
	line := "- " + status + " "
	log.Printf("line: %s", line)
	if tagText != "" {
		line += tagText + " "
	}
	line += entry.Title
	log.Printf("line: %s", line)
	lines[start] = line

	newContent := strings.Join(lines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	if err := h.App.FS.WriteFile(fullPath, []byte(newContent), 0o644); err != nil {
		writeInternalError(w, err)
		return
	}

	entryRow, err := h.syncJournalDayFromContent(r.Context(), year, month, day, []byte(newContent), position)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}
		writeInternalError(w, err)
		return
	}

	writeJSON(w, journalEntryDTOFromDB(entryRow))
}

func (h *Handlers) syncJournalDayFromContent(ctx context.Context, year, month, day int, data []byte, position int) (db.JournalEntry, error) {
	if err := h.syncJournalDayAfterEdit(ctx, year, month, day, data); err != nil {
		return db.JournalEntry{}, err
	}
	return h.App.Queries.GetJournalEntryByDatePosition(ctx, db.GetJournalEntryByDatePositionParams{
		Year:     int64(year),
		Month:    int64(month),
		Day:      int64(day),
		Position: int64(position),
	})
}

func (h *Handlers) syncJournalDayAfterEdit(ctx context.Context, year, month, day int, data []byte) error {
	tx, err := h.App.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	queries := h.App.Queries.WithTx(tx)
	hashBytes := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hashBytes[:])
	if err := syncJournalFromFile(ctx, queries, year, month, day, int64(len(data)), hashHex, data); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func findEntryRange(entries []logseqEntry, hash string) (int, int, bool) {
	lineIndex := 0
	for _, entry := range entries {
		start := lineIndex
		lineIndex++
		if entry.Body != "" {
			bodyLines := strings.Split(entry.Body, "\n")
			if extra := len(bodyLines) - 1; extra > 0 {
				lineIndex += extra
			}
		}
		end := lineIndex - 1

		if entry.Hash == hash {
			return start, end, true
		}
	}
	return 0, 0, false
}

func findEntryRangeByPosition(entries []logseqEntry, position int) (int, int, bool) {
	if position < 0 || position >= len(entries) {
		return 0, 0, false
	}
	lineIndex := 0
	for idx, entry := range entries {
		start := lineIndex
		lineIndex++
		if entry.Body != "" {
			bodyLines := strings.Split(entry.Body, "\n")
			if extra := len(bodyLines) - 1; extra > 0 {
				lineIndex += extra
			}
		}
		end := lineIndex - 1
		if idx == position {
			return start, end, true
		}
	}
	return 0, 0, false
}

func writeInternalError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
