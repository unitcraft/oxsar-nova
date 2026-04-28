package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"oxsar/game-nova/pkg/metrics"
)

const (
	defaultTimeout = 10 * time.Second

	// retryBackoff — пауза между attempt'ами при транзиентной ошибке.
	// Линейный backoff (одна попытка ретрая) — MVP, см. doc.go.
	retryBackoff = 200 * time.Millisecond

	// maxAttempts — общее число попыток (1 первичная + 1 retry).
	maxAttempts = 2
)

// Client — HTTP-клиент к billing-service для списаний/возвратов оксаров.
//
// Безопасен для конкурентного использования: http.Client потокобезопасен,
// никакого изменяемого состояния в Client больше нет.
type Client struct {
	billingURL string
	httpClient *http.Client
}

// New создаёт клиент с заданным базовым URL billing-service.
// Если billingURL пустой — клиент будет возвращать ErrNotConfigured на каждом
// вызове (по образцу portal/credits.go: позволяет game-nova стартовать на
// окружениях без billing — например, локальный dev без премиум-фич).
func New(billingURL string) *Client {
	return &Client{
		billingURL: billingURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
}

// SpendInput — параметры списания оксаров.
type SpendInput struct {
	// UserToken — RSA-JWT юзера, forward'ится в Authorization: Bearer ...
	UserToken string

	// Amount — сумма в минимальных единицах оксара (целые числа,
	// см. ADR-0009). > 0 для Spend; для Refund тоже > 0 (Refund сам
	// разбирает направление).
	Amount int64

	// Reason — символическое имя операции для аудита billing.
	// Примеры: "teleport_planet", "alien_buyout", "marketplace_permit".
	Reason string

	// RefID — ID связанной игровой сущности (planet_id, holding_id, lot_id).
	// Используется для трассировки и связки billing-tx с game-event.
	RefID string

	// ToAccount — символический «куда уходят оксары».
	// Примеры: "system:teleport", "system:alien_buyout".
	ToAccount string

	// IdempotencyKey — обязательно. Рекомендуемый формат:
	// "<user_id>:<operation>:<ref_id>" — гарантирует уникальность по
	// (user, operation, target).
	IdempotencyKey string
}

// Spend списывает Amount оксаров с user-кошелька через billing-service.
//
// Возвращает:
//   - nil при успехе (HTTP 200).
//   - ErrInsufficientOxsar при HTTP 402.
//   - ErrIdempotencyConflict при HTTP 409 (тот же ключ, другой body).
//   - ErrFrozenWallet при HTTP 423.
//   - ErrBillingUnavailable при network/timeout/5xx после retry.
//   - ErrNotConfigured если billingURL пустой.
//   - error прочее (для логирования + общий 500 наружу).
func (c *Client) Spend(ctx context.Context, in SpendInput) error {
	return c.do(ctx, "spend", "/billing/wallet/spend", in)
}

// Refund возвращает Amount оксаров на user-кошелёк (для отмены telepor'а,
// отказа от alien-buyout и т.п.). Idempotency-Key должен отличаться от
// исходного Spend — рекомендуется суффикс ":refund" к Spend-ключу.
//
// Реализация: POST /billing/wallet/credit (обратная операция Spend),
// от system-account'а на user-кошелёк. Reason должен явно указать причину
// возврата ("teleport_cancelled", "buyout_failed").
func (c *Client) Refund(ctx context.Context, in SpendInput) error {
	return c.do(ctx, "refund", "/billing/wallet/credit", in)
}

// do — общая отправка с retry на транзиентных ошибках.
func (c *Client) do(ctx context.Context, op, path string, in SpendInput) (retErr error) {
	if c.billingURL == "" {
		incSpendStatus(op, "unavailable")
		return ErrNotConfigured
	}

	start := time.Now()
	defer func() {
		observeDuration(op, time.Since(start))
	}()

	body, err := buildBody(op, in)
	if err != nil {
		incSpendStatus(op, "error")
		return err
	}

	endpoint, err := url.JoinPath(c.billingURL, path)
	if err != nil {
		incSpendStatus(op, "error")
		return fmt.Errorf("billing client: build URL: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := c.attempt(ctx, endpoint, in, body)
		if err == nil {
			incSpendStatus(op, "ok")
			return nil
		}
		// Терминальные доменные ошибки — возвращаем сразу, без retry.
		if errors.Is(err, ErrInsufficientOxsar) {
			incSpendStatus(op, "insufficient")
			return err
		}
		if errors.Is(err, ErrIdempotencyConflict) {
			incSpendStatus(op, "conflict")
			return err
		}
		if errors.Is(err, ErrFrozenWallet) {
			incSpendStatus(op, "frozen")
			return err
		}
		// Caller отменил контекст — не ретраим, не маркируем unavailable
		// (это решение caller'а, не сбой billing).
		if ctx.Err() != nil {
			incSpendStatus(op, "error")
			return ctx.Err()
		}
		lastErr = err
		if !isTransient(err) {
			incSpendStatus(op, "error")
			return err
		}
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				incSpendStatus(op, "error")
				return ctx.Err()
			case <-time.After(retryBackoff):
			}
		}
	}
	incSpendStatus(op, "unavailable")
	return fmt.Errorf("%w: %v", ErrBillingUnavailable, lastErr)
}

