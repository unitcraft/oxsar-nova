package payment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// MockGateway — фиктивный шлюз для E2E-тестов. Без подписей и внешних
// HTTP-вызовов. BuildPayURL возвращает ссылку на внутренний эндпоинт-симулятор,
// который сразу же дёргает webhook и редиректит игрока обратно.
//
// Активируется через PAYMENT_PROVIDER=mock. Использовать только в dev/test.
type MockGateway struct {
	// BaseURL — префикс сервера, к которому будут обращаться mock pay/webhook.
	// Пусто → относительный URL (фронт на том же origin — ок).
	BaseURL string
}

func NewMockGateway(baseURL string) *MockGateway {
	return &MockGateway{BaseURL: baseURL}
}

// BuildPayURL возвращает ссылку вида {base}/api/payment/mock/pay?order=...&amount=...&result=success.
// Фронт кликает → бэкенд (mock-handler) обрабатывает платёж и редиректит на ReturnURL.
func (g *MockGateway) BuildPayURL(_ context.Context, orderID, _ string, amountKop int, returnURL string) (string, error) {
	q := url.Values{}
	q.Set("order", orderID)
	q.Set("amount", strconv.Itoa(amountKop))
	q.Set("result", "success") // игрок может вручную поменять на fail в URL
	if returnURL != "" {
		q.Set("return", returnURL)
	}
	return g.BaseURL + "/api/payment/mock/pay?" + q.Encode(), nil
}

// VerifyWebhook — без проверки подписи. Читает order_id и amount_kop из формы.
// Используется как путь подтверждения: mock-handler сам дёргает ConfirmPayment,
// но интерфейс сохраняем, чтобы совпадать с другими Gateway.
func (g *MockGateway) VerifyWebhook(r *http.Request) (string, int, error) {
	if err := r.ParseForm(); err != nil {
		return "", 0, ErrWebhookInvalid
	}
	orderID := r.FormValue("order_id")
	if orderID == "" {
		orderID = r.FormValue("InvId")
	}
	if orderID == "" {
		return "", 0, ErrWebhookInvalid
	}
	amountStr := r.FormValue("amount_kop")
	amount, _ := strconv.Atoi(amountStr)
	return orderID, amount, nil
}

// SuccessResponse — ответ, аналогичный Робокассе: "OK{orderID}".
func (g *MockGateway) SuccessResponse(w http.ResponseWriter, orderID string) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "OK%s", orderID)
}

// IsMock — признак mock-режима для UI (баннер «Тестовый режим»).
func (g *MockGateway) IsMock() bool { return true }
