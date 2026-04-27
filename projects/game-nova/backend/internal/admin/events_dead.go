// Dead-letter события (план 14 Ф.3.2-3.3).
//
// events_dead заполняется отдельным cron'ом (см. worker) из events, где
// state='error' AND processed_at < now() - N days. Здесь — чтение +
// ручное воскрешение (перенос обратно в events со сбросом attempt).

package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/httpx"
)

type DeadEvent struct {
	ID          string          `json:"id"`
	UserID      *string         `json:"user_id,omitempty"`
	PlanetID    *string         `json:"planet_id,omitempty"`
	Kind        int             `json:"kind"`
	FireAt      time.Time       `json:"fire_at"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	ProcessedAt *time.Time      `json:"processed_at,omitempty"`
	Attempt     int             `json:"attempt"`
	LastError   string          `json:"last_error"`
	FailedAt    time.Time       `json:"failed_at"`
}

// ListDeadEvents GET /api/admin/events/dead?kind=&limit=&offset=
func (h *Handler) ListDeadEvents(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	q := `
		SELECT id, user_id, planet_id, kind, fire_at, payload,
		       created_at, processed_at, attempt, COALESCE(last_error, ''), failed_at
		FROM events_dead WHERE 1=1`
	args := []any{}

	if k := r.URL.Query().Get("kind"); k != "" {
		if n, err := strconv.Atoi(k); err == nil {
			args = append(args, n)
			q += " AND kind = $" + strconv.Itoa(len(args))
		}
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		args = append(args, uid)
		q += " AND user_id = $" + strconv.Itoa(len(args))
	}

	args = append(args, limit, offset)
	q += " ORDER BY failed_at DESC LIMIT $" + strconv.Itoa(len(args)-1) +
		" OFFSET $" + strconv.Itoa(len(args))

	rows, err := h.db.Pool().Query(r.Context(), q, args...)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	out := []DeadEvent{}
	for rows.Next() {
		var e DeadEvent
		var payloadBytes []byte
		if err := rows.Scan(&e.ID, &e.UserID, &e.PlanetID, &e.Kind,
			&e.FireAt, &payloadBytes, &e.CreatedAt, &e.ProcessedAt,
			&e.Attempt, &e.LastError, &e.FailedAt); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		if len(payloadBytes) > 0 {
			e.Payload = payloadBytes
		} else {
			e.Payload = json.RawMessage("{}")
		}
		out = append(out, e)
	}

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"events": out,
		"limit":  limit,
		"offset": offset,
	})
}

// ResurrectDeadEvent POST /api/admin/events/dead/{id}/resurrect
// — переносит событие обратно в events с attempt=0, fire_at=now().
// Удаляет из events_dead после успешного INSERT.
func (h *Handler) ResurrectDeadEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}

	ctx := r.Context()
	err := h.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Читаем исходное событие.
		var userID, planetID *string
		var kind int
		var payload []byte
		var createdAt time.Time
		err := tx.QueryRow(ctx, `
			SELECT user_id, planet_id, kind, payload, created_at
			FROM events_dead WHERE id = $1
		`, id).Scan(&userID, &planetID, &kind, &payload, &createdAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}

		// Вставляем в events с fire_at=now(), attempt=0. ID сохраняем.
		_, err = tx.Exec(ctx, `
			INSERT INTO events (id, user_id, planet_id, kind, fire_at, payload, state, attempt, created_at)
			VALUES ($1, $2, $3, $4, now(), $5, 'wait', 0, $6)
			ON CONFLICT (id) DO UPDATE
			  SET state = 'wait', attempt = 0, fire_at = now(),
			      processed_at = NULL, last_error = NULL
		`, id, userID, planetID, kind, payload, createdAt)
		if err != nil {
			return err
		}

		// Убираем из dead-letter.
		_, err = tx.Exec(ctx, `DELETE FROM events_dead WHERE id = $1`, id)
		return err
	})

	if err != nil {
		if errors.Is(err, httpx.ErrNotFound) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