// attempt — одна HTTP-попытка. Возвращает sentinel-ошибку для известных
// статусов или generic-ошибку для остальных.
func (c *Client) attempt(ctx context.Context, endpoint string, in SpendInput, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("billing client: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if in.UserToken != "" {
		req.Header.Set("Authorization", "Bearer "+in.UserToken)
	}
	if in.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", in.IdempotencyKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("billing client: do: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusPaymentRequired:
		return ErrInsufficientOxsar
	case http.StatusConflict:
		return ErrIdempotencyConflict
	case http.StatusLocked:
		return ErrFrozenWallet
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return fmt.Errorf("billing client: transient %d", resp.StatusCode)
	default:
		return fmt.Errorf("billing client: unexpected status %d", resp.StatusCode)
	}
}

// buildBody сериализует body для /wallet/spend или /wallet/credit.
//
// /wallet/spend ждёт {amount, reason, ref_id, to_account, currency?}.
// /wallet/credit ждёт {amount, reason, ref_id, from_account, currency?}.
// Здесь совмещаем: для refund поле trans-direction'a — это from_account,
// которое мы выводим из in.ToAccount (логически тот же system-account,
// просто роль меняется).
func buildBody(op string, in SpendInput) ([]byte, error) {
	payload := map[string]any{
		"amount": in.Amount,
		"reason": in.Reason,
		"ref_id": in.RefID,
	}
	switch op {
	case "spend":
		payload["to_account"] = in.ToAccount
	case "refund":
		payload["from_account"] = in.ToAccount
	default:
		return nil, fmt.Errorf("billing client: unknown op %q", op)
	}
	return json.Marshal(payload)
}

// isTransient распознаёт сетевые / временные ошибки, на которых retry
// имеет смысл. Caller-cancelled context отсеивается выше — здесь работаем
// только с ошибками HTTP-уровня и client-timeout'ом.
func isTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	// http.Client.Timeout проявляется как url.Error c DeadlineExceeded
	// внутри. Это transient-сбой billing (медленный ответ), а не caller-
	// cancellation (caller-cancel мы поймали выше через ctx.Err()).
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	// HTTP 5xx-сообщение от attempt() явно содержит "transient" — простая
	// маркировка без введения нового sentinel-типа.
	return strings.Contains(err.Error(), "transient ")
}

// incSpendStatus / observeDuration — обёртки над metrics-пакетом, безопасные
// до Register() (тогда no-op): не падать в тестах, где metrics не подняты.
func incSpendStatus(op, status string) {
	if metrics.BillingClientSpend == nil {
		return
	}
	metrics.BillingClientSpend.WithLabelValues(op, status).Inc()
}

func observeDuration(op string, d time.Duration) {
	if metrics.BillingClientDuration == nil {
		return
	}
	metrics.BillingClientDuration.WithLabelValues(op).Observe(d.Seconds())
}
