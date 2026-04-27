package admin

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/httpx"
)

// План 14 Ф.2.6 — soft-delete и restore.
//
//	DELETE /api/admin/users/{id}
//	POST   /api/admin/users/{id}/restore
//
// Soft-delete ставит `deleted_at = now()`, отвязывает игрока от альянса
// (alliance_id=NULL, alliance_members — DELETE). Планеты не трогаем —
// они останутся заброшены, могут быть колонизированы другими игроками
// (после истечения protection_period планеты-сироты попадают в общий
// доступ через cleanup-worker в будущем; сейчас они просто не
// обслуживаются).
//
// Restore снимает deleted_at. Восстановление в альянс — ручное
// (админ должен сделать join).

// UserSoftDelete DELETE /api/admin/users/{id}
func (h *Handler) UserSoftDelete(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	ctx := r.Context()
	err := h.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`,
			uid).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return errUserNotFoundOrDeleted
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET deleted_at = now(), alliance_id = NULL WHERE id = $1`,
			uid); err != nil {
			return err
		}
		// Отсоединяем от альянса (alliance_members).
		if _, err := tx.Exec(ctx,
			`DELETE FROM alliance_members WHERE user_id = $1`, uid); err != nil {
			return err
		}
		return nil
	})
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, errUserNotFoundOrDeleted):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// UserRestore POST /api/admin/users/{id}/restore
func (h *Handler) UserRestore(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "id")
	tag, err := h.db.Pool().Exec(r.Context(),
		`UPDATE users SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "user not found or not deleted"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var errUserNotFoundOrDeleted = errors.New("admin: user not found or already deleted")
