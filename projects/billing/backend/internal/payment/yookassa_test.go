package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeYooKassaServer имитирует YooKassa API: POST /v3/payments создаёт
// payment, GET /v3/payments/{id} возвращает его. Идемпотентность по
// Idempotence-Key-header.
type fakeYooKassaServer struct {
	server   *httptest.Server
	payments map[string]yookassaPayment
	idemKeys map[string]string // idem-key → payment.id (для повторов)
	requests []recordedRequest
}

type recordedRequest struct {
	Method string
	Path   string
	IdemKey string
	Auth    string
	Body    string
}

func newFakeYooKassa() *fakeYooKassaServer {
	f := &fakeYooKassaServer{
		payments: map[string]yookassaPayment{},
		idemKeys: map[string]string{},
	}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	return f
}

func (f *fakeYooKassaServer) URL() string { return f.server.URL }
func (f *fakeYooKassaServer) Close()      { f.server.Close() }

func (f *fakeYooKassaServer) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.requests = append(f.requests, recordedRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		IdemKey: r.Header.Get("Idempotence-Key"),
		Auth:    r.Header.Get("Authorization"),
		Body:    string(body),
	})

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/payments":
		idem := r.Header.Get("Idempotence-Key")
		if existingID, ok := f.idemKeys[idem]; ok {
			// Повтор — возвращаем существующий payment.
			p := f.payments[existingID]
			_ = json.NewEncoder(w).Encode(p)
			return
		}
		var req yookassaCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		paymentID := fmt.Sprintf("test-%d", len(f.payments)+1)
		p := yookassaPayment{
			ID:     paymentID,
			Status: "pending",
			Paid:   false,
			Amount: req.Amount,
			Confirmation: &yookassaConfirmationResp{
				Type:            "redirect",
				ConfirmationURL: f.server.URL + "/checkout/" + paymentID,
			},
			Metadata: req.Metadata,
		}
		f.payments[paymentID] = p
		f.idemKeys[idem] = paymentID
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(p)

	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/payments/"):
		id := strings.TrimPrefix(r.URL.Path, "/payments/")
		p, ok := f.payments[id]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(p)

	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// markPaid — тестовый helper, который меняет статус платежа на 'succeeded'
// (как будто YooKassa провела платёж и сейчас прислала бы webhook).
func (f *fakeYooKassaServer) markPaid(paymentID string) {
	p := f.payments[paymentID]
	p.Status = "succeeded"
	p.Paid = true
	f.payments[paymentID] = p
}

// TestYooKassa_BuildPayURL проверяет CreatePayment-flow:
//   - HTTP POST /payments с правильным телом и Basic Auth
//   - Idempotence-Key передаётся
//   - В ответе → confirmation_url отдан клиенту
func TestYooKassa_BuildPayURL(t *testing.T) {
	srv := newFakeYooKassa()
	defer srv.Close()

	gw := NewYooKassaGateway("shop_123", "secret_xyz", srv.URL(), "https://app/return")
	gw.SetTrustedNetworks(nil) // в тестах IP-allowlist не нужен

	url, err := gw.BuildPayURL(context.Background(), "order-uuid-1", "user-1", 50000, "")
	if err != nil {
		t.Fatalf("BuildPayURL: %v", err)
	}
	if !strings.HasPrefix(url, srv.URL()+"/checkout/") {
		t.Errorf("confirmation_url=%q, want prefix %s/checkout/", url, srv.URL())
	}

	if len(srv.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(srv.requests))
	}
	req := srv.requests[0]
	if req.Method != "POST" || req.Path != "/payments" {
		t.Errorf("req=%s %s, want POST /payments", req.Method, req.Path)
	}
	if req.IdemKey != "order-uuid-1" {
		t.Errorf("Idempotence-Key=%q, want order-uuid-1", req.IdemKey)
	}
	if !strings.HasPrefix(req.Auth, "Basic ") {
		t.Errorf("auth=%q, want Basic ...", req.Auth)
	}
	// Тело запроса должно содержать amount, capture, confirmation,
	// metadata.order_id.
	var sent yookassaCreateRequest
	if err := json.Unmarshal([]byte(req.Body), &sent); err != nil {
		t.Fatalf("unmarshal sent body: %v", err)
	}
	if sent.Amount.Value != "500.00" {
		t.Errorf("amount.value=%q, want 500.00", sent.Amount.Value)
	}
	if sent.Amount.Currency != "RUB" {
		t.Errorf("amount.currency=%q, want RUB", sent.Amount.Currency)
	}
	if !sent.Capture {
		t.Errorf("capture=false, want true")
	}
	if sent.Confirmation.ReturnURL != "https://app/return" {
		t.Errorf("return_url=%q, want https://app/return", sent.Confirmation.ReturnURL)
	}
	if sent.Metadata["order_id"] != "order-uuid-1" {
		t.Errorf("metadata.order_id=%q, want order-uuid-1", sent.Metadata["order_id"])
	}
}

