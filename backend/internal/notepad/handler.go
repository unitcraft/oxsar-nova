// Package notepad — личный блокнот игрока.
package notepad

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
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

// Get GET /api/notepad
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
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
		httpx.WriteJSON(w, r, http.StatusOK, notepadResponse{Content: "", UpdatedAt: ""})
		return
	}
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
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if len(req.Content) > MaxLength {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "content too long"))
		return
	}
	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO user_notepad (user_id, content, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id) DO UPDATE SET content = EXCLUDED.content, updated_at = now()
	`, uid, req.Content)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
