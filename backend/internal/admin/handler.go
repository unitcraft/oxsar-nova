// Package admin — административные эндпоинты (M8).
//
// Доступ: role IN ('admin', 'superadmin'). Middleware AdminOnly
// проверяет роль; обычному игроку возвращает 403.
//
// Реализованные эндпоинты:
//   GET  /api/admin/users         — список игроков (limit/offset)
//   POST /api/admin/users/{id}/ban    — заблокировать (banned_at)
//   POST /api/admin/users/{id}/unban  — разблокировать
//   POST /api/admin/users/{id}/credit — выдать/снять кредиты
//   POST /api/admin/users/{id}/role   — изменить роль
package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/repo"
)

type Handler struct {
	db repo.Exec
}

func NewHandler(db repo.Exec) *Handler { return &Handler{db: db} }

// AdminOnly — middleware, пропускающий только admin/superadmin.
func AdminOnly(db repo.Exec) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid, _ := auth.UserID(r.Context())
			if uid == "" {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			var role string
			err := db.Pool().QueryRow(r.Context(),
				`SELECT role FROM users WHERE id = $1`, uid).Scan(&role)
			if err != nil || (role != "admin" && role != "superadmin") {
				httpx.WriteError(w, r, httpx.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserRow — строка таблицы игроков для админ-просмотра.
type UserRow struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Credit     int64      `json:"credit"`
	BannedAt   *time.Time `json:"banned_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt time.Time  `json:"last_seen_at"`
}

// ListUsers GET /api/admin/users?limit=50&offset=0
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT id, username, email, role, credit, banned_at, created_at, last_seen_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	var out []UserRow
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Credit,
			&u.BannedAt, &u.CreatedAt, &u.LastSeenAt); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"users": out})
}

// Ban POST /api/admin/users/{id}/ban
func (h *Handler) Ban(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	tag, err := h.db.Pool().Exec(r.Context(),
		`UPDATE users SET banned_at = now() WHERE id = $1 AND banned_at IS NULL`, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "user not found or already banned"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unban POST /api/admin/users/{id}/unban
func (h *Handler) Unban(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	tag, err := h.db.Pool().Exec(r.Context(),
		`UPDATE users SET banned_at = NULL WHERE id = $1`, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "user not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Credit POST /api/admin/users/{id}/credit  body: {"delta": N}
func (h *Handler) Credit(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	var body struct {
		Delta int64 `json:"delta"`
	}
	if err := decodeJSON(r, &body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	var newCredit int64
	err := h.db.Pool().QueryRow(r.Context(), `
		UPDATE users SET credit = credit + $1 WHERE id = $2
		RETURNING credit
	`, body.Delta, uid).Scan(&newCredit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"credit": newCredit})
}

// SetRole POST /api/admin/users/{id}/role  body: {"role": "admin"|"player"|...}
func (h *Handler) SetRole(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	var body struct {
		Role string `json:"role"`
	}
	if err := decodeJSON(r, &body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	validRoles := map[string]bool{"player": true, "support": true, "admin": true, "superadmin": true}
	if !validRoles[body.Role] {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest,
			fmt.Sprintf("invalid role %q", body.Role)))
		return
	}
	tag, err := h.db.Pool().Exec(r.Context(),
		`UPDATE users SET role = $1 WHERE id = $2`, body.Role, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, into any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(into)
}

// Stats GET /api/admin/stats — базовая статистика сервера.
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var totalUsers, activeLast24h, totalPlanets int64
	_ = h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalUsers)
	_ = h.db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE last_seen_at > now() - interval '24 hours'`,
	).Scan(&activeLast24h)
	_ = h.db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM planets WHERE destroyed_at IS NULL`,
	).Scan(&totalPlanets)

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"total_users":    totalUsers,
		"active_last24h": activeLast24h,
		"total_planets":  totalPlanets,
	})
}

