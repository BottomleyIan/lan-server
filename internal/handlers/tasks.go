package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
)

// ListTasks godoc
// @Summary List tasks
// @Tags tasks
// @Produce json
// @Success 200 {array} TaskDTO
// @Router /tasks [get]
func (h *Handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
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
		fileYear, fileMonth, fileDay, ok := parseJournalFilename(name)
		if !ok {
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

	rows, err := h.App.Queries.ListTasks(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, tasksDTOFromDB(rows))
}
