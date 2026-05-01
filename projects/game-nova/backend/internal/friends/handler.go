// Package friends — двусторонний friendship c подтверждением (план 72.1.14).
//
// Соответствует legacy `Friends.class.php`/`buddylist (friend1, friend2,
// accepted)`. Schema (миграция 0086):
//
//	friends (user_id, friend_id, accepted bool, created_at)
//	  PRIMARY KEY (user_id, friend_id), CHECK user_id <> friend_id.
//
// Семантика accepted:
//
//	false → pending: A→B запрос отправлен, B видит «входящий».
//	true  → mutual: симметричная пара (A,B,true) и (B,A,true) — оба
//	        видят друг друга в списке друзей.
//
// Add создаёт pending (A→B). Если встречная (B,A,_) уже существует —
// mutual auto-accept (legacy эквивалент: A добавляет B, ранее B уже
// добавил A → friendship сразу принят с обеих сторон).
//
// Accept (POST /api/friends/{userId}/accept): транзакция, обновляет
// встречную запись (friend, me, accepted=true) и создаёт симметричную
// (me, friend, accepted=true).
//
// Remove (DELETE /api/friends/{userId}): удаляет ОБЕ стороны
// одновременно. Соответствует legacy `delete WHERE relid AND
// (friend1=me OR friend2=me)` — удалить запись, в которой я фигурирую.
//
// AutoMsg на add/accept/remove НЕ реализуем: legacy имеет TODO-комментарии
// на этих местах, но фактически сообщения не отправляет (план 72.1.14
// §AutoMsg — байтовое соответствие legacy).
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
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	Points      float64 `json:"points"`
	LastSeen    *string `json:"last_seen,omitempty"`
	AllianceTag *string `json:"alliance_tag,omitempty"`
	// Accepted=true → mutual friend; false → pending (см. направление).
	Accepted bool `json:"accepted"`
	// Direction: "incoming" — мне отправили запрос, "outgoing" — я отправил,
	// "mutual" — accepted в обе стороны.
	Direction string `json:"direction"`
}

