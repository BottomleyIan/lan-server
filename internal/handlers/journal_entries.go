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
	"strings"

	"bottomley.ian/musicserver/internal/db"

	"github.com/go-chi/chi/v5"
)

// UpdateTaskByHash godoc
// @Summary Update task by hash
// @Tags tasks
// @Accept json
// @Produce json
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param day path int true "Day"
// @Param hash path string true "Entry hash"
// @Param If-Match header string true "Entry hash for concurrency"
// @Param request body updateJournalEntryRequest true "Updated entry payload"
// @Success 200 {object} TaskDTO
// @Router /tasks/{year}/{month}/{day}/{hash} [put]
func (h *Handlers) UpdateTaskByHash(w http.ResponseWriter, r *http.Request) {
	h.updateJournalEntryByHash(w, r, "task")
}

// DeleteTaskByHash godoc
// @Summary Delete task by hash
// @Tags tasks
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param day path int true "Day"
// @Param hash path string true "Entry hash"
// @Success 204
// @Router /tasks/{year}/{month}/{day}/{hash} [delete]
func (h *Handlers) DeleteTaskByHash(w http.ResponseWriter, r *http.Request) {
	h.deleteJournalEntryByHash(w, r, "task")
}

// UpdateNoteByHash godoc
// @Summary Update note by hash
// @Tags notes
// @Accept json
// @Produce json
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param day path int true "Day"
// @Param hash path string true "Entry hash"
// @Param If-Match header string true "Entry hash for concurrency"
// @Param request body updateJournalEntryRequest true "Updated entry payload"
// @Success 200 {object} NoteDTO
// @Router /notes/{year}/{month}/{day}/{hash} [put]
func (h *Handlers) UpdateNoteByHash(w http.ResponseWriter, r *http.Request) {
	h.updateJournalEntryByHash(w, r, "misc")
}

// DeleteNoteByHash godoc
// @Summary Delete note by hash
// @Tags notes
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param day path int true "Day"
// @Param hash path string true "Entry hash"
// @Success 204
// @Router /notes/{year}/{month}/{day}/{hash} [delete]
func (h *Handlers) DeleteNoteByHash(w http.ResponseWriter, r *http.Request) {
	h.deleteJournalEntryByHash(w, r, "misc")
}

func (h *Handlers) updateJournalEntryByHash(w http.ResponseWriter, r *http.Request, entryType string) {
	if r.Method != http.MethodPut {
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

	match := strings.TrimSpace(r.Header.Get("If-Match"))
	if match == "" {
		http.Error(w, "If-Match required", http.StatusBadRequest)
		return
	}
	if match != hash {
		http.Error(w, "hash mismatch", http.StatusPreconditionFailed)
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

	newEntry, ok := parseLogseqEntryBlock(body.Raw)
	if !ok {
		http.Error(w, "invalid entry", http.StatusBadRequest)
		return
	}
	if entryType == "task" && newEntry.Type != "task" {
		http.Error(w, "task entry required", http.StatusBadRequest)
		return
	}
	if entryType == "misc" && newEntry.Type == "task" {
		http.Error(w, "note entry required", http.StatusBadRequest)
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
	start, end, ok := findEntryRange(entries, hash, entryType)
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
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	entryRow, err := h.syncJournalDayFromContent(r.Context(), year, month, day, []byte(newContent), newEntry.Hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if entryType == "task" {
		writeJSON(w, taskDTOFromDB(entryRow))
		return
	}
	writeJSON(w, noteDTOFromDB(entryRow))
}

func (h *Handlers) deleteJournalEntryByHash(w http.ResponseWriter, r *http.Request, entryType string) {
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
	start, end, ok := findEntryRange(entries, hash, entryType)
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

func (h *Handlers) syncJournalDayFromContent(ctx context.Context, year, month, day int, data []byte, hash string) (db.JournalEntry, error) {
	if err := h.syncJournalDayAfterEdit(ctx, year, month, day, data); err != nil {
		return db.JournalEntry{}, err
	}
	return h.App.Queries.GetJournalEntryByDateHash(ctx, db.GetJournalEntryByDateHashParams{
		Year:  int64(year),
		Month: int64(month),
		Day:   int64(day),
		Hash:  hash,
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

func findEntryRange(entries []logseqEntry, hash string, entryType string) (int, int, bool) {
	lineIndex := 0
	for _, entry := range entries {
		start := lineIndex
		lineIndex++
		if entry.Body != "" {
			bodyLines := strings.Split(entry.Body, "\n")
			lineIndex += len(bodyLines)
		}
		end := lineIndex - 1

		if entry.Hash == hash {
			if entryType == "task" && entry.Type != "task" {
				return 0, 0, false
			}
			if entryType == "misc" && entry.Type == "task" {
				return 0, 0, false
			}
			return start, end, true
		}
	}
	return 0, 0, false
}
