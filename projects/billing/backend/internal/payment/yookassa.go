package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// План 42: YooKassa как основной платёжный шлюз.
//
// Документация: https://yookassa.ru/developers/api
//
// Особенности YooKassa, существенные для архитектуры:
//
//   1. confirmation_url возвращается в ответе POST /v3/payments,
//      а не строится локально. Поэтому BuildPayURL делает реальный HTTP-вызов.
//
//   2. Webhook НЕ подписывается. Безопасность даётся:
//      a) IP-allowlist (списки CIDR от YooKassa);
//      b) re-fetch payment по id через GET /v3/payments/{id} — подтверждаем
//         статус у самой YooKassa и не верим только webhook-телу.
//      Это сознательное решение YooKassa, описанное в их документации.
//
//   3. Чек ФНС для самозанятых (НПД): передаём receipt в CreatePayment,
//      YooKassa автоматически передаёт его в "Мой налог". vat_code=1
//      (без НДС), payment_subject=service.
//
//   4. Идемпотентность создания платежа: header `Idempotence-Key`. Используем
//      order_id (uuid) — повторный POST с тем же ключом возвращает
//      существующий payment.
//
//   5. Авторизация: HTTP Basic Auth (shopId:secretKey).
//
// Конфиг:
//   - YOOKASSA_SHOP_ID         — ID магазина в ЛК YooKassa
//   - YOOKASSA_SECRET_KEY      — secret ключ
//   - YOOKASSA_API_URL         — обычно https://api.yookassa.ru/v3 (можно переопределить для тестов)
//   - YOOKASSA_RETURN_URL      — куда YooKassa редиректит юзера после оплаты
//   - YOOKASSA_VERIFY_FETCH    — "1" по умолчанию: re-fetch для верификации webhook
//   - YOOKASSA_TRUSTED_PROXIES — список CIDR/IP для IP-allowlist (если за nginx)
//
// Тестовый магазин YooKassa имеет отдельные shopId/secretKey — на dev-стенде
// создаём пустой magazin или используем mock-провайдер.

// YooKassaTrustedNetworks — официальные подсети YooKassa для webhook'ов.
// Источник: https://yookassa.ru/developers/using-api/webhooks
var YooKassaTrustedNetworks = []string{
	"185.71.76.0/27",
	"185.71.77.0/27",
	"77.75.153.0/25",
	"77.75.156.11/32",
	"77.75.156.35/32",
	"77.75.154.128/25",
	"2a02:5180::/32",
}

// YooKassaGateway реализует payment.Gateway для YooKassa.
type YooKassaGateway struct {
	ShopID     string
	SecretKey  string
	APIURL     string // base, например "https://api.yookassa.ru/v3"
	ReturnURL  string
	HTTPClient *http.Client

	// VerifyByFetch — при получении webhook повторно запрашивать платёж
	// у YooKassa и сверять статус (защита от подделки webhook).
	// Включено по умолчанию.
	VerifyByFetch bool

	// trustedNets — закэшированные parsed IP-сети для IP-allowlist.
	// Если len == 0 — IP-проверка отключена (для тестов и dev).
	trustedNets []*net.IPNet
}

// NewYooKassaGateway создаёт gateway. apiURL пустой — используется prod
// "https://api.yookassa.ru/v3". В dev можно подсунуть mock-сервер.
func NewYooKassaGateway(shopID, secretKey, apiURL, returnURL string) *YooKassaGateway {
	if apiURL == "" {
		apiURL = "https://api.yookassa.ru/v3"
	}
	g := &YooKassaGateway{
		ShopID:        shopID,
		SecretKey:     secretKey,
		APIURL:        strings.TrimRight(apiURL, "/"),
		ReturnURL:     returnURL,
		HTTPClient:    &http.Client{Timeout: 15 * time.Second},
		VerifyByFetch: true,
	}
	g.SetTrustedNetworks(YooKassaTrustedNetworks)
	return g
}

// SetTrustedNetworks парсит CIDR-список и кэширует. Передавать nil чтобы
// отключить IP-allowlist (только для тестов).
func (g *YooKassaGateway) SetTrustedNetworks(cidrs []string) {
	g.trustedNets = nil
	for _, c := range cidrs {
		// Поддерживаем чистые IP без /32 префикса.
		if !strings.Contains(c, "/") {
			if strings.Contains(c, ":") {
				c += "/128"
			} else {
				c += "/32"
			}
		}
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			continue
		}
		g.trustedNets = append(g.trustedNets, n)
	}
}

func (g *YooKassaGateway) Name() string { return "yookassa" }

