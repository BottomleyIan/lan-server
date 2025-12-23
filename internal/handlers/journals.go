package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"bottomley.ian/musicserver/internal/db"
	"bottomley.ian/musicserver/internal/dbtypes"
	myfs "bottomley.ian/musicserver/internal/services/fs"

	"github.com/go-chi/chi/v5"
)

var journalFilenameRe = regexp.MustCompile(`^(\d{4})_(\d{2})_(\d{2})\.md$`)
var journalTagRe = regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
var logseqTaskStatusSet = map[string]struct{}{
	"LATER":       {},
	"NOW":         {},
	"DONE":        {},
	"TODO":        {},
	"DOING":       {},
	"CANCELLED":   {},
	"IN-PROGRESS": {},
	"WAITING":     {},
}

type createJournalEntryRequest struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Body        *string  `json:"body,omitempty"`
}

type createLogseqTaskRequest struct {
	Status    string   `json:"status" example:"TODO"`
	Tags      []string `json:"tags,omitempty" example:"tag1,tag2"`
	Desc      string   `json:"description" example:"Write release notes"`
	Deadline  *string  `json:"deadline,omitempty" example:"2025-12-23 Tue 17:00"`
	Scheduled *string  `json:"scheduled,omitempty" example:"2025-12-23 Tue 11:00"`
	Body      *string  `json:"body,omitempty" example:"Include the new journals endpoints."`
}

// ListJournalsByMonth godoc
// @Summary List journals for a month
// @Tags journals
// @Produce json
// @Param refresh query bool false "Force refresh from disk"
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Success 200 {array} JournalDTO
// @Router /journals/{year}/{month} [get]
func (h *Handlers) ListJournalsByMonth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	year, month, ok := parseYearMonthParams(w, r)
	if !ok {
		return
	}

	folder, ok, err := h.journalsFolder(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "journals folder not found", http.StatusNotFound)
		return
	}

	refresh := strings.EqualFold(r.URL.Query().Get("refresh"), "true")
	entries, err := os.ReadDir(folder)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	tx, err := h.App.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	queries := h.App.Queries.WithTx(tx)

	if refresh {
		if err := queries.DeleteJournalsByMonth(r.Context(), db.DeleteJournalsByMonthParams{
			Year:  int64(year),
			Month: int64(month),
		}); err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := queries.DeleteTasksByMonth(r.Context(), db.DeleteTasksByMonthParams{
			Year:  int64(year),
			Month: int64(month),
		}); err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		fileYear, fileMonth, fileDay, ok := parseJournalFilename(name)
		if !ok {
			continue
		}
		if fileYear != year || fileMonth != month {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		fullPath := filepath.Join(folder, name)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		hash := sha256.Sum256(data)
		hashHex := hex.EncodeToString(hash[:])
		if err := syncJournalFromFile(r.Context(), queries, fileYear, fileMonth, fileDay, info.Size(), hashHex, data); err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rows, err := h.App.Queries.ListJournalsByMonth(r.Context(), db.ListJournalsByMonthParams{
		Year:  int64(year),
		Month: int64(month),
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, journalsDTOFromDB(rows))
}

// CreateJournalEntry godoc
// @Summary Append a journal entry for today
// @Tags journals
// @Accept json
// @Produce json
// @Param request body createJournalEntryRequest true "Journal entry payload"
// @Success 204
// @Router /journals [post]
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

// GetJournalDay godoc
// @Summary Get journal entry for a day
// @Tags journals
// @Produce json
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param day path int true "Day"
// @Success 200 {object} JournalDayDTO
// @Router /journals/{year}/{month}/{day} [get]
func (h *Handlers) GetJournalDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	year, month, day, ok := parseYearMonthDayParams(w, r)
	if !ok {
		return
	}

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

	writeJSON(w, JournalDayDTO{
		Year:  int64(year),
		Month: int64(month),
		Day:   int64(day),
		Raw:   string(data),
	})
}

func (h *Handlers) journalsFolder(ctx context.Context) (string, bool, error) {
	setting, err := h.App.Queries.GetSetting(ctx, settingKeyJournalsFolder)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}

	path := strings.TrimSpace(setting.Value)
	if path == "" {
		return "", false, nil
	}
	expanded, err := myfs.ExpandUserPath(path)
	if err != nil {
		return "", false, err
	}
	path = expanded
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	if !info.IsDir() {
		return "", false, nil
	}
	return path, true, nil
}

