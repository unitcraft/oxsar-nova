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
	Score      int64      `json:"score"`
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
		SELECT u.id, u.username, u.email, COALESCE(u.role::text,''), u.credit,
		       COALESCE(s.score, 0), u.banned_at, u.created_at, u.last_seen_at
		FROM users u
		LEFT JOIN scores s ON s.user_id = u.id
		ORDER BY u.created_at DESC LIMIT $1 OFFSET $2
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
			&u.Score, &u.BannedAt, &u.CreatedAt, &u.LastSeenAt); err != nil {
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

// Credit POST /api/admin/users/{id}/credit  body: {"amount": N}
func (h *Handler) Credit(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	var body struct {
		Amount int64 `json:"amount"`
	}
	if err := decodeJSON(r, &body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	var newCredit int64
	err := h.db.Pool().QueryRow(r.Context(), `
		UPDATE users SET credit = credit + $1 WHERE id = $2
		RETURNING credit
	`, body.Amount, uid).Scan(&newCredit)
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

// AutomsgDef — шаблон системного сообщения.
type AutomsgDef struct {
	Key          string `json:"key"`
	Title        string `json:"title"`
	BodyTemplate string `json:"body_template"`
	Folder       int    `json:"folder"`
}

// ListAutomsgs GET /api/admin/automsgs
func (h *Handler) ListAutomsgs(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Pool().Query(r.Context(),
		`SELECT key, title, body_template, folder FROM automsg_defs ORDER BY key`)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()
	var out []AutomsgDef
	for rows.Next() {
		var d AutomsgDef
		if err := rows.Scan(&d.Key, &d.Title, &d.BodyTemplate, &d.Folder); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"defs": out})
}

// UpdateAutomsg PUT /api/admin/automsgs/{key}
// Body: {"title":"...","body_template":"...","folder":2}
func (h *Handler) UpdateAutomsg(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var body struct {
		Title        string `json:"title"`
		BodyTemplate string `json:"body_template"`
		Folder       int    `json:"folder"`
	}
	if err := decodeJSON(r, &body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	if body.Title == "" || body.BodyTemplate == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "title and body_template required"))
		return
	}
	tag, err := h.db.Pool().Exec(r.Context(), `
		UPDATE automsg_defs SET title=$1, body_template=$2, folder=$3 WHERE key=$4
	`, body.Title, body.BodyTemplate, body.Folder, key)
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
	var users, planets, fleetsActive, eventsPending int64
	_ = h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&users)
	_ = h.db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM planets WHERE destroyed_at IS NULL`).Scan(&planets)
	_ = h.db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM fleets WHERE state IN ('outbound','inbound')`).Scan(&fleetsActive)
	_ = h.db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM events WHERE fire_at > now()`).Scan(&eventsPending)

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"users":          users,
		"planets":        planets,
		"fleets_active":  fleetsActive,
		"events_pending": eventsPending,
	})
}