// TestYooKassa_BuildPayURL_Idempotency проверяет что повторный вызов с тем
// же orderID не создаёт второй payment.
func TestYooKassa_BuildPayURL_Idempotency(t *testing.T) {
	srv := newFakeYooKassa()
	defer srv.Close()
	gw := NewYooKassaGateway("shop", "secret", srv.URL(), "https://app/return")
	gw.SetTrustedNetworks(nil)

	url1, err := gw.BuildPayURL(context.Background(), "order-A", "user-1", 1000, "")
	if err != nil {
		t.Fatal(err)
	}
	url2, err := gw.BuildPayURL(context.Background(), "order-A", "user-1", 1000, "")
	if err != nil {
		t.Fatal(err)
	}
	if url1 != url2 {
		t.Errorf("idempotency broken: url1=%q url2=%q", url1, url2)
	}
	if len(srv.payments) != 1 {
		t.Errorf("expected 1 payment in storage, got %d", len(srv.payments))
	}
}

// TestYooKassa_VerifyWebhook_Success проверяет happy-path:
//   - YooKassa → POST /webhook с {event:"payment.succeeded", object:...}
//   - re-fetch подтверждает status=succeeded
//   - Возвращается (orderID, amount).
func TestYooKassa_VerifyWebhook_Success(t *testing.T) {
	srv := newFakeYooKassa()
	defer srv.Close()
	gw := NewYooKassaGateway("shop", "secret", srv.URL(), "")
	gw.SetTrustedNetworks(nil) // disable IP allowlist for unit-test

	// Сначала создадим платёж, потом «оплатим» (ставим status=succeeded).
	if _, err := gw.BuildPayURL(context.Background(), "order-X", "user-1", 50000, ""); err != nil {
		t.Fatal(err)
	}
	srv.markPaid("test-1")

	// Имитируем webhook от YooKassa.
	hook := map[string]any{
		"type":  "notification",
		"event": "payment.succeeded",
		"object": yookassaPayment{
			ID:       "test-1",
			Status:   "succeeded",
			Paid:     true,
			Amount:   yookassaAmount{Value: "500.00", Currency: "RUB"},
			Metadata: map[string]string{"order_id": "order-X"},
		},
	}
	body, _ := json.Marshal(hook)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.RemoteAddr = "1.2.3.4:1000"

	orderID, amountKop, err := gw.VerifyWebhook(req, body)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if orderID != "order-X" {
		t.Errorf("orderID=%q, want order-X", orderID)
	}
	if amountKop != 50000 {
		t.Errorf("amount=%d, want 50000", amountKop)
	}
}

// TestYooKassa_VerifyWebhook_RejectsCanceled проверяет что только
// payment.succeeded трактуется как настоящий top-up.
func TestYooKassa_VerifyWebhook_RejectsCanceled(t *testing.T) {
	srv := newFakeYooKassa()
	defer srv.Close()
	gw := NewYooKassaGateway("shop", "secret", srv.URL(), "")
	gw.SetTrustedNetworks(nil)

	hook := map[string]any{
		"type":  "notification",
		"event": "payment.canceled",
		"object": yookassaPayment{
			ID:       "test-X",
			Status:   "canceled",
			Metadata: map[string]string{"order_id": "order-Y"},
		},
	}
	body, _ := json.Marshal(hook)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	_, _, err := gw.VerifyWebhook(req, body)
	if err == nil || !errors.Is(err, ErrWebhookInvalid) {
		t.Errorf("err=%v, want ErrWebhookInvalid", err)
	}
}

// TestYooKassa_VerifyWebhook_RefetchMismatch проверяет: если webhook говорит
// succeeded, а re-fetch возвращает pending — отклоняем (возможная подделка).
func TestYooKassa_VerifyWebhook_RefetchMismatch(t *testing.T) {
	srv := newFakeYooKassa()
	defer srv.Close()
	gw := NewYooKassaGateway("shop", "secret", srv.URL(), "")
	gw.SetTrustedNetworks(nil)

	if _, err := gw.BuildPayURL(context.Background(), "order-Z", "user", 1000, ""); err != nil {
		t.Fatal(err)
	}
	// НЕ ставим status=succeeded в fake-сервере → re-fetch вернёт pending.

	hook := map[string]any{
		"type":  "notification",
		"event": "payment.succeeded",
		"object": yookassaPayment{
			ID:       "test-1",
			Status:   "succeeded", // фальшивая инфа в webhook
			Paid:     true,
			Amount:   yookassaAmount{Value: "10.00", Currency: "RUB"},
			Metadata: map[string]string{"order_id": "order-Z"},
		},
	}
	body, _ := json.Marshal(hook)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	_, _, err := gw.VerifyWebhook(req, body)
	if err == nil {
		t.Errorf("expected error from refetch mismatch, got nil")
	}
}