// CreateTask godoc
// @Summary Append a task entry for today
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body createLogseqTaskRequest true "Task payload"
// @Success 204
// @Router /tasks [post]
func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body createLogseqTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	status := strings.TrimSpace(body.Status)
	if status == "" {
		http.Error(w, "status required", http.StatusBadRequest)
		return
	}
	if _, ok := logseqTaskStatusSet[status]; !ok {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	description := strings.TrimSpace(body.Desc)
	if description == "" {
		http.Error(w, "description required", http.StatusBadRequest)
		return
	}

	entry := renderTaskEntry(status, body.Tags, description, body.Deadline, body.Scheduled, body.Body)
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

func parseYearMonthParams(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	yearStr := chi.URLParam(r, "year")
	monthStr := chi.URLParam(r, "month")
	if yearStr == "" || monthStr == "" {
		http.Error(w, "invalid year or month", http.StatusBadRequest)
		return 0, 0, false
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 0 {
		http.Error(w, "invalid year", http.StatusBadRequest)
		return 0, 0, false
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		http.Error(w, "invalid month", http.StatusBadRequest)
		return 0, 0, false
	}
	return year, month, true
}

func parseYearMonthDayParams(w http.ResponseWriter, r *http.Request) (int, int, int, bool) {
	year, month, ok := parseYearMonthParams(w, r)
	if !ok {
		return 0, 0, 0, false
	}
	dayStr := chi.URLParam(r, "day")
	day, err := strconv.Atoi(dayStr)
	if err != nil || day < 1 || day > 31 {
		http.Error(w, "invalid day", http.StatusBadRequest)
		return 0, 0, 0, false
	}
	return year, month, day, true
}

func parseJournalFilename(name string) (int, int, int, bool) {
	matches := journalFilenameRe.FindStringSubmatch(name)
	if matches == nil {
		return 0, 0, 0, false
	}
	year, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, false
	}
	month, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, false
	}
	day, err := strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, false
	}
	return year, month, day, true
}

func journalFilename(year, month, day int) string {
	return strconv.Itoa(year) + "_" + pad2(month) + "_" + pad2(day) + ".md"
}

func pad2(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}

func extractJournalTags(content string) []string {
	matches := journalTagRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	unique := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		tag := strings.TrimSpace(match[1])
		if tag == "" {
			continue
		}
		unique[tag] = struct{}{}
	}
	if len(unique) == 0 {
		return nil
	}
	out := make([]string, 0, len(unique))
	for tag := range unique {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

type logseqTask struct {
	Title       string
	Body        string
	Status      string
	ScheduledAt string
	DeadlineAt  string
}

var errJournalsFolderNotFound = errors.New("journals folder not found")

func syncJournalFromFile(ctx context.Context, queries *db.Queries, year, month, day int, sizeBytes int64, hashHex string, data []byte) error {
	existing, err := queries.GetJournalByDate(ctx, db.GetJournalByDateParams{
		Year:  int64(year),
		Month: int64(month),
		Day:   int64(day),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil && existing.SizeBytes == sizeBytes && existing.Hash == hashHex {
		return queries.UpdateJournalLastChecked(ctx, db.UpdateJournalLastCheckedParams{
			Year:  int64(year),
			Month: int64(month),
			Day:   int64(day),
		})
	}

	tags := extractJournalTags(string(data))
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}

	if _, err := queries.UpsertJournal(ctx, db.UpsertJournalParams{
		Year:      int64(year),
		Month:     int64(month),
		Day:       int64(day),
		SizeBytes: sizeBytes,
		Hash:      hashHex,
		Tags:      string(tagsJSON),
	}); err != nil {
		return err
	}

	if err := queries.DeleteTasksByDate(ctx, db.DeleteTasksByDateParams{
		Year:  int64(year),
		Month: int64(month),
		Day:   int64(day),
	}); err != nil {
		return err
	}

	tasks := parseLogseqTasks(string(data))
	for idx, task := range tasks {
		title := strings.TrimSpace(task.Title)
		if title == "" {
			continue
		}
		body := nullStringFromString(strings.TrimRight(task.Body, "\n"))
		scheduled := nullStringFromString(task.ScheduledAt)
		deadline := nullStringFromString(task.DeadlineAt)
		if _, err := queries.CreateTask(ctx, db.CreateTaskParams{
			Year:        int64(year),
			Month:       int64(month),
			Day:         int64(day),
			Position:    int64(idx),
			Title:       title,
			Body:        body,
			Status:      task.Status,
			ScheduledAt: scheduled,
			DeadlineAt:  deadline,
		}); err != nil {
			return err
		}
	}

	return nil
}

func parseLogseqTasks(content string) []logseqTask {
	lines := strings.Split(content, "\n")
	tasks := make([]logseqTask, 0)
	var current *logseqTask
	var bodyLines []string

	flush := func() {
		if current == nil {
			return
		}
		current.Body = strings.Join(bodyLines, "\n")
		tasks = append(tasks, *current)
		current = nil
		bodyLines = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			flush()
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if rest == "" {
				continue
			}
			parts := strings.Fields(rest)
			if len(parts) == 0 {
				continue
			}
			status := parts[0]
			if _, ok := logseqTaskStatusSet[status]; !ok {
				continue
			}
			title := strings.TrimSpace(strings.TrimPrefix(rest, status))
			title = strings.TrimSpace(stripLogseqTags(title))
			current = &logseqTask{
				Title:  title,
				Status: status,
			}
			bodyLines = []string{}
			continue
		}

		if current != nil {
			bodyLines = append(bodyLines, line)
			if current.ScheduledAt == "" && strings.HasPrefix(trimmed, "SCHEDULED:") {
				current.ScheduledAt = parseLogseqTimestamp(trimmed, "SCHEDULED:")
			}
			if current.DeadlineAt == "" && strings.HasPrefix(trimmed, "DEADLINE:") {
				current.DeadlineAt = parseLogseqTimestamp(trimmed, "DEADLINE:")
			}
		}
	}

	flush()
	return tasks
}

func stripLogseqTags(value string) string {
	return journalTagRe.ReplaceAllString(value, "")
}

func parseLogseqTimestamp(line, prefix string) string {
	value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">") {
		value = strings.TrimSuffix(strings.TrimPrefix(value, "<"), ">")
	}
	return strings.TrimSpace(value)
}

