// Package notepad — личный блокнот игрока.
package notepad

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/pkg/metrics"
)

const MaxLength = 50_000

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

type notepadResponse struct {
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
}

// recordAction увеличивает Prometheus-счётчик и histogram (R8).
func recordAction(action, status string, started time.Time) {
	if metrics.NotepadActions != nil {
		metrics.NotepadActions.WithLabelValues(action, status).Inc()
	}
	if metrics.NotepadDuration != nil {
		metrics.NotepadDuration.WithLabelValues(action).Observe(time.Since(started).Seconds())
	}
}

// Get GET /api/notepad
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	uid, ok := auth.UserID(r.Context())
	if !ok {
		recordAction("get", "unauthorized", started)
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var content string
	var updatedAt time.Time
	err := h.pool.QueryRow(r.Context(),
		`SELECT content, updated_at FROM user_notepad WHERE user_id = $1`,
		uid,
	).Scan(&content, &updatedAt)
	if err != nil {
		// Нет записи — возвращаем пустой блокнот.
		recordAction("get", "ok", started)
		httpx.WriteJSON(w, r, http.StatusOK, notepadResponse{Content: "", UpdatedAt: ""})
		return
	}
	recordAction("get", "ok", started)
	httpx.WriteJSON(w, r, http.StatusOK, notepadResponse{
		Content:   content,
		UpdatedAt: updatedAt.UTC().Format(time.RFC3339),
	})
}

type saveRequest struct {
	Content string `json:"content"`
}

// Save PUT /api/notepad — идемпотентное сохранение.
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	uid, ok := auth.UserID(r.Context())
	if !ok {
		recordAction("save", "unauthorized", started)
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		recordAction("save", "bad_request", started)
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if len(req.Content) > MaxLength {
		recordAction("save", "too_long", started)
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "content too long"))
		return
	}
	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO user_notepad (user_id, content, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id) DO UPDATE SET content = EXCLUDED.content, updated_at = now()
	`, uid, req.Content)
	if err != nil {
		recordAction("save", "error", started)
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	recordAction("save", "ok", started)
	w.WriteHeader(http.StatusNoContent)
}
