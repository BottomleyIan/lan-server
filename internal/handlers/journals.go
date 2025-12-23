package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"bottomley.ian/musicserver/internal/db"

	"github.com/go-chi/chi/v5"
)

var journalFilenameRe = regexp.MustCompile(`^(\d{4})_(\d{2})_(\d{2})\.md$`)
var journalTagRe = regexp.MustCompile(`(?:^|\s)#([A-Za-z0-9/_-]+)`)

// ListJournalsByMonth godoc
// @Summary List journals for a month
// @Tags journals
// @Produce json
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

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		matches := journalFilenameRe.FindStringSubmatch(name)
		if matches == nil {
			continue
		}
		fileYear, _ := strconv.Atoi(matches[1])
		fileMonth, _ := strconv.Atoi(matches[2])
		fileDay, _ := strconv.Atoi(matches[3])
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
		tags := extractJournalTags(string(data))
		tagsJSON, err := json.Marshal(tags)
		if err != nil {
			_ = tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		_, err = queries.UpsertJournal(r.Context(), db.UpsertJournalParams{
			Year:      int64(fileYear),
			Month:     int64(fileMonth),
			Day:       int64(fileDay),
			SizeBytes: info.Size(),
			Hash:      hashHex,
			Tags:      string(tagsJSON),
		})
		if err != nil {
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