func nullStringFromString(value string) dbtypes.NullString {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return dbtypes.NullString{}
	}
	return dbtypes.NullString{String: trimmed, Valid: true}
}

func renderJournalEntry(tags []string, description string, body *string) string {
	tagText := formatLogseqTags(tags)
	line := "- "
	if tagText != "" {
		line += tagText + " "
	}
	line += description
	lines := []string{line}
	if body != nil && strings.TrimSpace(*body) != "" {
		lines = append(lines, indentBody(*body)...)
	}
	return strings.Join(lines, "\n")
}

func renderTaskEntry(status string, tags []string, description string, deadline *string, scheduled *string, body *string) string {
	tagText := formatLogseqTags(tags)
	line := "- " + status + " "
	if tagText != "" {
		line += tagText + " "
	}
	line += description

	lines := []string{line}
	if deadline != nil && strings.TrimSpace(*deadline) != "" {
		lines = append(lines, "  DEADLINE: "+formatLogseqTimestamp(*deadline))
	}
	if scheduled != nil && strings.TrimSpace(*scheduled) != "" {
		lines = append(lines, "  SCHEDULED: "+formatLogseqTimestamp(*scheduled))
	}
	if body != nil && strings.TrimSpace(*body) != "" {
		lines = append(lines, indentBody(*body)...)
	}
	return strings.Join(lines, "\n")
}

func formatLogseqTags(tags []string) string {
	seen := make(map[string]struct{}, len(tags))
	var out strings.Builder
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out.WriteString("[[")
		out.WriteString(tag)
		out.WriteString("]]")
	}
	return out.String()
}

func indentBody(body string) []string {
	lines := strings.Split(body, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, "  "+line)
	}
	return out
}

func formatLogseqTimestamp(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
		return trimmed
	}
	return "<" + trimmed + ">"
}

func (h *Handlers) appendToTodayJournal(ctx context.Context, entry string) error {
	folder, ok, err := h.journalsFolder(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errJournalsFolderNotFound
	}

	now := time.Now()
	filename := journalFilename(now.Year(), int(now.Month()), now.Day())
	fullPath := filepath.Join(folder, filename)

	if err := h.App.FS.MkdirAll(folder, 0o755); err != nil {
		return err
	}

	if err := ensureTrailingNewline(fullPath); err != nil {
		return err
	}

	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(entry + "\n"); err != nil {
		return err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])

	tx, err := h.App.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	queries := h.App.Queries.WithTx(tx)
	if err := syncJournalFromFile(ctx, queries, now.Year(), int(now.Month()), now.Day(), info.Size(), hashHex, data); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func ensureTrailingNewline(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Size() == 0 {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Seek(-1, io.SeekEnd); err != nil {
		return err
	}
	buf := make([]byte, 1)
	if _, err := f.Read(buf); err != nil {
		return err
	}
	if buf[0] == '\n' {
		return nil
	}

	af, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer af.Close()
	_, err = af.WriteString("\n")
	return err
}