// receiptItem — позиция в чеке ФНС (для НПД).
type yookassaReceiptItem struct {
	Description    string                  `json:"description"`
	Quantity       string                  `json:"quantity"`
	Amount         yookassaAmount          `json:"amount"`
	VATCode        int                     `json:"vat_code"`
	PaymentSubject string                  `json:"payment_subject"`
	PaymentMode    string                  `json:"payment_mode"`
}

type yookassaCustomer struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type yookassaReceipt struct {
	Customer yookassaCustomer       `json:"customer"`
	Items    []yookassaReceiptItem  `json:"items"`
}

type yookassaAmount struct {
	Value    string `json:"value"`    // "500.00"
	Currency string `json:"currency"` // "RUB"
}

type yookassaConfirmation struct {
	Type      string `json:"type"`       // "redirect"
	ReturnURL string `json:"return_url"`
}

type yookassaCreateRequest struct {
	Amount       yookassaAmount       `json:"amount"`
	Capture      bool                 `json:"capture"`
	Confirmation yookassaConfirmation `json:"confirmation"`
	Description  string               `json:"description"`
	Metadata     map[string]string    `json:"metadata"`
	Receipt      *yookassaReceipt     `json:"receipt,omitempty"`
}

type yookassaConfirmationResp struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
}

type yookassaPayment struct {
	ID           string                  `json:"id"`
	Status       string                  `json:"status"` // "pending" | "waiting_for_capture" | "succeeded" | "canceled"
	Paid         bool                    `json:"paid"`
	Amount       yookassaAmount          `json:"amount"`
	Confirmation *yookassaConfirmationResp `json:"confirmation"`
	Metadata     map[string]string       `json:"metadata"`
}

type yookassaErrorResp struct {
	Type        string `json:"type"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// BuildPayURL создаёт платёж в YooKassa и возвращает confirmation_url.
//
// Идемпотентность: Idempotence-Key = orderID. Повторный вызов с тем же
// orderID возвращает существующий payment (а не создаёт второй).
//
// Чек: если customerEmail/customerPhone заданы через метаданные — добавляем
// receipt с одной позицией "Игровые кредиты, пакет N". Без email/phone
// чек не отправится (требование ФНС для НПД).
//
// На текущей сигнатуре Gateway.BuildPayURL не имеем доступ к email/phone
// юзера — для полноценного НПД-чека нужно расширение интерфейса
// (BuildPayURL → CreateOrder с дополнительными параметрами). Пока чек
// строится только если email пробрасывается через переменную окружения
// или приходит позже через CreateOrder-расширение. См. план 42 §2.
func (g *YooKassaGateway) BuildPayURL(ctx context.Context, orderID, _ string, amountKop int64, returnURL string) (string, error) {
	if g.ShopID == "" || g.SecretKey == "" {
		return "", errors.New("yookassa: shop_id and secret_key are required")
	}
	ru := returnURL
	if ru == "" {
		ru = g.ReturnURL
	}
	req := yookassaCreateRequest{
		Amount:  yookassaAmount{Value: kopToRubString(amountKop), Currency: "RUB"},
		Capture: true,
		Confirmation: yookassaConfirmation{
			Type:      "redirect",
			ReturnURL: ru,
		},
		Description: fmt.Sprintf("Покупка кредитов OXC, заказ %s", orderID),
		Metadata:    map[string]string{"order_id": orderID},
		// Receipt пока nil. Чек НПД добавляется при расширенной CreateOrder
		// с email/phone клиента (план 42 §2). Без email YooKassa
		// откажет в чеке (для НПД email/phone обязательны).
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.APIURL+"/payments", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", orderID)
	httpReq.SetBasicAuth(g.ShopID, g.SecretKey)

	resp, err := g.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("yookassa request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if resp.StatusCode >= 400 {
		var ye yookassaErrorResp
		_ = json.Unmarshal(respBody, &ye)
		return "", fmt.Errorf("yookassa %d: %s %s", resp.StatusCode, ye.Code, ye.Description)
	}
	var p yookassaPayment
	if err := json.Unmarshal(respBody, &p); err != nil {
		return "", fmt.Errorf("yookassa unmarshal: %w", err)
	}
	if p.Confirmation == nil || p.Confirmation.ConfirmationURL == "" {
		return "", errors.New("yookassa: empty confirmation_url in response")
	}
	return p.Confirmation.ConfirmationURL, nil
}

// VerifyWebhook парсит webhook event от YooKassa.
//
// YooKassa НЕ подписывает webhook. Безопасность:
//   1. IP allowlist (set up через SetTrustedNetworks).
//   2. Re-fetch платежа через GET /v3/payments/{id} (если VerifyByFetch=true).
//
// Возвращает (orderID, amountKop, err) согласно интерфейсу Gateway.
// orderID — это metadata.order_id, который мы передавали при CreatePayment.
func (g *YooKassaGateway) VerifyWebhook(r *http.Request, body []byte) (string, int64, error) {
	// 1. IP allowlist (если настроен).
	if len(g.trustedNets) > 0 {
		clientIP := extractClientIP(r)
		ip := net.ParseIP(clientIP)
		if ip == nil {
			return "", 0, fmt.Errorf("%w: bad client IP %q", ErrWebhookInvalid, clientIP)
		}
		ok := false
		for _, n := range g.trustedNets {
			if n.Contains(ip) {
				ok = true
				break
			}
		}
		if !ok {
			return "", 0, fmt.Errorf("%w: client IP %s not in YooKassa allowlist",
				ErrSignatureMismatch, clientIP)
		}
	}

	// 2. Парсинг события.
	var event struct {
		Type   string          `json:"type"`  // "notification"
		Event  string          `json:"event"` // "payment.succeeded" | "payment.canceled" | ...
		Object yookassaPayment `json:"object"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return "", 0, fmt.Errorf("%w: %v", ErrWebhookInvalid, err)
	}
	if event.Object.ID == "" {
		return "", 0, fmt.Errorf("%w: missing payment.id", ErrWebhookInvalid)
	}
	orderID := event.Object.Metadata["order_id"]
	if orderID == "" {
		return "", 0, fmt.Errorf("%w: missing metadata.order_id", ErrWebhookInvalid)
	}

	// 3. Принимаем только успешные оплаты. Все остальные события
	// (canceled, waiting_for_capture) — webhook прилетел, но платежа НЕТ.
	// Возвращаем специфическую ошибку, чтобы handler не трактовал как top-up.
	if event.Event != "payment.succeeded" {
		return "", 0, fmt.Errorf("%w: event %q is not payment.succeeded",
			ErrWebhookInvalid, event.Event)
	}

	// 4. Re-fetch — самое надёжное подтверждение.
	if g.VerifyByFetch {
		fetched, err := g.fetchPayment(r.Context(), event.Object.ID)
		if err != nil {
			return "", 0, fmt.Errorf("yookassa fetch payment: %w", err)
		}
		if fetched.Status != "succeeded" || !fetched.Paid {
			return "", 0, fmt.Errorf("%w: re-fetch shows status=%s paid=%v",
				ErrWebhookInvalid, fetched.Status, fetched.Paid)
		}
		// Используем сумму из re-fetch — она авторитетна.
		event.Object.Amount = fetched.Amount
	}

	amountKop, err := rubStringToKop(event.Object.Amount.Value)
	if err != nil {
		return "", 0, fmt.Errorf("%w: amount %v", ErrWebhookInvalid, err)
	}
	return orderID, amountKop, nil
}

