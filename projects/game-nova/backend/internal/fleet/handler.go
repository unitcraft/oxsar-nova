package fleet

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/galaxy"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/pkg/idempotency"
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
	Mission      int           `json:"mission"`      // 7=TRANSPORT, 10=ATTACK, 12=ACS, 17=HOLDING …
	ACSGroupID   string        `json:"acs_group_id"` // только для mission=12; пусто → создать группу
	ColonyName   string        `json:"colony_name"`  // только для mission=8; пусто → «Colony»
	HoldingHours int           `json:"holding_hours"` // только для mission=17; clamp 0..99
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
		req.Mission != 10 && req.Mission != 11 && req.Mission != 12 && req.Mission != 15 &&
		req.Mission != 17 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest,
			"supported missions: 7=TRANSPORT, 8=COLONIZE, 9=RECYCLING, 10=ATTACK_SINGLE, 11=SPY, 12=ACS, 15=EXPEDITION, 17=HOLDING"))
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
		HoldingHours: req.HoldingHours,
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
		errors.Is(err, ErrBashingLimit),
		errors.Is(err, ErrPositionNotAllowed),
		errors.Is(err, ErrExpeditionSlotsFull):
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

// Phalanx GET /api/phalanx?source_planet_id=UUID&target_galaxy=N&target_system=M
// Сенсорная Фаланга — план 20 Ф.4.
func (h *Handler) Phalanx(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	q := r.URL.Query()
	source := q.Get("source_planet_id")
	if source == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "source_planet_id required"))
		return
	}
	g, err1 := parseIntParam(q.Get("target_galaxy"))
	s, err2 := parseIntParam(q.Get("target_system"))
	if err1 != nil || err2 != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "target_galaxy and target_system must be integers"))
		return
	}
	scans, err := h.transport.Phalanx(r.Context(), uid, source, g, s)
	switch {
	case err == nil:
		if scans == nil {
			scans = []PhalanxScan{}
		}
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"scans": scans})
	case errors.Is(err, ErrPhalanxNotAMoon),
		errors.Is(err, ErrPhalanxNotInstalled),
		errors.Is(err, ErrPhalanxDifferentGalax):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPhalanxOutOfRange),
		errors.Is(err, ErrPhalanxNoHydrogen):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Load POST /api/fleet/{id}/load — загрузить ресурсы с current_planet во флот.
// План 72.1.47: legacy `Mission.class.php::loadResourcesToFleet`.
// Body: { "current_planet_id": "...", "metal": N, "silicon": N, "hydrogen": N }
func (h *Handler) Load(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		CurrentPlanetID string `json:"current_planet_id"`
		Metal           int64  `json:"metal"`
		Silicon         int64  `json:"silicon"`
		Hydrogen        int64  `json:"hydrogen"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	err := h.transport.LoadResources(r.Context(), LoadUnloadInput{
		UserID: uid, FleetID: fleetID, CurrentPlanetID: body.CurrentPlanetID,
		Metal: body.Metal, Silicon: body.Silicon, Hydrogen: body.Hydrogen,
	})
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrFleetNotFound), errors.Is(err, ErrTargetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrFleetNotHolding),
		errors.Is(err, ErrPlanetNotDst),
		errors.Is(err, ErrLoadFromOwnSrc),
		errors.Is(err, ErrLoadCapacity),
		errors.Is(err, ErrControlsExhausted):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrInvalidDispatch):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Unload POST /api/fleet/{id}/unload — выгрузить ресурсы с флота на current_planet.
// План 72.1.47: legacy `Mission.class.php::unloadResourcesFromFleet`.
func (h *Handler) Unload(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		CurrentPlanetID string `json:"current_planet_id"`
		Metal           int64  `json:"metal"`
		Silicon         int64  `json:"silicon"`
		Hydrogen        int64  `json:"hydrogen"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	err := h.transport.UnloadResources(r.Context(), LoadUnloadInput{
		UserID: uid, FleetID: fleetID, CurrentPlanetID: body.CurrentPlanetID,
		Metal: body.Metal, Silicon: body.Silicon, Hydrogen: body.Hydrogen,
	})
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrFleetNotFound), errors.Is(err, ErrTargetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrFleetNotHolding),
		errors.Is(err, ErrPlanetNotDst),
		errors.Is(err, ErrControlsExhausted),
		errors.Is(err, ErrInsufficientReturnFuel):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrInvalidDispatch):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// PromoteToACS POST /api/fleet/{id}/promote-to-acs
// План 72.1.48: legacy `Mission.class.php::formation` — конверсия
// уже летящего ATTACK_SINGLE в именованную ACS-группу.
// Body: { "name": "..." }
func (h *Handler) PromoteToACS(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	g, err := h.transport.PromoteToACS(r.Context(), uid, fleetID, body.Name)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, g)
	case errors.Is(err, ErrFleetNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrFleetNotPromotable),
		errors.Is(err, ErrAlreadyPromoted):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrInvalidName):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// InviteACS POST /api/acs/{groupId}/invite
// Body: { "username": "..." }
func (h *Handler) InviteACS(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	groupID := chi.URLParam(r, "groupId")
	if groupID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing group id"))
		return
	}
	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	err := h.transport.InviteToFormation(r.Context(), uid, groupID, body.Username)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrInvitationNotFound), errors.Is(err, ErrInviteeNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotLeader):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrSelfInvite),
		errors.Is(err, ErrNoRelation):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// ListACSInvitations GET /api/acs/invitations — pending+accepted
// инвайты для текущего юзера.
func (h *Handler) ListACSInvitations(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	out, err := h.transport.ListInvitations(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"invitations": out})
}

// AcceptACSInvitation POST /api/acs/invitations/{groupId}/accept
func (h *Handler) AcceptACSInvitation(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	groupID := chi.URLParam(r, "groupId")
	if groupID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing group id"))
		return
	}
	err := h.transport.AcceptInvitation(r.Context(), uid, groupID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrInvitationNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Stargate POST /api/stargate
// Body: { "src_planet_id": "...", "dst_planet_id": "...", "ships": {"unit_id": count} }
func (h *Handler) Stargate(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var body struct {
		SrcPlanetID string          `json:"src_planet_id"`
		DstPlanetID string          `json:"dst_planet_id"`
		Ships       map[int]int64   `json:"ships"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	res, err := h.transport.StargateJump(r.Context(), StargateJumpInput{
		UserID:      uid,
		SrcPlanetID: body.SrcPlanetID,
		DstPlanetID: body.DstPlanetID,
		Ships:       body.Ships,
	})
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, res)
	case errors.Is(err, ErrInvalidDispatch),
		errors.Is(err, ErrStargateBannedUnit),
		errors.Is(err, ErrStargateNotMoon),
		errors.Is(err, ErrStargateNotInstalled),
		errors.Is(err, ErrStargatePositionLimit),
		errors.Is(err, ErrTargetNotFound):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrStargateNotOwner),
		errors.Is(err, ErrPositionNotAllowed):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrStargateCooldown),
		errors.Is(err, ErrStargateNotEnoughShip):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

func parseIntParam(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty")
	}
	n := 0
	neg := false
	i := 0
	if s[0] == '-' {
		neg = true
		i = 1
	}
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, errors.New("invalid")
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}
