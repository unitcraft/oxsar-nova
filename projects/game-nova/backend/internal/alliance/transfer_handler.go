// HTTP handlers для transfer-leadership (план 67 Ф.3).

package alliance

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/pkg/idempotency"
)

// Action-имена для Prometheus-метрики (вне alliance_audit_log —
// audit пишется только на confirm).
const (
	actionTransferCodeRequested = "leadership_transfer_code_requested"
	actionTransferConfirmed     = "leadership_transferred"
)

// RequestTransferLeadership POST /api/alliances/{id}/transfer-leadership/code
//
// Body: {"new_owner_id": uuid}.
// Только owner альянса. Идемпотентность через UPSERT в БД (последний
// запрос затирает предыдущий код); Idempotency-Key на этот запрос
// также применяется (R9) — одинаковый key вернёт тот же код-issuance
// без повторной отправки сообщения.
func (h *Handler) RequestTransferLeadership(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		NewOwnerID string `json:"new_owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if body.NewOwnerID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "new_owner_id required"))
		return
	}

	out, err := h.svc.RequestTransferCode(r.Context(), uid, id, body.NewOwnerID)
	recordAction(actionTransferCodeRequested, statusFromErr(err))
	switch {
	case err == nil:
		buf := httpx.MarshalJSON(out)
		idem.Record(http.StatusAccepted, buf)
		httpx.WriteJSONBytes(w, r, http.StatusAccepted, buf)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrTransferTargetIsSelf),
		errors.Is(err, ErrTransferTargetNotMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrTransferRateLimit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// ConfirmTransferLeadership POST /api/alliances/{id}/transfer-leadership
//
// Body: {"new_owner_id": uuid, "code": "XXXXXXXX"}.
// Только owner альянса. Idempotency-Key (R9): повторный confirm с тем
// же key вернёт сохранённый ответ — реальная передача не повторится.
func (h *Handler) ConfirmTransferLeadership(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	idem := idempotency.FromRequest(r, h.rdb)
	if idem.Replay(w) {
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		NewOwnerID string `json:"new_owner_id"`
		Code       string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if body.NewOwnerID == "" || body.Code == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "new_owner_id and code required"))
		return
	}

	err := h.svc.ConfirmTransferLeadership(r.Context(), uid, id, body.NewOwnerID, body.Code)
	recordAction(actionTransferConfirmed, statusFromErr(err))
	switch {
	case err == nil:
		idem.Record(http.StatusNoContent, nil)
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner), errors.Is(err, ErrTransferOwnerChanged):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrTransferNoCode),
		errors.Is(err, ErrTransferCodeExpired),
		errors.Is(err, ErrTransferCodeInvalid),
		errors.Is(err, ErrTransferTooManyAttempts),
		errors.Is(err, ErrTransferTargetMismatch),
		errors.Is(err, ErrTransferTargetIsSelf),
		errors.Is(err, ErrTransferTargetNotMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
