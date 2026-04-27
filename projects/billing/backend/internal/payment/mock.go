package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MockGateway — fake payment gateway для dev/test.
//
// В отличие от тривиальной заглушки, делает HMAC-подпись webhook-payload
// тем же секретом, что и настоящий шлюз. Это:
//   - Позволяет E2E-тестам проверять signature-verify (тот же код, что в проде).
//   - Защищает от случайного запуска с PAYMENT_PROVIDER=mock на проде:
//     если кто-то сломал Mock-flow, обманные webhook'и будут отклонены.
//
// В проде НЕ используется (хоть и не опасен сам по себе).
type MockGateway struct {
	BaseURL   string // префикс для построения pay/webhook URL
	Secret    []byte // HMAC ключ
	Clock     func() time.Time
}

// NewMockGateway создаёт Mock с заданным секретом. Секрет конфигурируется
// через PAYMENT_MOCK_SECRET (env). Если пуст — генерируется из BaseURL
// (только для CI/test, в prod должен быть явно задан).
func NewMockGateway(baseURL, secret string) *MockGateway {
	if secret == "" {
		secret = "mock-default-secret-do-not-use-in-prod"
	}
	return &MockGateway{
		BaseURL: baseURL,
		Secret:  []byte(secret),
		Clock:   func() time.Time { return time.Now().UTC() },
	}
}

func (g *MockGateway) Name() string { return "mock" }

// BuildPayURL возвращает ссылку на симулятор:
//   {base}/api/payment/mock/pay?order=...&amount=...&return=...
//
// Mock pay-handler должен:
//   1. Вернуть простую HTML-страницу «Симулятор оплаты» с кнопками success/fail.
//   2. При success — POST в /billing/webhooks/mock с подписанным payload.
//   3. Редирект на returnURL.
//
// Сам pay-handler в коде НЕ реализован для billing-service — если нужно,
// добавляется в Ф.3.5 для удобной E2E-демки. В тестах webhook вызывается
// напрямую через curl/http.Post.
func (g *MockGateway) BuildPayURL(_ context.Context, orderID, _ string, amountKop int64, returnURL string) (string, error) {
	q := url.Values{}
	q.Set("order", orderID)
	q.Set("amount", strconv.FormatInt(amountKop, 10))
	q.Set("result", "success")
	if returnURL != "" {
		q.Set("return", returnURL)
	}
	return g.BaseURL + "/api/payment/mock/pay?" + q.Encode(), nil
}

// VerifyWebhook парсит webhook payload (form-encoded), проверяет HMAC и timestamp.
//
// Ожидаемые поля:
//   order_id  — UUID заказа
//   amount    — сумма в копейках
//   ts        — Unix timestamp UTC (для replay protection, окно ±5 минут)
//   signature — hex(HMAC-SHA256(secret, "<order_id>|<amount>|<ts>"))
func (g *MockGateway) VerifyWebhook(r *http.Request, body []byte) (string, int64, error) {
	// Парсим form-data вручную, потому что body уже прочитан middleware-ом.
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", 0, fmt.Errorf("%w: parse: %v", ErrWebhookInvalid, err)
	}
	orderID := values.Get("order_id")
	amountStr := values.Get("amount")
	tsStr := values.Get("ts")
	sig := values.Get("signature")
	if orderID == "" || amountStr == "" || tsStr == "" || sig == "" {
		return "", 0, fmt.Errorf("%w: missing fields", ErrWebhookInvalid)
	}
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("%w: amount: %v", ErrWebhookInvalid, err)
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("%w: ts: %v", ErrWebhookInvalid, err)
	}
	// Replay protection: ±5 минут от now.
	now := g.Clock().Unix()
	if abs(now-ts) > 5*60 {
		return "", 0, ErrTimestampOld
	}
	// HMAC verify.
	expected := g.Sign(orderID, amount, ts)
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", 0, ErrSignatureMismatch
	}
	return orderID, amount, nil
}

// Sign формирует HMAC-подпись для payload (используется test-helper-ами
// и mock pay-handler-ом для имитации webhook'а).
func (g *MockGateway) Sign(orderID string, amount, ts int64) string {
	mac := hmac.New(sha256.New, g.Secret)
	fmt.Fprintf(mac, "%s|%d|%d", orderID, amount, ts)
	return hex.EncodeToString(mac.Sum(nil))
}

func (g *MockGateway) SuccessResponse(w http.ResponseWriter, _ string) {
	w.WriteHeader(http.StatusNoContent)
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
