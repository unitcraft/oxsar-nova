package admin

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/fleet"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/repo"
)

// План 14 Ф.2.4 — force-recall fleet.
//
//	POST /api/admin/fleets/{fleet_id}/recall
//
// Возвращает флот в источник как обычный recall, но без owner-check
// (админ может recall-нуть чужой флот). Работает только для
// state='outbound'.

// FleetAdminHandler изолирован от основного admin.Handler, чтобы не
// плодить зависимости: transport нужен только этому методу.
type FleetAdminHandler struct {
	transport *fleet.TransportService
	db        repo.Exec
}

func NewFleetAdminHandler(transport *fleet.TransportService, db repo.Exec) *FleetAdminHandler {
	return &FleetAdminHandler{transport: transport, db: db}
}

// ForceRecall POST /api/admin/fleets/{fleet_id}/recall
func (h *FleetAdminHandler) ForceRecall(w http.ResponseWriter, r *http.Request) {
	fleetID := chi.URLParam(r, "fleet_id")
	if fleetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "fleet_id required"))
		return
	}
	// Узнаём владельца и вызываем Recall от его имени.
	var ownerID string
	err := h.db.Pool().QueryRow(r.Context(),
		`SELECT owner_user_id FROM fleets WHERE id = $1`, fleetID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	f, err := h.transport.Recall(r.Context(), ownerID, fleetID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, f)
	case errors.Is(err, fleet.ErrFleetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, fleet.ErrFleetNotRecallable):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
