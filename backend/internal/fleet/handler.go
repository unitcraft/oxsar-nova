package fleet

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/galaxy"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/pkg/idempotency"
)

// Handler — HTTP-адаптер к TransportService.
//
// POST /api/fleet       — отправить транспорт (пока единственная миссия)
// GET  /api/fleet       — список активных флотов игрока
//
// Расширение на остальные миссии (ATTACK/SPY/COLONIZE) пойдёт в
// M4/M5 — тогда handler переедет в общий Service с диспетчером
// по полю mission.
type Handler struct {
	transport *TransportService
	rdb       *redis.Client // может быть nil — тогда idempotency отключена
}

func NewHandler(t *TransportService, rdb *redis.Client) *Handler {
	return &Handler{transport: t, rdb: rdb}
}

type sendRequest struct {
	SrcPlanetID  string        `json:"src_planet_id"`
	Dst          galaxy.Coords `json:"dst"`
	Ships        map[int]int64 `json:"ships"`
	CarryMetal   int64         `json:"carry_metal"`
	CarrySilicon int64         `json:"carry_silicon"`
	CarryHydro   int64         `json:"carry_hydrogen"`
	SpeedPercent int           `json:"speed_percent"`
	Mission      int           `json:"mission"`      // 7=TRANSPORT, 10=ATTACK, 12=ACS, …
	ACSGroupID   string        `json:"acs_group_id"` // только для mission=12; пусто → создать группу
	ColonyName   string        `json:"colony_name"`  // только для mission=8; пусто → «Colony»
}

// Send POST /api/fleet
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if req.Mission != 0 && req.Mission != 7 && req.Mission != 8 && req.Mission != 9 &&
		req.Mission != 10 && req.Mission != 11 && req.Mission != 12 && req.Mission != 15 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest,
			"supported missions: 7=TRANSPORT, 8=COLONIZE, 9=RECYCLING, 10=ATTACK_SINGLE, 11=SPY, 12=ACS, 15=EXPEDITION"))
		return
	}
	in := TransportInput{
		UserID:       uid,
		SrcPlanetID:  req.SrcPlanetID,
		Dst:          req.Dst,
		Mission:      req.Mission,
		ACSGroupID:   req.ACSGroupID,
		ColonyName:   req.ColonyName,
		Ships:        req.Ships,
		CarryMetal:   req.CarryMetal,
		CarrySilicon: req.CarrySilicon,
		CarryHydro:   req.CarryHydro,
		SpeedPercent: req.SpeedPercent,
	}
	f, err := h.transport.Send(r.Context(), in)
	switch {
	case err == nil:
		body := httpx.MarshalJSON(f)
		idem.Record(http.StatusCreated, body)
		httpx.WriteJSONBytes(w, r, http.StatusCreated, body)
	case errors.Is(err, ErrInvalidDispatch),
		errors.Is(err, ErrNotEnoughShips),
		errors.Is(err, ErrNotEnoughCarry),
		errors.Is(err, ErrExceedCargoCap),
		errors.Is(err, ErrTargetNotFound),
		errors.Is(err, ErrUnknownShip):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrFleetSlotsExceeded):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrTargetOnVacation),
		errors.Is(err, ErrSenderOnVacation),
		errors.Is(err, ErrBashingLimit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// List GET /api/fleet
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	list, err := h.transport.List(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	slotsUsed, slotsMax, err := h.transport.Slots(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"fleets":     list,
		"slots_used": slotsUsed,
		"slots_max":  slotsMax,
	})
}

// Incoming GET /api/fleet/incoming — вражеские атакующие флоты к планетам игрока.
func (h *Handler) Incoming(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	list, err := h.transport.ListIncoming(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if list == nil {
		list = []IncomingFleet{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"fleets": list})
}

// Recall POST /api/fleet/{id}/recall — досрочный возврат флота. Работает
// только для флотов в состоянии outbound.
func (h *Handler) Recall(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	fleetID := chi.URLParam(r, "id")
	if fleetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing fleet id"))
		return
	}
	f, err := h.transport.Recall(r.Context(), uid, fleetID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, f)
	case errors.Is(err, ErrFleetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrFleetNotRecallable):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
