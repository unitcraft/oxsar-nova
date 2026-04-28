package alien

// HTTP-адаптер для платного выкупа удержания (план 66 Ф.5).
//
// Отдельный тип BuyoutHandler, не размытие Service: Service нужен в
// worker'е (для event.Handler), а HTTP-handler нужен в server'е и
// требует БД-pool + billing-client + config. Чтобы не тащить эти
// зависимости в worker — отдельная структура с явным конструктором.
//
// R9 idempotency: ожидается, что роут обёрнут idempotency.Middleware.Wrap;
// здесь же мы только читаем заголовок Idempotency-Key и форвардим в
// billing.Spend для дедупликации на стороне billing-service.

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/repo"
)

// BuyoutHandler — HTTP-адаптер выкупа.
type BuyoutHandler struct {
	db      repo.Exec
	billing BuyoutBilling
	cfg     Config
}

// NewBuyoutHandler — конструктор. cfg — обычно Service.Config()
// (один и тот же default-конфиг alien-сервиса).
func NewBuyoutHandler(db repo.Exec, billing BuyoutBilling, cfg Config) *BuyoutHandler {
	return &BuyoutHandler{db: db, billing: billing, cfg: cfg}
}

// Buyout — POST /api/alien-missions/{mission_id}/buyout.
//
// Header Idempotency-Key обязателен (R9). Body не используется (URL +
// header достаточно), но idempotency-middleware всё равно хеширует
// тело — пустое тело = константный hash.
//
// Response codes (см. openapi.yaml /api/alien-missions/{mission_id}/buyout):
//   - 200: AlienBuyoutResponse {mission_id, cost_oxsars, freed_at}.
//   - 401: unauthorized (нет JWT).
//   - 402: insufficient_oxsars.
//   - 404: mission_not_found.
//   - 409: not_in_holding_state | idempotency_conflict.
//   - 503: billing_unavailable.
func (h *BuyoutHandler) Buyout(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	missionID := chi.URLParam(r, "mission_id")
	if missionID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "mission_id required"))
		return
	}
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "Idempotency-Key header required"))
		return
	}

	// User-token для billing — forward'им JWT из Authorization-заголовка.
	token := bearerToken(r)

	res, err := Buyout(r.Context(), h.db, h.billing, h.cfg, uid, missionID, token, idemKey)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissionNotFound):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "mission_not_found"))
		case errors.Is(err, ErrMissionAlreadyClosed):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "not_in_holding_state"))
		case errors.Is(err, ErrInsufficientOxsars):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrPaymentRequired, "insufficient_oxsars"))
		case errors.Is(err, ErrIdempotencyConflict):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "idempotency_conflict"))
		case errors.Is(err, ErrBillingUnavailable):
			httpx.WriteError(w, r, &httpx.Error{
				Status: http.StatusServiceUnavailable,
				Code:   "billing_unavailable",
				Message: "billing service unavailable; retry later",
			})
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, res)
}

// bearerToken извлекает RSA-JWT из Authorization: Bearer <token>.
// Если заголовка нет — пустая строка (billing-client это разрулит:
// для system-account'а Spend требует системного, не пользовательского,
// токена; в game-nova billing-вызовы делаются с user-токеном).
func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if len(auth) <= len(prefix) {
		return ""
	}
	if auth[:len(prefix)] != prefix {
		return ""
	}
	return auth[len(prefix):]
}
