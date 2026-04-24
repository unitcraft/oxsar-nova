package payment

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Packages GET /api/payment/packages — публичный список пакетов.
func (h *Handler) Packages(w http.ResponseWriter, r *http.Request) {
	type packageDTO struct {
		Key          string  `json:"key"`
		Label        string  `json:"label"`
		Credits      int     `json:"credits"`
		BonusCredits int     `json:"bonus_credits"`
		TotalCredits int     `json:"total_credits"`
		PriceRub     float64 `json:"price_rub"`
	}
	type response struct {
		Packages []packageDTO `json:"packages"`
		TestMode bool         `json:"test_mode"`
	}
	packagesDTO := make([]packageDTO, len(Packages))
	for i, p := range Packages {
		packagesDTO[i] = packageDTO{
			Key:          p.Key,
			Label:        p.Label,
			Credits:      p.Credits,
			BonusCredits: p.BonusCredits,
			TotalCredits: p.TotalCredits(),
			PriceRub:     p.PriceRub(),
		}
	}
	httpx.WriteJSON(w, r, http.StatusOK, response{
		Packages: packagesDTO,
		TestMode: h.svc.IsMock(),
	})
}

type createOrderRequest struct {
	PackageKey string `json:"package_key"`
}

type createOrderResponse struct {
	OrderID string `json:"order_id"`
	PayURL  string `json:"pay_url"`
}

// CreateOrder POST /api/payment/order — создать заказ.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "unauthorized"))
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PackageKey == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "package_key is required"))
		return
	}

	orderID, payURL, err := h.svc.CreateOrder(r.Context(), uid, req.PackageKey)
	if err != nil {
		switch {
		case errors.Is(err, ErrPackageNotFound):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown package"))
		case errors.Is(err, ErrGatewayDisabled):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "payments not configured"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}

	httpx.WriteJSON(w, r, http.StatusOK, createOrderResponse{OrderID: orderID, PayURL: payURL})
}

// Webhook POST /api/payment/webhook — callback от шлюза (без авторизации).
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	if h.svc.gateway == nil {
		http.Error(w, "disabled", http.StatusServiceUnavailable)
		return
	}

	orderID, _, err := h.svc.gateway.VerifyWebhook(r)
	if err != nil {
		slog.Warn("payment: webhook verify failed", "err", err, "remote", r.RemoteAddr)
		http.Error(w, "bad signature", http.StatusBadRequest)
		return
	}

	// providerID — уникальный ID от шлюза для идемпотентности.
	providerID := r.FormValue("InvId")
	if providerID == "" {
		providerID = orderID
	}

	if err = h.svc.ConfirmPayment(r.Context(), orderID, providerID); err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			slog.Warn("payment: webhook unknown order", "order_id", orderID)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("payment: webhook confirm failed", "order_id", orderID, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.svc.gateway.SuccessResponse(w, orderID)
}

type purchaseDTO struct {
	ID           string     `json:"id"`
	PackageKey   string     `json:"package_key"`
	PackageLabel string     `json:"package_label"`
	Credits      int        `json:"credits"`
	PriceRub     float64    `json:"price_rub"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
}

// MockPay GET /api/payment/mock/pay — симулятор платёжной страницы (только provider=mock).
// Принимает ?order=<id>&result=success|fail, подтверждает платёж и редиректит
// игрока на ReturnURL с ?payment=success|fail.
func (h *Handler) MockPay(w http.ResponseWriter, r *http.Request) {
	if !h.svc.IsMock() {
		http.Error(w, "mock gateway is not active", http.StatusNotFound)
		return
	}

	orderID := r.URL.Query().Get("order")
	result := r.URL.Query().Get("result")
	if result == "" {
		result = "success"
	}
	returnURL := r.URL.Query().Get("return")
	if returnURL == "" {
		returnURL = h.svc.cfg.ReturnURL
	}
	if returnURL == "" {
		returnURL = "/"
	}

	paymentStatus := "fail"
	if result == "success" && orderID != "" {
		if err := h.svc.ConfirmPayment(r.Context(), orderID, "mock-"+orderID); err != nil {
			if !errors.Is(err, ErrOrderNotFound) {
				slog.Error("payment: mock confirm failed", "order_id", orderID, "err", err)
			}
		} else {
			paymentStatus = "success"
		}
	}

	sep := "?"
	if strings.Contains(returnURL, "?") {
		sep = "&"
	}
	http.Redirect(w, r, returnURL+sep+"payment="+paymentStatus, http.StatusFound)
}

// History GET /api/payment/history — история покупок игрока.
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "unauthorized"))
		return
	}

	purchases, err := h.svc.ListPurchases(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	result := make([]purchaseDTO, len(purchases))
	for i, p := range purchases {
		result[i] = purchaseDTO{
			ID:           p.ID,
			PackageKey:   p.PackageKey,
			PackageLabel: p.PackageLabel,
			Credits:      p.Credits,
			PriceRub:     p.PriceRub,
			Status:       p.Status,
			CreatedAt:    p.CreatedAt,
			PaidAt:       p.PaidAt,
		}
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}
