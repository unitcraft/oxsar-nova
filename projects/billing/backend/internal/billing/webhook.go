package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"oxsar/billing/internal/httpx"
	"oxsar/billing/internal/payment"
)

// WebhookHandler принимает webhook от платёжного шлюза.
//
// Алгоритм:
//   1. Прочитать сырой body (LimitReader 64KB).
//   2. INSERT webhook_log (raw body, headers) — ДО verify, чтобы audit trail
//      был даже у невалидных запросов (полезно при разборе атак).
//   3. Gateway.VerifyWebhook(r, body) → orderID, amount, err.
//      Внутри: проверка signature (HMAC) + timestamp (±5 мин) — replay protection.
//   4. Если verify упал — UPDATE webhook_log SET signature_ok=false, error=...
//      и вернуть 400 (или 401 для подписи).
//   5. Если verify ОК — UPDATE webhook_log SET signature_ok=true, order_id=<id>
//      → svc.PayOrder(orderID): атомарное UPDATE order.status=paid +
//      INSERT transaction + UPDATE wallet.balance.
//   6. UPDATE webhook_log SET processed_at=now().
//   7. Gateway.SuccessResponse(w, orderID).
type WebhookHandler struct {
	svc     *Service
	gateway payment.Gateway
}

func NewWebhookHandler(svc *Service, gw payment.Gateway) *WebhookHandler {
	return &WebhookHandler{svc: svc, gateway: gw}
}

// Handle — POST /billing/webhooks/{provider}.
//
// Public endpoint (rate-limit отдельно — TODO Ф.4.1, можно через Redis-RL
// как в identity-service). IP allowlist — через nginx на проде.
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "read body: "+err.Error()))
		return
	}
	logID, logErr := h.logWebhookRaw(r.Context(), body, r.Header)
	if logErr != nil {
		// audit log не должен блокировать обработку — но логируем проблему.
		slog.WarnContext(r.Context(), "webhook log insert failed",
			slog.String("err", logErr.Error()))
	}

	orderID, amountKop, verifyErr := h.gateway.VerifyWebhook(r, body)
	if verifyErr != nil {
		_ = h.markVerifyFailed(r.Context(), logID, verifyErr)
		switch {
		case errors.Is(verifyErr, payment.ErrSignatureMismatch):
			WebhooksTotal.WithLabelValues(h.gateway.Name(), "invalid_signature").Inc()
			httpx.WriteError(w, r, httpx.ErrUnauthorized)
		case errors.Is(verifyErr, payment.ErrTimestampOld):
			WebhooksTotal.WithLabelValues(h.gateway.Name(), "expired").Inc()
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "timestamp out of window"))
		default:
			WebhooksTotal.WithLabelValues(h.gateway.Name(), "invalid_payload").Inc()
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, verifyErr.Error()))
		}
		return
	}

	// Подпись валидна — фиксируем в логе и пробуем pay-order.
	_ = h.markVerifyOK(r.Context(), logID, orderID)

	// Защита: проверим, что amount в payload совпадает с amount в order.
	// Это защищает от inflated webhook (тот же signature, другая сумма —
	// невозможно для HMAC, но всё равно).
	order, err := h.svc.GetOrder(r.Context(), orderID)
	if err != nil {
		_ = h.markProcessFailed(r.Context(), logID, err)
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "order not found"))
		return
	}
	if order.AmountKop != amountKop {
		_ = h.markProcessFailed(r.Context(), logID, fmt.Errorf("amount mismatch: webhook=%d order=%d",
			amountKop, order.AmountKop))
		slog.WarnContext(r.Context(), "webhook amount mismatch",
			slog.String("order_id", orderID),
			slog.Int64("webhook_amount", amountKop),
			slog.Int64("order_amount", order.AmountKop))
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "amount mismatch"))
		return
	}

	if err := h.svc.PayOrder(r.Context(), orderID); err != nil {
		_ = h.markProcessFailed(r.Context(), logID, err)
		switch {
		case errors.Is(err, ErrOrderExpired):
			WebhooksTotal.WithLabelValues(h.gateway.Name(), "order_expired").Inc()
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "order expired"))
		default:
			WebhooksTotal.WithLabelValues(h.gateway.Name(), "error").Inc()
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	_ = h.markProcessed(r.Context(), logID)
	WebhooksTotal.WithLabelValues(h.gateway.Name(), "ok").Inc()
	TransactionsTotal.WithLabelValues("top_up", "credit").Inc()
	h.gateway.SuccessResponse(w, orderID)
}

// logWebhookRaw записывает сырой webhook в таблицу webhook_log.
// Возвращает id записи (для последующих UPDATE при verify/process).
func (h *WebhookHandler) logWebhookRaw(ctx context.Context, body []byte, headers http.Header) (string, error) {
	headersJSON, _ := json.Marshal(headers)
	var id string
	row := h.svc.db.Pool().QueryRow(ctx, `
		INSERT INTO webhook_log (provider, headers, body)
		VALUES ($1, $2, $3)
		RETURNING id
	`, h.gateway.Name(), headersJSON, body)
	err := row.Scan(&id)
	return id, err
}

func (h *WebhookHandler) markVerifyFailed(ctx context.Context, logID string, e error) error {
	if logID == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	_, err := h.svc.db.Pool().Exec(ctx, `
		UPDATE webhook_log
		SET signature_ok = false, error = $1, processed_at = now()
		WHERE id = $2
	`, e.Error(), logID)
	return err
}

func (h *WebhookHandler) markVerifyOK(ctx context.Context, logID, orderID string) error {
	if logID == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	_, err := h.svc.db.Pool().Exec(ctx, `
		UPDATE webhook_log
		SET signature_ok = true, order_id = $1::uuid
		WHERE id = $2
	`, orderID, logID)
	return err
}

func (h *WebhookHandler) markProcessed(ctx context.Context, logID string) error {
	if logID == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	_, err := h.svc.db.Pool().Exec(ctx, `
		UPDATE webhook_log SET processed_at = now() WHERE id = $1
	`, logID)
	return err
}

func (h *WebhookHandler) markProcessFailed(ctx context.Context, logID string, e error) error {
	if logID == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	_, err := h.svc.db.Pool().Exec(ctx, `
		UPDATE webhook_log SET error = $1, processed_at = now() WHERE id = $2
	`, e.Error(), logID)
	return err
}
