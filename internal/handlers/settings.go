package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"bottomley.ian/musicserver/internal/db"

	"github.com/go-chi/chi/v5"
)

type createSettingRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type updateSettingRequest struct {
	Value string `json:"value"`
}

var settingKeyDefinitions = []SettingKeyDTO{
	{
		Key:         "journals_folder",
		Description: "Filesystem path for journal entries.",
	},
	{
		Key:         "theme",
		Description: "UI theme name.",
	},
}

var settingKeysIndex = func() map[string]string {
	index := make(map[string]string, len(settingKeyDefinitions))
	for _, def := range settingKeyDefinitions {
		index[def.Key] = def.Description
	}
	return index
}()

// ListSettings godoc
// @Summary List settings
// @Tags settings
// @Produce json
// @Success 200 {array} SettingDTO
// @Router /settings [get]
func (h *Handlers) ListSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rows, err := h.App.Queries.ListSettings(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, settingsDTOFromDB(rows))
}

// ListSettingKeys godoc
// @Summary List available setting keys
// @Tags settings
// @Produce json
// @Success 200 {array} SettingKeyDTO
// @Router /settings/keys [get]
func (h *Handlers) ListSettingKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	out := make([]SettingKeyDTO, 0, len(settingKeyDefinitions))
	out = append(out, settingKeyDefinitions...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})

	writeJSON(w, out)
}

// CreateSetting godoc
// @Summary Create setting
// @Tags settings
// @Accept json
// @Produce json
// @Param request body createSettingRequest true "Setting payload"
// @Success 200 {object} SettingDTO
// @Router /settings [post]
func (h *Handlers) CreateSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body createSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	key := strings.TrimSpace(body.Key)
	value := strings.TrimSpace(body.Value)
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}
	if value == "" {
		http.Error(w, "value required", http.StatusBadRequest)
		return
	}
	if _, ok := settingKeysIndex[key]; !ok {
		http.Error(w, "unknown setting key", http.StatusBadRequest)
		return
	}

	if _, err := h.App.Queries.GetSetting(r.Context(), key); err == nil {
		http.Error(w, "setting already exists", http.StatusConflict)
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	row, err := h.App.Queries.CreateSetting(r.Context(), db.CreateSettingParams{
		Key:   key,
		Value: value,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, settingDTOFromDB(row))
}

// GetSetting godoc
// @Summary Get setting
// @Tags settings
// @Produce json
// @Param key path string true "Setting key"
// @Success 200 {object} SettingDTO
// @Router /settings/{key} [get]
func (h *Handlers) GetSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimSpace(chi.URLParam(r, "key"))
	if key == "" {
		http.Error(w, "invalid key", http.StatusBadRequest)
		return
	}
	if _, ok := settingKeysIndex[key]; !ok {
		http.Error(w, "unknown setting key", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.GetSetting(r.Context(), key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "setting not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, settingDTOFromDB(row))
}

// UpdateSetting godoc
// @Summary Update setting
// @Tags settings
// @Accept json
// @Produce json
// @Param key path string true "Setting key"
// @Param request body updateSettingRequest true "Setting payload"
// @Success 200 {object} SettingDTO
// @Router /settings/{key} [put]
func (h *Handlers) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimSpace(chi.URLParam(r, "key"))
	if key == "" {
		http.Error(w, "invalid key", http.StatusBadRequest)
		return
	}
	if _, ok := settingKeysIndex[key]; !ok {
		http.Error(w, "unknown setting key", http.StatusBadRequest)
		return
	}

	var body updateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	value := strings.TrimSpace(body.Value)
	if value == "" {
		http.Error(w, "value required", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.UpdateSetting(r.Context(), db.UpdateSettingParams{
		Key:   key,
		Value: value,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "setting not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, settingDTOFromDB(row))
}

// DeleteSetting godoc
// @Summary Delete setting
// @Tags settings
// @Param key path string true "Setting key"
// @Success 204
// @Router /settings/{key} [delete]
func (h *Handlers) DeleteSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimSpace(chi.URLParam(r, "key"))
	if key == "" {
		http.Error(w, "invalid key", http.StatusBadRequest)
		return
	}
	if _, ok := settingKeysIndex[key]; !ok {
		http.Error(w, "unknown setting key", http.StatusBadRequest)
		return
	}

	affected, err := h.App.Queries.DeleteSetting(r.Context(), key)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		http.Error(w, "setting not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
