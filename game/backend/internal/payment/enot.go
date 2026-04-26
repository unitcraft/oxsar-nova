package payment

import (
	"context"
	"crypto/md5" //nolint:gosec — Enot.io использует MD5 для подписи
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// EnotGateway — второй провайдер оплаты (Enot.io). Используется как
// резервный к Робокассе: если у одной проблемы с платежами, переключаем
// PAYMENT_PROVIDER без релиза.
//
// Протокол Enot.io v1 (упрощённо):
//   - Создание платежа: GET https://enot.io/pay?m=SHOP&oa=SUM&o=ORDER&s=SIG&cr=RUB
//     где s = MD5(SHOP:SUM:PASS1:ORDER) — по документации Enot.io.
//   - Webhook: POST с form/json. Поля: merchant, amount, order_id, sign_2.
//     sign_2 = MD5(merchant:amount:PASS2:order_id).
//
// Секретами управляют ENOT_SHOP_ID, ENOT_API_KEY. ENOT_API_KEY здесь
// играет роль pass1 (для создания платежа) и pass2 (для верификации
// webhook) — Enot традиционно использует одно поле `secret`, часто
// разделяя на 2 в личном кабинете. Для простоты единый ENOT_API_KEY.
const enotBaseURL = "https://enot.io/pay"

type EnotGateway struct {
	shopID string
	secret string
}

func NewEnotGateway(shopID, secret string) *EnotGateway {
	return &EnotGateway{shopID: shopID, secret: secret}
}

// BuildPayURL строит URL на страницу оплаты Enot.
func (g *EnotGateway) BuildPayURL(_ context.Context, orderID, description string, amountKop int, returnURL string) (string, error) {
	amount := fmt.Sprintf("%.2f", float64(amountKop)/100)
	sig := enotMD5(g.shopID, amount, g.secret, orderID)

	q := url.Values{}
	q.Set("m", g.shopID)
	q.Set("oa", amount)
	q.Set("o", orderID)
	q.Set("s", sig)
	q.Set("cr", "RUB")
	if description != "" {
		q.Set("i", description)
	}
	if returnURL != "" {
		q.Set("success_url", returnURL)
	}
	return enotBaseURL + "?" + q.Encode(), nil
}

// VerifyWebhook разбирает callback от Enot. Поддерживает form-urlencoded
// и JSON-тело (Enot отдаёт form по умолчанию, но некоторые интеграции
// отправляют JSON).
func (g *EnotGateway) VerifyWebhook(r *http.Request) (orderID string, amountKop int, err error) {
	merchant := ""
	amount := ""
	sig := ""

	if ct := r.Header.Get("Content-Type"); strings.Contains(ct, "application/json") {
		var body struct {
			Merchant string `json:"merchant"`
			Amount   string `json:"amount"`
			OrderID  string `json:"order_id"`
			Sign2    string `json:"sign_2"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return "", 0, ErrWebhookInvalid
		}
		merchant = body.Merchant
		amount = body.Amount
		orderID = body.OrderID
		sig = body.Sign2
	} else {
		if err := r.ParseForm(); err != nil {
			return "", 0, ErrWebhookInvalid
		}
		merchant = r.FormValue("merchant")
		amount = r.FormValue("amount")
		orderID = r.FormValue("order_id")
		sig = r.FormValue("sign_2")
	}
	if merchant == "" || amount == "" || orderID == "" || sig == "" {
		return "", 0, ErrWebhookInvalid
	}
	if merchant != g.shopID {
		return "", 0, ErrWebhookInvalid
	}
	expected := enotMD5(merchant, amount, g.secret, orderID)
	if !strings.EqualFold(expected, sig) {
		return "", 0, ErrWebhookInvalid
	}
	var rub float64
	if _, scanErr := fmt.Sscanf(amount, "%f", &rub); scanErr != nil {
		return "", 0, ErrWebhookInvalid
	}
	return orderID, int(rub * 100), nil
}

// SuccessResponse — Enot ожидает ответ с текстом `success` для
// подтверждения обработки (иначе шлёт повторные webhook'и).
func (g *EnotGateway) SuccessResponse(w http.ResponseWriter, _ string) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprint(w, "success")
}

func enotMD5(parts ...string) string {
	h := md5.New() //nolint:gosec
	h.Write([]byte(strings.Join(parts, ":")))
	return fmt.Sprintf("%x", h.Sum(nil))
}
