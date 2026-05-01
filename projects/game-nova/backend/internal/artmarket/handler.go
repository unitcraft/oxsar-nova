package artmarket

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

type Handler struct {
	svc *Service
	rdb *redis.Client
}

func NewHandler(s *Service, rdb *redis.Client) *Handler { return &Handler{svc: s, rdb: rdb} }

// Offers GET /api/artefact-market/offers
func (h *Handler) Offers(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	list, err := h.svc.ListOffers(r.Context())
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"offers": list})
}

// Credit GET /api/artefact-market/credit
func (h *Handler) Credit(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	n, err := h.svc.Credit(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]int64{"credit": n})
}

type listForSaleRequest struct {
	Price int64 `json:"price"`
}

// ListForSale POST /api/artefacts/{id}/sell
func (h *Handler) ListForSale(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	artID := chi.URLParam(r, "id")
	if artID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing artefact id"))
		return
	}
	var req listForSaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	off, err := h.svc.ListForSale(r.Context(), uid, artID, req.Price)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, off)
	case errors.Is(err, ErrInvalidPrice),
		errors.Is(err, ErrArtefactNotHeld):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrArtefactNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrUmodeBlocked), errors.Is(err, ErrSellerBanned):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Buy POST /api/artefact-market/offers/{id}/buy
func (h *Handler) Buy(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}

	offerID := chi.URLParam(r, "id")
	if offerID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing offer id"))
		return
	}
	err := h.svc.Buy(r.Context(), uid, offerID)
	switch {
	case err == nil:
		idem.Record(http.StatusNoContent, nil)
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotEnoughCredit),
		errors.Is(err, ErrOwnOffer):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrOfferNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrUmodeBlocked):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Cancel DELETE /api/artefact-market/offers/{id}
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	offerID := chi.URLParam(r, "id")
	if offerID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing offer id"))
		return
	}
	err := h.svc.Cancel(r.Context(), uid, offerID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrOfferNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
