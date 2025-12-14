package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"bottomley.ian/musicserver/internal/app"
)

type Handlers struct {
	App *app.App
}

func New(a *app.App) *Handlers {
	return &Handlers{App: a}
}

type Health struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