// fetchPayment делает GET /v3/payments/{id} для верификации.
func (g *YooKassaGateway) fetchPayment(ctx context.Context, paymentID string) (*yookassaPayment, error) {
	if g.ShopID == "" || g.SecretKey == "" {
		return nil, errors.New("yookassa: credentials missing")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		g.APIURL+"/payments/"+paymentID, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(g.ShopID, g.SecretKey)
	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yookassa fetch %d: %s", resp.StatusCode, string(respBody))
	}
	var p yookassaPayment
	if err := json.Unmarshal(respBody, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// SuccessResponse — YooKassa ожидает 200 без тела (любого содержимого).
func (g *YooKassaGateway) SuccessResponse(w http.ResponseWriter, _ string) {
	w.WriteHeader(http.StatusOK)
}

// kopToRubString конвертирует копейки в "500.00" с двумя знаками.
func kopToRubString(kop int64) string {
	rub := kop / 100
	rem := kop % 100
	if rem < 0 {
		rem = -rem
	}
	return fmt.Sprintf("%d.%02d", rub, rem)
}

// rubStringToKop парсит "500.00" → 50000 копеек. Терпит "500", "500.5",
// "500.50". Отрицательные — ошибка.
func rubStringToKop(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty amount")
	}
	parts := strings.SplitN(s, ".", 2)
	rub, err := atoi64(parts[0])
	if err != nil || rub < 0 {
		return 0, fmt.Errorf("bad rubles: %q", s)
	}
	var kop int64
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 2 {
			frac = frac[:2]
		} else if len(frac) == 1 {
			frac += "0"
		}
		kop, err = atoi64(frac)
		if err != nil || kop < 0 {
			return 0, fmt.Errorf("bad kopecks: %q", s)
		}
	}
	return rub*100 + kop, nil
}

func atoi64(s string) (int64, error) {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a digit: %q", c)
		}
		n = n*10 + int64(c-'0')
	}
	return n, nil
}

// extractClientIP — IP клиента с учётом X-Forwarded-For (один прокси-хоп).
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// _ — proof that uuid is used (поле получаем извне для idempotence-key,
// но если хотим внутри — сюда).
var _ = uuid.NewString