// List GET /api/friends?pending=all|incoming|outgoing
//
// По умолчанию (без pending) возвращает только accepted=true (mutual
// друзья). pending=incoming → только входящие запросы (friend_id=me,
// accepted=false). outgoing → отправленные мной (user_id=me,
// accepted=false). all → объединение accepted+pending обоих направлений.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	mode := r.URL.Query().Get("pending")

	// SQL подбирается по mode. Везде JOIN users для имени/очков/last_seen.
	var (
		sql  string
		args = []any{uid}
	)
	switch mode {
	case "incoming":
		// Запросы, пришедшие мне: их сторона добавила, я ещё не принял.
		sql = `
			SELECT u.id, u.username, u.points, u.last_seen, a.tag,
			       false AS accepted, 'incoming' AS direction
			FROM friends f
			JOIN users u ON u.id = f.user_id AND u.deleted_at IS NULL
			LEFT JOIN alliances a ON a.id = u.alliance_id
			WHERE f.friend_id = $1 AND NOT f.accepted
			ORDER BY u.username
		`
	case "outgoing":
		// Я отправил, ждёт подтверждения с их стороны.
		sql = `
			SELECT u.id, u.username, u.points, u.last_seen, a.tag,
			       false AS accepted, 'outgoing' AS direction
			FROM friends f
			JOIN users u ON u.id = f.friend_id AND u.deleted_at IS NULL
			LEFT JOIN alliances a ON a.id = u.alliance_id
			WHERE f.user_id = $1 AND NOT f.accepted
			ORDER BY u.username
		`
	case "all":
		// Все стороны: accepted (моё view) + pending in/out.
		sql = `
			SELECT u.id, u.username, u.points, u.last_seen, a.tag,
			       f.accepted,
			       CASE
			           WHEN f.accepted THEN 'mutual'
			           WHEN f.user_id = $1 THEN 'outgoing'
			           ELSE 'incoming'
			       END AS direction
			FROM friends f
			JOIN users u ON u.id = CASE WHEN f.user_id = $1 THEN f.friend_id ELSE f.user_id END
			            AND u.deleted_at IS NULL
			LEFT JOIN alliances a ON a.id = u.alliance_id
			WHERE (f.user_id = $1 AND f.accepted)
			   OR (f.user_id = $1 AND NOT f.accepted)
			   OR (f.friend_id = $1 AND NOT f.accepted)
			ORDER BY u.username
		`
	default:
		// Mutual только (значение по умолчанию — список «реальных» друзей).
		sql = `
			SELECT u.id, u.username, u.points, u.last_seen, a.tag,
			       true AS accepted, 'mutual' AS direction
			FROM friends f
			JOIN users u ON u.id = f.friend_id AND u.deleted_at IS NULL
			LEFT JOIN alliances a ON a.id = u.alliance_id
			WHERE f.user_id = $1 AND f.accepted
			ORDER BY u.username
		`
	}

	rows, err := h.pool.Query(r.Context(), sql, args...)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	out := []friendRow{}
	for rows.Next() {
		var f friendRow
		var lastSeen *time.Time
		if err := rows.Scan(&f.UserID, &f.Username, &f.Points, &lastSeen, &f.AllianceTag, &f.Accepted, &f.Direction); err != nil {
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

// Add POST /api/friends/{userId} — отправить запрос в друзья.
//
// Семантика (legacy buddylist):
//   - Если уже есть запись (me, target, _) — no-op (или 409, но legacy
//     просто игнорирует через count проверку).
//   - Если есть встречная (target, me, accepted=false) — mutual
//     auto-accept: создать (me, target, true) И обновить встречную
//     accepted=true (legacy эквивалент: добавление переводит pending в
//     accepted). Транзакция.
//   - Если есть встречная (target, me, accepted=true) — что бы это ни
//     значило (рассинхрон), делаем (me, target, true) — mutual.
//   - Иначе обычный pending: INSERT (me, target, accepted=false).
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

	ctx := r.Context()

	// Проверка существования цели.
	var exists bool
	if err := h.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`,
		friendID).Scan(&exists); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if !exists {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}

	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Если моя запись (me, target) уже есть — no-op.
	var mine bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM friends WHERE user_id = $1 AND friend_id = $2)`,
		uid, friendID).Scan(&mine); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if mine {
		// Идемпотентный no-op (legacy: count > 0 → ничего не делает).
		_ = tx.Commit(ctx)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Есть ли встречная (target, me)? Если есть — auto-accept.
	var counter bool
	var counterAccepted bool
	if err := tx.QueryRow(ctx, `
		SELECT TRUE, accepted FROM friends
		WHERE user_id = $1 AND friend_id = $2
	`, friendID, uid).Scan(&counter, &counterAccepted); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	if counter {
		// Auto-accept: создаём свою сторону как accepted=true, встречную
		// тоже выставляем true (если ещё не).
		if _, err := tx.Exec(ctx, `
			INSERT INTO friends (user_id, friend_id, accepted)
			VALUES ($1, $2, true)
		`, uid, friendID); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		if !counterAccepted {
			if _, err := tx.Exec(ctx, `
				UPDATE friends SET accepted = true
				WHERE user_id = $1 AND friend_id = $2
			`, friendID, uid); err != nil {
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
				return
			}
		}
	} else {
		// Обычный pending запрос.
		if _, err := tx.Exec(ctx, `
			INSERT INTO friends (user_id, friend_id, accepted)
			VALUES ($1, $2, false)
		`, uid, friendID); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Accept POST /api/friends/{userId}/accept — принять входящий запрос.
//
// Условие: существует (sender=userId, me, accepted=false).
// Транзакция:
//  1. UPDATE встречной записи accepted=true.
//  2. INSERT моей симметричной (me, sender, accepted=true).
//
// Идемпотентен: если уже accepted=true с обеих сторон — 204 без изменений.
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	senderID := chi.URLParam(r, "userId")
	if senderID == "" || senderID == uid {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid user id"))
		return
	}

	ctx := r.Context()
	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Должна быть запись от sender ко мне.
	var senderAccepted bool
	if err := tx.QueryRow(ctx, `
		SELECT accepted FROM friends
		WHERE user_id = $1 AND friend_id = $2
	`, senderID, uid).Scan(&senderAccepted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	// Перевести встречную в accepted (no-op если уже true).
	if !senderAccepted {
		if _, err := tx.Exec(ctx, `
			UPDATE friends SET accepted = true
			WHERE user_id = $1 AND friend_id = $2
		`, senderID, uid); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}

	// Моя симметричная сторона (idempotent).
	if _, err := tx.Exec(ctx, `
		INSERT INTO friends (user_id, friend_id, accepted)
		VALUES ($1, $2, true)
		ON CONFLICT (user_id, friend_id) DO UPDATE SET accepted = true
	`, uid, senderID); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Remove DELETE /api/friends/{userId} — удалить friendship с обеих сторон.
//
// Соответствует legacy `delete WHERE relid AND (friend1=me OR friend2=me)`:
// обе записи (me,$1) и ($1,me) удаляются в одной транзакции. Работает
// и для accepted, и для pending (отмена своей outgoing-заявки или
// отклонение incoming).
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	friendID := chi.URLParam(r, "userId")
	if _, err := h.pool.Exec(r.Context(), `
		DELETE FROM friends
		WHERE (user_id = $1 AND friend_id = $2)
		   OR (user_id = $2 AND friend_id = $1)
	`, uid, friendID); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
