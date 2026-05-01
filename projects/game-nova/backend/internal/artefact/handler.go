package artefact

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List GET /api/artefacts — инвентарь текущего пользователя.
//
// План 72.1.45: legacy `Artefacts.class.php` показывает в шапке
// storage_slots/used_slots/tech_level (research UNIT_ARTEFACTS_TECH=111).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	items, err := h.svc.ListUser(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	techLevel, usedSlots, err := h.svc.SlotsInfo(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"artefacts":     items,
		"tech_level":    techLevel,
		"storage_slots": techLevel,
		"used_slots":    usedSlots,
	})
}

// Activate POST /api/artefacts/{id}/activate
func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	rec, err := h.svc.Activate(r.Context(), uid, id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, rec)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrAlreadyActive):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "already active"))
	case errors.Is(err, ErrNonStackable):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "non-stackable already active"))
	case errors.Is(err, ErrMaxStacksReached):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "max stacks already active"))
	case errors.Is(err, ErrPlanetRequired):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "artefact requires planet"))
	case errors.Is(err, ErrUnknownArtefact):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown artefact"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// PackBuilding POST /api/planets/{id}/buildings/{unitId}/pack (план 72.1.33 ч.2).
//
// Legacy `BuildingInfo::packCurrentConstruction`: если у игрока на этой
// планете есть held packing-building артефакт (unit_id=323), упаковывает
// здание (level - 1) в packed-building артефакт (unit_id=321) с
// payload {construction_id, level=1}.
func (h *Handler) PackBuilding(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	unitID, err := strconv.Atoi(chi.URLParam(r, "unitId"))
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid unit_id"))
		return
	}
	rec, err := h.svc.PackBuilding(r.Context(), uid, planetID, unitID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, rec)
	case errors.Is(err, ErrPackingArtefactNotFound):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "no held packing-building artefact on this planet"))
	case errors.Is(err, ErrNothingToPack):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "building level=0, nothing to pack"))
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// PackResearch POST /api/planets/{id}/research/{unitId}/pack.
//
// Аналог PackBuilding для исследований (legacy `packCurrentResearch`).
// planetID нужен для контекста активации (legacy берёт NS::getPlanet()).
func (h *Handler) PackResearch(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	unitID, err := strconv.Atoi(chi.URLParam(r, "unitId"))
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid unit_id"))
		return
	}
	rec, err := h.svc.PackResearch(r.Context(), uid, planetID, unitID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, rec)
	case errors.Is(err, ErrPackingArtefactNotFound):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "no held packing-research artefact on this planet"))
	case errors.Is(err, ErrNothingToPack):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "research level=0, nothing to pack"))
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// History GET /api/artefacts/info/{unitId}/history — план 72.1.45 §2.
//
// Legacy `ArtefactInfo::showInfo` (L.66) SELECT artefact_history JOIN user
// JOIN assault — кому/когда достался артефакт + ссылка на отчёт боя.
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	unitID, err := strconv.Atoi(chi.URLParam(r, "unitId"))
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid unit_id"))
		return
	}
	entries, err := h.svc.History(r.Context(), unitID, 50)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"entries": entries,
	})
}

// Deactivate POST /api/artefacts/{id}/deactivate
func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Deactivate(r.Context(), uid, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
