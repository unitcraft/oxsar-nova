package portalsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// BillingClient вызывает billing-service для списания кредитов.
//
// План 38 Ф.6: portal больше не дёргает identity-service для credits — кошельки
// переехали в billing-service.
//
// Использует тот же RSA-JWT, что прислал клиент (forwarded через Authorization
// header). Idempotency-Key привязывается к (user, feedback_id) — повторный
// клик по «голосовать» не списывает второй раз.
type BillingClient struct {
	billingURL string
	httpClient *http.Client
}

// NewBillingClient создаёт клиент для работы с billing.
func NewBillingClient(billingURL string) *BillingClient {
	return &BillingClient{
		billingURL: billingURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SpendInput — параметры списания.
type SpendInput struct {
	UserToken      string // RSA-JWT юзера (forwarded из исходного запроса)
	Amount         int64
	Reason         string
	RefID          string // feedback_id
	ToAccount      string // 'vote:feedback:<id>'
	IdempotencyKey string
}

// Spend списывает amount кредитов через billing-service.
//
// Возвращает ErrInsufficientCredits при HTTP 402 — это ошибка домена
// (кредитов недостаточно), её portal должен прокинуть клиенту как 402.
func (c *BillingClient) Spend(ctx context.Context, in SpendInput) error {
	if c.billingURL == "" {
		return errors.New("portalsvc: billing URL not configured")
	}
	body, _ := json.Marshal(map[string]any{
		"amount":     in.Amount,
		"reason":     in.Reason,
		"ref_id":     in.RefID,
		"to_account": in.ToAccount,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.billingURL+"/billing/wallet/spend", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+in.UserToken)
	if in.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", in.IdempotencyKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusPaymentRequired:
		return ErrInsufficientCredits
	default:
		return fmt.Errorf("billing /wallet/spend: status %d", resp.StatusCode)
	}
}
