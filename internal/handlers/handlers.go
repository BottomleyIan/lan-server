package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"bottomley.ian/musicserver/internal/app"
	"bottomley.ian/musicserver/internal/services/scanner"
)

type Handlers struct {
	App           *app.App
	Scanner       *scanner.Scanner
	journalSyncMu sync.Mutex
}

func New(a *app.App, s *scanner.Scanner) *Handlers {
	return &Handlers{
		App:     a,
		Scanner: s,
	}
}

type Health struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
