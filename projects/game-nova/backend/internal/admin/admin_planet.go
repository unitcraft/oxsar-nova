package admin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/httpx"
)

// План 14 Ф.2.5 — admin-управление планетами.
//
//	POST   /api/admin/planets/{id}/rename    {"name":"..."}
//	POST   /api/admin/planets/{id}/transfer  {"new_user_id":"..."}
//	DELETE /api/admin/planets/{id}
//
// Защита: нельзя удалить последнюю живую планету игрока (тогда он
// остаётся без точки входа). rename не затрагивает координаты и
// postprocessing. transfer — меняет user_id, все дочерние сущности
// (buildings, ships, research — за исключением) следуют за планетой.

// PlanetRename POST /api/admin/planets/{id}/rename
func (h *Handler) PlanetRename(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "id")
	var body struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" || len(body.Name) > 40 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "name must be 1..40 chars"))
		return
	}
	tag, err := h.db.Pool().Exec(r.Context(),
		`UPDATE planets SET name = $1 WHERE id = $2 AND destroyed_at IS NULL`,
		body.Name, pid)
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

// PlanetTransfer POST /api/admin/planets/{id}/transfer
func (h *Handler) PlanetTransfer(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "id")
	var body struct {
		NewUserID string `json:"new_user_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	body.NewUserID = strings.TrimSpace(body.NewUserID)
	if body.NewUserID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "new_user_id required"))
		return
	}
	// Проверяем, что такой user существует и не удалён.
	var exists bool
	if err := h.db.Pool().QueryRow(r.Context(),
		`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`,
		body.NewUserID).Scan(&exists); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if !exists {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "new_user_id not found or deleted"))
		return
	}
	tag, err := h.db.Pool().Exec(r.Context(),
		`UPDATE planets SET user_id = $1 WHERE id = $2 AND destroyed_at IS NULL`,
		body.NewUserID, pid)
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

// PlanetDelete DELETE /api/admin/planets/{id} — soft-delete (destroyed_at=now).
// Не даёт удалить последнюю живую планету игрока.
func (h *Handler) PlanetDelete(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "id")

	// Владелец и количество живых планет.
	var ownerID string
	var aliveCount int
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT p.user_id,
		       (SELECT COUNT(*) FROM planets
		        WHERE user_id = p.user_id AND destroyed_at IS NULL AND is_moon = false)
		FROM planets p WHERE p.id = $1 AND p.destroyed_at IS NULL
	`, pid).Scan(&ownerID, &aliveCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if aliveCount <= 1 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict,
			"cannot delete last planet of user; transfer ownership first"))
		return
	}
	if _, err := h.db.Pool().Exec(r.Context(),
		`UPDATE planets SET destroyed_at = now() WHERE id = $1`, pid); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
