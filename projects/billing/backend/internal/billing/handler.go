package billing

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"oxsar/billing/internal/httpx"
	"oxsar/billing/internal/payment"
)

// Handler — HTTP-адаптер billing.
type Handler struct {
	svc       *Service
	gateway   payment.Gateway
	returnURL string
}

func NewHandler(svc *Service, gw payment.Gateway, returnURL string) *Handler {
	return &Handler{svc: svc, gateway: gw, returnURL: returnURL}
}

// SpendInput — request body для POST /billing/wallet/spend.
type spendRequest struct {
	Amount    int64  `json:"amount"`
	Reason    string `json:"reason"`
	RefID     string `json:"ref_id,omitempty"`
	ToAccount string `json:"to_account"` // куда логически уходят деньги
	Currency  string `json:"currency,omitempty"`
}

// Spend — POST /billing/wallet/spend (требует JWT).
//
// Headers:
//   Authorization: Bearer <jwt>
//   Idempotency-Key: <client-uuid>   (опционально)
//
// Body:
//   { "amount": 100, "reason": "feedback_vote", "ref_id": "<uuid>",
//     "to_account": "vote:feedback:<uuid>" }
func (h *Handler) Spend(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req spendRequest
	if err := decodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	if req.Reason == "" || req.ToAccount == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "reason and to_account required"))
		return
	}
	tx, err := h.svc.Spend(r.Context(), SpendInput{
		UserID:         userID,
		Currency:       req.Currency,
		Amount:         req.Amount,
		Reason:         req.Reason,
		RefID:          req.RefID,
		ToAccount:      req.ToAccount,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		writeBillingError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, tx)
}

// creditRequest — POST /billing/wallet/credit.
//
// Внутренний endpoint: вызывается из webhook-handler-а при пополнении.
// Защищён JWT, но также проверяет роль 'admin' или 'system' (TBD: в Ф.4
// webhook будет вызывать ServiceMethod напрямую, без HTTP).
//
// Сейчас оставлен для admin-grant-операций.
type creditRequest struct {
	Amount      int64  `json:"amount"`
	Reason      string `json:"reason"`
	RefID       string `json:"ref_id,omitempty"`
	FromAccount string `json:"from_account"`
	Currency    string `json:"currency,omitempty"`
	TargetUser  string `json:"target_user_id,omitempty"` // admin может пополнять другому
}

func (h *Handler) Credit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req creditRequest
	if err := decodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	target := userID
	if req.TargetUser != "" && req.TargetUser != userID {
		// Пополнение чужого кошелька — только admin.
		if !HasRole(r, "admin") {
			httpx.WriteError(w, r, httpx.ErrForbidden)
			return
		}
		target = req.TargetUser
	}
	if req.Reason == "" || req.FromAccount == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "reason and from_account required"))
		return
	}
	tx, err := h.svc.Credit(r.Context(), CreditInput{
		UserID:         target,
		Currency:       req.Currency,
		Amount:         req.Amount,
		Reason:         req.Reason,
		RefID:          req.RefID,
		FromAccount:    req.FromAccount,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		writeBillingError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, tx)
}

// Balance — GET /billing/wallet/balance?currency=OXC
func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	currency := r.URL.Query().Get("currency")
	wal, err := h.svc.Balance(r.Context(), userID, currency)
	if err != nil {
		writeBillingError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, wal)
}

// Packages — GET /billing/packages (публично, без auth).
func (h *Handler) Packages(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"packages": payment.Packages})
}

// CreateOrder — POST /billing/orders { "package_id": "pack_500" } (требует JWT).
type createOrderRequest struct {
	PackageID string `json:"package_id"`
	ReturnURL string `json:"return_url,omitempty"`
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req createOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	if req.PackageID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "package_id required"))
		return
	}
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = h.returnURL
	}
	order, payURL, err := h.svc.CreateOrder(r.Context(), userID, req.PackageID, returnURL, h.gateway)
	if err != nil {
		switch {
		case errors.Is(err, payment.ErrPackageNotFound):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "package not found"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusCreated, map[string]any{
		"order":   order,
		"pay_url": payURL,
	})
}

// History — GET /billing/wallet/history?limit=50&offset=0&currency=OXC
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromCtx(r)
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	currency := r.URL.Query().Get("currency")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	txs, err := h.svc.History(r.Context(), userID, currency, limit, offset)
	if err != nil {
		writeBillingError(w, r, err)
		return
	}
	if txs == nil {
		txs = []Transaction{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"transactions": txs})
}

// writeBillingError транслирует доменные ошибки в HTTP-статусы.
func writeBillingError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInsufficient):
		httpx.WriteError(w, r, httpx.ErrPaymentRequired)
	case errors.Is(err, ErrFrozen):
		httpx.WriteError(w, r, httpx.ErrLocked)
	case errors.Is(err, ErrInvalidAmount):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

func decodeJSON(r *http.Request, into any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(into)
}
