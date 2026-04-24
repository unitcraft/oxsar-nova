package alien

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/repo"
)

// Handler — HTTP-адаптер alien-сервиса. Нужен отдельный тип (а не методы
// на Service), чтобы не тащить БД-зависимость в AttackHandler-воркер.
type Handler struct{ db repo.Exec }

func NewHandler(db repo.Exec) *Handler { return &Handler{db: db} }

// Pay — POST /api/alien/holding/{event_id}/pay
// body: { "amount": <int64> }
// Списывает amount кредитов у текущего пользователя и продлевает HOLDING
// по формуле 2ч/50 кредитов с cap на 15 дней от начала HOLDING.
func (h *Handler) Pay(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	eventID := chi.URLParam(r, "event_id")
	if eventID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "event_id required"))
		return
	}
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	res, err := PayHolding(r.Context(), h.db, uid, eventID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, ErrPayAmountInvalid):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "amount must be positive"))
		case errors.Is(err, ErrHoldingNotFound):
			httpx.WriteError(w, r, httpx.ErrNotFound)
		case errors.Is(err, ErrHoldingNotOwner):
			httpx.WriteError(w, r, httpx.ErrForbidden)
		case errors.Is(err, ErrInsufficientCred):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "insufficient credits"))
		case errors.Is(err, ErrHoldingAtCap):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "holding already at max duration"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, res)
}
