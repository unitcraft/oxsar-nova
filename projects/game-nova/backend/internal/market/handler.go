package market

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/pkg/idempotency"
)

// ListLots GET /api/market/lots?sell=metal&limit=50
func (h *Handler) ListLots(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	sell := r.URL.Query().Get("sell")
	lots, err := h.svc.ListLots(r.Context(), sell, 50)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if lots == nil {
		lots = []Lot{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, lots)
}

// CreateLot POST /api/market/lots
func (h *Handler) CreateLot(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		PlanetID     string `json:"planet_id"`
		SellResource string `json:"sell_resource"`
		SellAmount   int64  `json:"sell_amount"`
		BuyResource  string `json:"buy_resource"`
		BuyAmount    int64  `json:"buy_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	lot, err := h.svc.CreateLot(r.Context(), uid, req.PlanetID,
		req.SellResource, req.SellAmount, req.BuyResource, req.BuyAmount)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, lot)
	case errors.Is(err, ErrInvalidResource), errors.Is(err, ErrSameResource),
		errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrNotEnough):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrPlanetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// CancelLot DELETE /api/market/lots/{id}
func (h *Handler) CancelLot(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	lotID := chi.URLParam(r, "id")
	err := h.svc.CancelLot(r.Context(), uid, lotID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrLotNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrLotNotOpen):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// AcceptLot POST /api/market/lots/{id}/accept
func (h *Handler) AcceptLot(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}

	lotID := chi.URLParam(r, "id")
	var req struct {
		PlanetID string `json:"planet_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	err := h.svc.AcceptLot(r.Context(), uid, req.PlanetID, lotID)
	switch {
	case err == nil:
		idem.Record(http.StatusNoContent, nil)
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrLotNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrLotNotOpen), errors.Is(err, ErrOwnLot),
		errors.Is(err, ErrNotEnough):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrPlanetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

type Handler struct {
	svc *Service
	rdb *redis.Client
}

func NewHandler(s *Service, rdb *redis.Client) *Handler { return &Handler{svc: s, rdb: rdb} }

type exchangeRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

// Exchange POST /api/planets/{id}/market/exchange
func (h *Handler) Exchange(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	if planetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing planet id"))
		return
	}
	var req exchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	res, err := h.svc.Exchange(r.Context(), uid, planetID, req.From, req.To, req.Amount)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, res)
	case errors.Is(err, ErrInvalidResource),
		errors.Is(err, ErrSameResource),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrNotEnough):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrPlanetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

type creditExchangeRequest struct {
	Direction string  `json:"direction"` // только "from_credit" (оставлено для совместимости)
	Resource  string  `json:"resource"`
	Amount    float64 `json:"amount"`
}

// ExchangeCredit POST /api/planets/{id}/market/credit — покупка ресурса
// за кредиты. Обратное направление (продажа ресурсов за кредиты) удалено
// 2026-04-26 — было уязвимостью (бесконечный фарминг premium-валюты).
func (h *Handler) ExchangeCredit(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	if planetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing planet id"))
		return
	}
	var req creditExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	res, err := h.svc.ExchangeCredit(r.Context(), uid, planetID, req.Direction, req.Resource, req.Amount)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, res)
	case errors.Is(err, ErrInvalidResource),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrNotEnough):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrPlanetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// ListFleetLots GET /api/market/fleet-lots
func (h *Handler) ListFleetLots(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	lots, err := h.svc.ListFleetLots(r.Context(), 100)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if lots == nil {
		lots = []FleetLot{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"lots": lots})
}

type createFleetLotRequest struct {
	Fleet       map[int]int64 `json:"fleet"`
	BuyResource string        `json:"buy_resource"`
	BuyAmount   int64         `json:"buy_amount"`
}

// CreateFleetLot POST /api/planets/{id}/market/fleet-lots
func (h *Handler) CreateFleetLot(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	var req createFleetLotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	lot, err := h.svc.CreateFleetLot(r.Context(), uid, planetID, req.Fleet, req.BuyResource, req.BuyAmount)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, lot)
	case errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrInvalidResource), errors.Is(err, ErrNotEnough):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrPlanetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

type acceptFleetLotRequest struct {
	PlanetID string `json:"planet_id"`
}

// AcceptFleetLot POST /api/market/fleet-lots/{lotId}/accept
func (h *Handler) AcceptFleetLot(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	lotID := chi.URLParam(r, "lotId")
	var req acceptFleetLotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	err := h.svc.AcceptFleetLot(r.Context(), uid, req.PlanetID, lotID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotEnough):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrLotNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrLotNotOpen), errors.Is(err, ErrOwnLot):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrPlanetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// CancelFleetLot DELETE /api/market/fleet-lots/{lotId}
func (h *Handler) CancelFleetLot(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	lotID := chi.URLParam(r, "lotId")
	err := h.svc.CancelFleetLot(r.Context(), uid, lotID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrLotNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrLotNotOpen), errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Rates GET /api/market/rates
func (h *Handler) Rates(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	rates, err := h.svc.Rates(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, rates)
}
