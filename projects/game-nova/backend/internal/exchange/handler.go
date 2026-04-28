package exchange

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// Routes регистрирует REST-эндпоинты биржи на router'е chi.
//
// Idempotency-Key middleware применяется СНАРУЖИ (в cmd/server/main.go),
// потому что у нас уже есть pkg/idempotency/Middleware.Wrap с Redis.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/api/exchange/lots", h.List)
	r.Post("/api/exchange/lots", h.Create)
	r.Get("/api/exchange/lots/{id}", h.Get)
	r.Delete("/api/exchange/lots/{id}", h.Cancel)
	r.Post("/api/exchange/lots/{id}/buy", h.Buy)
	r.Get("/api/exchange/stats", h.Stats)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	q := r.URL.Query()
	f := ListFilters{Limit: parseIntDefault(q.Get("limit"), 50)}
	if v := q.Get("artifact_unit_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.ArtifactUnitID = &n
		}
	}
	if v := q.Get("min_price"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.MinPrice = &n
		}
	}
	if v := q.Get("max_price"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.MaxPrice = &n
		}
	}
	if v := q.Get("seller_id"); v != "" {
		f.SellerID = &v
	}
	if v := q.Get("status"); v != "" {
		f.Status = &v
	}
	f.Cursor = q.Get("cursor")

	lots, nextCursor, err := h.svc.ListLots(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	resp := map[string]any{
		"lots": lots,
	}
	if nextCursor != "" {
		resp["next_cursor"] = nextCursor
	} else {
		resp["next_cursor"] = nil
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "Idempotency-Key required"))
		return
	}
	var req struct {
		ArtifactUnitID int   `json:"artifact_unit_id"`
		Quantity       int   `json:"quantity"`
		PriceOxsarit   int64 `json:"price_oxsarit"`
		ExpiresInHours int   `json:"expires_in_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	lot, err := h.svc.CreateLot(r.Context(), CreateLotInput{
		SellerUserID:   uid,
		ArtifactUnitID: req.ArtifactUnitID,
		Quantity:       req.Quantity,
		PriceOxsarit:   req.PriceOxsarit,
		ExpiresInHours: req.ExpiresInHours,
		IdempotencyKey: idemKey,
	})
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, map[string]any{"lot": lot})
	case errors.Is(err, ErrInvalidQuantity), errors.Is(err, ErrInvalidPrice),
		errors.Is(err, ErrInvalidExpiry):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrInsufficientArtefacts):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnprocessable, "insufficient_artefacts"))
	case errors.Is(err, ErrPriceCapExceeded):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnprocessable, "price_cap_exceeded"))
	case errors.Is(err, ErrPermitRequired):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnprocessable, "permit_required"))
	case errors.Is(err, ErrMaxActiveLots):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnprocessable, "max_active_lots"))
	case errors.Is(err, ErrMaxQuantity):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnprocessable, "max_quantity"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	lot, items, err := h.svc.GetLotWithItems(r.Context(), id)
	switch {
	case err == nil:
		itemsOut := make([]map[string]any, 0, len(items))
		for _, aid := range items {
			itemsOut = append(itemsOut, map[string]any{"artefact_id": aid})
		}
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
			"lot":   lot,
			"items": itemsOut,
		})
	case errors.Is(err, ErrLotNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	err := h.svc.CancelLot(r.Context(), id, uid)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrLotNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotASeller):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrLotNotActive):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "lot_not_active"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

func (h *Handler) Buy(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	if r.Header.Get("Idempotency-Key") == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "Idempotency-Key required"))
		return
	}
	id := chi.URLParam(r, "id")
	lot, err := h.svc.BuyLot(r.Context(), id, uid)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"lot": lot})
	case errors.Is(err, ErrLotNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrLotNotActive):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "lot_not_active"))
	case errors.Is(err, ErrCannotBuyOwnLot):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "cannot_buy_own_lot"))
	case errors.Is(err, ErrInsufficientOxsarits):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrPaymentRequired, "insufficient_oxsarits"))
	case errors.Is(err, ErrUserHasNoPlanet):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "buyer_has_no_planet"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	out := make([]map[string]any, 0, len(stats))
	for _, s := range stats {
		row := map[string]any{
			"artifact_unit_id": s.ArtifactUnitID,
			"active_lots":      s.ActiveLots,
			"avg_unit_price":   s.AvgUnitPrice,
			"last_30d_volume":  s.Last30dVolume,
		}
		out = append(out, row)
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"items": out})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
