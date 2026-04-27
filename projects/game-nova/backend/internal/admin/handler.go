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
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/repo"
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


func decodeJSON(r *http.Request, into any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(into)
}

// EventRow — строка для admin events monitor.
type EventRow struct {
	ID          string     `json:"id"`
	UserID      *string    `json:"user_id,omitempty"`
	PlanetID    *string    `json:"planet_id,omitempty"`
	Kind        int        `json:"kind"`
	State       string     `json:"state"`
	FireAt      time.Time  `json:"fire_at"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	Attempt     int        `json:"attempt"`
	LastError   *string    `json:"last_error,omitempty"`
}

// EventsList GET /api/admin/events?state=wait|error|ok&kind=N&limit=...&offset=...
func (h *Handler) EventsList(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	state := r.URL.Query().Get("state")
	kindStr := r.URL.Query().Get("kind")

	where := []string{"1=1"}
	args := []any{}
	argIdx := 0
	if state == "wait" || state == "ok" || state == "error" {
		argIdx++
		args = append(args, state)
		where = append(where, fmt.Sprintf("state = $%d", argIdx))
	}
	if kindStr != "" {
		if k, err := strconv.Atoi(kindStr); err == nil {
			argIdx++
			args = append(args, k)
			where = append(where, fmt.Sprintf("kind = $%d", argIdx))
		}
	}
	argIdx++
	args = append(args, limit)
	limitIdx := argIdx
	argIdx++
	args = append(args, offset)
	offsetIdx := argIdx

	query := fmt.Sprintf(`
		SELECT id, user_id, planet_id, kind, state, fire_at, created_at,
		       processed_at, attempt, last_error
		FROM events
		WHERE %s
		ORDER BY COALESCE(processed_at, fire_at) DESC, id
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), limitIdx, offsetIdx)

	rows, err := h.db.Pool().Query(r.Context(), query, args...)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()
	out := []EventRow{}
	for rows.Next() {
		var e EventRow
		if err := rows.Scan(&e.ID, &e.UserID, &e.PlanetID, &e.Kind, &e.State,
			&e.FireAt, &e.CreatedAt, &e.ProcessedAt, &e.Attempt, &e.LastError); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		out = append(out, e)
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"events": out})
}

// EventsStats GET /api/admin/events/stats — агрегаты для дашборда.
func (h *Handler) EventsStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type stateCount struct {
		State string `json:"state"`
		Count int64  `json:"count"`
	}
	var byState []stateCount
	rows, err := h.db.Pool().Query(ctx,
		`SELECT state, COUNT(*) FROM events GROUP BY state`)
	if err == nil {
		for rows.Next() {
			var sc stateCount
			if err := rows.Scan(&sc.State, &sc.Count); err == nil {
				byState = append(byState, sc)
			}
		}
		rows.Close()
	}

	type kindCount struct {
		Kind  int   `json:"kind"`
		Count int64 `json:"count"`
	}
	var topErrors []kindCount
	rows, err = h.db.Pool().Query(ctx, `
		SELECT kind, COUNT(*) AS c FROM events
		WHERE state = 'error' AND processed_at > now() - interval '24 hours'
		GROUP BY kind ORDER BY c DESC LIMIT 10
	`)
	if err == nil {
		for rows.Next() {
			var kc kindCount
			if err := rows.Scan(&kc.Kind, &kc.Count); err == nil {
				topErrors = append(topErrors, kc)
			}
		}
		rows.Close()
	}

	// Lag: сколько секунд не обрабатывается самое старое wait с fire_at<=now.
	var lagSec *float64
	_ = h.db.Pool().QueryRow(ctx, `
		SELECT EXTRACT(EPOCH FROM (now() - MIN(fire_at)))
		FROM events WHERE state='wait' AND fire_at <= now()
	`).Scan(&lagSec)

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"by_state":          byState,
		"top_errors_24h":    topErrors,
		"oldest_wait_lag_s": lagSec,
	})
}

// EventRetry POST /api/admin/events/{id}/retry — сбросить error/сбросить attempt и поставить на fire_at=now().
func (h *Handler) EventRetry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}
	tag, err := h.db.Pool().Exec(r.Context(), `
		UPDATE events
		SET state = 'wait', attempt = 0, fire_at = now(), processed_at = NULL, last_error = NULL
		WHERE id = $1
	`, id)
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

// EventCancel POST /api/admin/events/{id}/cancel — пометить как ok без выполнения
// (soft-delete, чтобы worker не обрабатывал).
func (h *Handler) EventCancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.db.Pool().Exec(r.Context(), `
		UPDATE events
		SET state = 'ok', processed_at = now(),
		    last_error = COALESCE(last_error, '') || ' [cancelled by admin]'
		WHERE id = $1 AND state IN ('wait', 'error')
	`, id)
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

