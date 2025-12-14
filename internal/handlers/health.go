package handlers

import (
	"net/http"
	"time"
)

// Health godoc
// @Summary Health check
// @Description Returns ok + current server time
// @Tags system
// @Produce json
// @Success 200 {object} handlers.Health
// @Router /health [get]
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, Health{Status: "ok", Time: time.Now()})
}
