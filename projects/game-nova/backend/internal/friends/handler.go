// Package friends — список друзей игрока (односторонний: добавление
// не требует подтверждения).
package friends

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

type friendRow struct {
	UserID     string  `json:"user_id"`
	Username   string  `json:"username"`
	Points     float64 `json:"points"`
	LastSeen   *string `json:"last_seen,omitempty"`
	AllianceTag *string `json:"alliance_tag,omitempty"`
}

// List GET /api/friends — список друзей текущего пользователя.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	rows, err := h.pool.Query(r.Context(), `
		SELECT u.id, u.username, u.points, u.last_seen, a.tag
		FROM friends f
		JOIN users u ON u.id = f.friend_id AND u.deleted_at IS NULL
		LEFT JOIN alliances a ON a.id = u.alliance_id
		WHERE f.user_id = $1
		ORDER BY u.username
	`, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	out := []friendRow{}
	for rows.Next() {
		var f friendRow
		var lastSeen *time.Time
		if err := rows.Scan(&f.UserID, &f.Username, &f.Points, &lastSeen, &f.AllianceTag); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		if lastSeen != nil {
			s := lastSeen.UTC().Format(time.RFC3339)
			f.LastSeen = &s
		}
		out = append(out, f)
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"friends": out})
}

// Add POST /api/friends/{userId} — добавить друга.
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	friendID := chi.URLParam(r, "userId")
	if friendID == "" || friendID == uid {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid user id"))
		return
	}
	// Проверка существования и что не deleted.
	var exists bool
	if err := h.pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`,
		friendID).Scan(&exists); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if !exists {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}

	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO friends (user_id, friend_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, uid, friendID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Remove DELETE /api/friends/{userId}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	friendID := chi.URLParam(r, "userId")
	if _, err := h.pool.Exec(r.Context(),
		`DELETE FROM friends WHERE user_id = $1 AND friend_id = $2`,
		uid, friendID); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
