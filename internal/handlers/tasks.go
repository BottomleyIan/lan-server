package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"bottomley.ian/musicserver/internal/db"
	"bottomley.ian/musicserver/internal/dbtypes"
)

type createTaskRequest struct {
	Title      string   `json:"title"`
	Body       *string  `json:"body,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	StatusCode *string  `json:"status_code,omitempty"`
	Reason     *string  `json:"reason,omitempty"`
}

type updateTaskRequest struct {
	Title string   `json:"title"`
	Body  *string  `json:"body,omitempty"`
	Tags  []string `json:"tags,omitempty"`
}

type createTaskTransitionRequest struct {
	StatusCode string  `json:"status_code"`
	Reason     *string `json:"reason,omitempty"`
}

// ListTaskStatuses godoc
// @Summary List task statuses
// @Tags tasks
// @Produce json
// @Success 200 {array} TaskStatusDTO
// @Router /tasks/statuses [get]
func (h *Handlers) ListTaskStatuses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rows, err := h.App.Queries.ListTaskStatuses(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, taskStatusesDTOFromDB(rows))
}

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

	rows, err := h.App.Queries.ListTasks(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, tasksDTOFromDB(rows))
}

// CreateTask godoc
// @Summary Create task
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body createTaskRequest true "Task to create"
// @Success 200 {object} TaskDTO
// @Router /tasks [post]
func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	statusCode := "BACKLOG"
	if body.StatusCode != nil && strings.TrimSpace(*body.StatusCode) != "" {
		statusCode = strings.TrimSpace(*body.StatusCode)
	}

	var taskBody dbtypes.NullString
	if body.Body != nil {
		taskBody = dbtypes.NullString{String: *body.Body, Valid: true}
	}

	tagsValue, err := tagsToNullString(body.Tags)
	if err != nil {
		http.Error(w, "invalid tags", http.StatusBadRequest)
		return
	}

	var reason dbtypes.NullString
	if body.Reason != nil {
		reason = dbtypes.NullString{String: *body.Reason, Valid: true}
	}

	tx, err := h.App.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	q := h.App.Queries.WithTx(tx)

	row, err := q.CreateTask(r.Context(), db.CreateTaskParams{
		Title:      title,
		Body:       taskBody,
		Tags:       tagsValue,
		StatusCode: statusCode,
	})
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, err = q.AddTaskTransition(r.Context(), db.AddTaskTransitionParams{
		TaskID:     row.ID,
		StatusCode: row.StatusCode,
		Reason:     reason,
	})
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, taskDTOFromDB(row))
}

// GetTask godoc
// @Summary Get task
// @Tags tasks
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} TaskDTO
// @Router /tasks/{id} [get]
func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	row, err := h.App.Queries.GetTaskByID(r.Context(), id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	writeJSON(w, taskDTOFromDB(row))
}

// UpdateTask godoc
// @Summary Update task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body updateTaskRequest true "Task update payload"
// @Success 200 {object} TaskDTO
// @Router /tasks/{id} [put]
func (h *Handlers) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	var body updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	var taskBody dbtypes.NullString
	if body.Body != nil {
		taskBody = dbtypes.NullString{String: *body.Body, Valid: true}
	}

	tagsValue, err := tagsToNullString(body.Tags)
	if err != nil {
		http.Error(w, "invalid tags", http.StatusBadRequest)
		return
	}

	row, err := h.App.Queries.UpdateTask(r.Context(), db.UpdateTaskParams{
		Title: title,
		Body:  taskBody,
		Tags:  tagsValue,
		ID:    id,
	})
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	writeJSON(w, taskDTOFromDB(row))
}

// DeleteTask godoc
// @Summary Delete task
// @Tags tasks
// @Param id path int true "Task ID"
// @Success 204
// @Router /tasks/{id} [delete]
func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	affected, err := h.App.Queries.SoftDeleteTask(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListTaskTransitions godoc
// @Summary List task transitions
// @Tags tasks
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {array} TaskTransitionDTO
// @Router /tasks/{id}/transitions [get]
func (h *Handlers) ListTaskTransitions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	if _, err := h.App.Queries.GetTaskByID(r.Context(), id); err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	rows, err := h.App.Queries.ListTaskTransitions(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, taskTransitionsDTOFromDB(rows))
}

// CreateTaskTransition godoc
// @Summary Add task transition
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body createTaskTransitionRequest true "Transition payload"
// @Success 200 {object} TaskDTO
// @Router /tasks/{id}/transitions [post]
func (h *Handlers) CreateTaskTransition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	var body createTaskTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	statusCode := strings.TrimSpace(body.StatusCode)
	if statusCode == "" {
		http.Error(w, "status_code required", http.StatusBadRequest)
		return
	}

	var reason dbtypes.NullString
	if body.Reason != nil {
		reason = dbtypes.NullString{String: *body.Reason, Valid: true}
	}

	tx, err := h.App.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	q := h.App.Queries.WithTx(tx)

	row, err := q.UpdateTaskStatus(r.Context(), db.UpdateTaskStatusParams{
		StatusCode: statusCode,
		ID:         id,
	})
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, err = q.AddTaskTransition(r.Context(), db.AddTaskTransitionParams{
		TaskID:     row.ID,
		StatusCode: statusCode,
		Reason:     reason,
	})
	if err != nil {
		_ = tx.Rollback()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, taskDTOFromDB(row))
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		clean := strings.TrimSpace(tag)
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func tagsToNullString(tags []string) (dbtypes.NullString, error) {
	clean := normalizeTags(tags)
	if len(clean) == 0 {
		return dbtypes.NullString{}, nil
	}
	payload, err := json.Marshal(clean)
	if err != nil {
		return dbtypes.NullString{}, err
	}
	return dbtypes.NullString{String: string(payload), Valid: true}, nil
}