// TestYooKassa_VerifyWebhook_IPAllowlist проверяет что webhook от
// неразрешённого IP отклоняется.
func TestYooKassa_VerifyWebhook_IPAllowlist(t *testing.T) {
	srv := newFakeYooKassa()
	defer srv.Close()
	gw := NewYooKassaGateway("shop", "secret", srv.URL(), "")
	// Разрешаем только 10.0.0.0/24, реальные YooKassa IP не вкладываются.
	gw.SetTrustedNetworks([]string{"10.0.0.0/24"})

	hook := map[string]any{
		"type":   "notification",
		"event":  "payment.succeeded",
		"object": yookassaPayment{ID: "x", Metadata: map[string]string{"order_id": "y"}},
	}
	body, _ := json.Marshal(hook)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.RemoteAddr = "1.2.3.4:1000"

	_, _, err := gw.VerifyWebhook(req, body)
	if err == nil || !errors.Is(err, ErrSignatureMismatch) {
		t.Errorf("err=%v, want ErrSignatureMismatch (IP not allowed)", err)
	}

	// А с правильным X-Forwarded-For — пропускаем (потом упадём на чём-то другом).
	req2 := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req2.Header.Set("X-Forwarded-For", "10.0.0.42")
	_, _, err = gw.VerifyWebhook(req2, body)
	if err != nil && errors.Is(err, ErrSignatureMismatch) {
		t.Errorf("IP 10.0.0.42 should be allowed, got %v", err)
	}
}

// TestKopToRubString проверяет конвертер.
func TestKopToRubString(t *testing.T) {
	cases := []struct {
		kop  int64
		want string
	}{
		{0, "0.00"},
		{1, "0.01"},
		{99, "0.99"},
		{100, "1.00"},
		{50000, "500.00"},
		{50001, "500.01"},
		{1234567, "12345.67"},
	}
	for _, c := range cases {
		got := kopToRubString(c.kop)
		if got != c.want {
			t.Errorf("kopToRubString(%d)=%q, want %q", c.kop, got, c.want)
		}
	}
}

// TestRubStringToKop — обратная конвертация, ключевой кейс что webhook
// прислал "500.50" а мы хотим 50050 копеек.
func TestRubStringToKop(t *testing.T) {
	cases := []struct {
		s    string
		want int64
		fail bool
	}{
		{"0", 0, false},
		{"0.00", 0, false},
		{"500", 50000, false},
		{"500.00", 50000, false},
		{"500.50", 50050, false},
		{"500.5", 50050, false},
		{"500.01", 50001, false},
		{"abc", 0, true},
		{"-5.00", 0, true},
		{"5.xy", 0, true},
	}
	for _, c := range cases {
		got, err := rubStringToKop(c.s)
		if c.fail {
			if err == nil {
				t.Errorf("rubStringToKop(%q)=%d, want error", c.s, got)
			}
		} else {
			if err != nil {
				t.Errorf("rubStringToKop(%q): %v", c.s, err)
				continue
			}
			if got != c.want {
				t.Errorf("rubStringToKop(%q)=%d, want %d", c.s, got, c.want)
			}
		}
	}
}

// TestNewGateway_Factory проверяет factory: mock и yookassa создаются,
// неизвестный провайдер → ErrUnknownProvider.
func TestNewGateway_Factory(t *testing.T) {
	t.Run("mock", func(t *testing.T) {
		gw, err := NewGateway("mock", FactoryConfig{MockBaseURL: "http://x", MockSecret: "s"})
		if err != nil {
			t.Fatal(err)
		}
		if gw.Name() != "mock" {
			t.Errorf("name=%q", gw.Name())
		}
	})
	t.Run("yookassa OK", func(t *testing.T) {
		gw, err := NewGateway("yookassa", FactoryConfig{
			YooKassaShopID:    "shop",
			YooKassaSecretKey: "secret",
			ReturnURL:         "https://x",
		})
		if err != nil {
			t.Fatal(err)
		}
		if gw.Name() != "yookassa" {
			t.Errorf("name=%q", gw.Name())
		}
	})
	t.Run("yookassa missing creds", func(t *testing.T) {
		_, err := NewGateway("yookassa", FactoryConfig{})
		if err == nil {
			t.Errorf("expected error on missing creds")
		}
	})
	t.Run("unknown", func(t *testing.T) {
		_, err := NewGateway("unknown", FactoryConfig{})
		if !errors.Is(err, ErrUnknownProvider) {
			t.Errorf("err=%v, want ErrUnknownProvider", err)
		}
	})
}
