// Command yookassa-mock — fake YooKassa API сервер для dev/staging.
//
// Имитирует:
//   - POST /v3/payments               → создаёт payment, возвращает confirmation_url
//   - GET  /v3/payments/{id}          → отдаёт payment (для re-fetch верификации)
//   - GET  /checkout/{id}             → HTML «страница оплаты» с кнопками
//                                       «Оплатить» / «Отменить» (заменяет
//                                       реальный YooKassa-checkout)
//   - POST /checkout/{id}/pay         → меняет status=succeeded, шлёт webhook
//                                       на billing /billing/webhooks/yookassa
//   - GET  /healthz                   → 200
//
// Хранение payment'ов — in-memory (map). Перезапуск контейнера обнуляет
// состояние; для dev/staging этого достаточно.
//
// Идемпотентность POST /v3/payments — по header Idempotence-Key (как у
// настоящей YooKassa).
//
// Запуск:
//
//	docker run -p 9101:9101 \
//	  -e YOOKASSA_MOCK_ADDR=:9101 \
//	  -e BILLING_WEBHOOK_URL=http://billing-service:9100/billing/webhooks/yookassa \
//	  yookassa-mock
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type confirmation struct {
	Type            string `json:"type"`
	ReturnURL       string `json:"return_url,omitempty"`
	ConfirmationURL string `json:"confirmation_url,omitempty"`
}

type payment struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"` // pending | succeeded | canceled
	Paid         bool              `json:"paid"`
	Amount       amount            `json:"amount"`
	Confirmation *confirmation     `json:"confirmation"`
	Metadata     map[string]string `json:"metadata"`
	ReturnURL    string            `json:"-"`
}

type createReq struct {
	Amount       amount            `json:"amount"`
	Capture      bool              `json:"capture"`
	Confirmation confirmation      `json:"confirmation"`
	Description  string            `json:"description"`
	Metadata     map[string]string `json:"metadata"`
}

type store struct {
	mu       sync.Mutex
	payments map[string]*payment // id → payment
	idemKeys map[string]string   // idem-key → payment.id
	counter  int
}

func newStore() *store {
	return &store{
		payments: map[string]*payment{},
		idemKeys: map[string]string{},
	}
}

func (s *store) create(idem string, req createReq, baseURL string) *payment {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.idemKeys[idem]; ok {
		return s.payments[id]
	}
	s.counter++
	id := fmt.Sprintf("ym-%06d", s.counter)
	p := &payment{
		ID:     id,
		Status: "pending",
		Paid:   false,
		Amount: req.Amount,
		Confirmation: &confirmation{
			Type:            "redirect",
			ConfirmationURL: baseURL + "/checkout/" + id,
		},
		Metadata:  req.Metadata,
		ReturnURL: req.Confirmation.ReturnURL,
	}
	s.payments[id] = p
	if idem != "" {
		s.idemKeys[idem] = id
	}
	return p
}

func (s *store) get(id string) *payment {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.payments[id]
}

func (s *store) markPaid(id string) *payment {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.payments[id]
	if p == nil {
		return nil
	}
	p.Status = "succeeded"
	p.Paid = true
	return p
}

func (s *store) markCanceled(id string) *payment {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.payments[id]
	if p == nil {
		return nil
	}
	p.Status = "canceled"
	p.Paid = false
	return p
}

func main() {
	if err := run(); err != nil {
		slog.Error("yookassa-mock exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := envStr("YOOKASSA_MOCK_ADDR", ":9101")
	publicURL := envStr("YOOKASSA_MOCK_PUBLIC_URL", "http://localhost:9101")
	webhookURL := envStr("BILLING_WEBHOOK_URL", "http://billing-service:9100/billing/webhooks/yookassa")

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)
	log.InfoContext(ctx, "yookassa-mock starting",
		slog.String("addr", addr),
		slog.String("public_url", publicURL),
		slog.String("webhook_url", webhookURL))

	st := newStore()
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// --- API: POST /v3/payments ---
	mux.HandleFunc("/v3/payments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreatePayment(w, r, st, publicURL)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// --- API: GET /v3/payments/{id} ---
	// и одновременно чек-флоу /checkout/{id} (отдаёт HTML).
	// Эндпоинты различаются по префиксу: /v3/payments/<id> vs /checkout/<id>.
	mux.HandleFunc("/v3/payments/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/v3/payments/"):]
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		p := st.get(id)
		if p == nil {
			http.Error(w, `{"type":"error","code":"not_found"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, p)
	})

	// --- Checkout HTML ---
	mux.HandleFunc("/checkout/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/checkout/"):]
		// /checkout/<id> или /checkout/<id>/pay или /checkout/<id>/cancel
		var action string
		if i := indexByte(id, '/'); i >= 0 {
			action = id[i+1:]
			id = id[:i]
		}
		p := st.get(id)
		if p == nil {
			http.Error(w, "payment not found", http.StatusNotFound)
			return
		}
		switch action {
		case "":
			renderCheckout(w, p)
		case "pay":
			handleCheckoutPay(w, r, st, id, webhookURL, log)
		case "cancel":
			handleCheckoutCancel(w, r, st, id)
		default:
			http.Error(w, "unknown action", http.StatusBadRequest)
		}
	})

	// --- Admin: список платежей (для dev-debug) ---
	mux.HandleFunc("/admin/payments", func(w http.ResponseWriter, _ *http.Request) {
		st.mu.Lock()
		defer st.mu.Unlock()
		writeJSON(w, http.StatusOK, st.payments)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "listening", slog.String("addr", addr))
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		log.InfoContext(ctx, "shutdown requested")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func handleCreatePayment(w http.ResponseWriter, r *http.Request, st *store, publicURL string) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	var req createReq
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, `{"type":"error","code":"bad_json"}`, http.StatusBadRequest)
		return
	}
	idem := r.Header.Get("Idempotence-Key")
	p := st.create(idem, req, publicURL)
	writeJSON(w, http.StatusOK, p)
}

func handleCheckoutPay(w http.ResponseWriter, r *http.Request, st *store, id, webhookURL string, log *slog.Logger) {
	p := st.markPaid(id)
	if p == nil {
		http.Error(w, "payment not found", http.StatusNotFound)
		return
	}
	// Шлём webhook в billing-service (fire-and-forget).
	go sendWebhook(webhookURL, p, "payment.succeeded", log)
	// Редирект на ReturnURL (с ?yookassa_status=success для удобства).
	if p.ReturnURL != "" {
		ret := p.ReturnURL
		if !contains(ret, "?") {
			ret += "?yookassa_status=success"
		} else {
			ret += "&yookassa_status=success"
		}
		http.Redirect(w, r, ret, http.StatusFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("paid (no return_url)"))
}

func handleCheckoutCancel(w http.ResponseWriter, r *http.Request, st *store, id string) {
	p := st.markCanceled(id)
	if p == nil {
		http.Error(w, "payment not found", http.StatusNotFound)
		return
	}
	if p.ReturnURL != "" {
		http.Redirect(w, r, p.ReturnURL+"?yookassa_status=canceled", http.StatusFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("canceled"))
}

func sendWebhook(webhookURL string, p *payment, event string, log *slog.Logger) {
	body, _ := json.Marshal(map[string]any{
		"type":   "notification",
		"event":  event,
		"object": p,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		log.Error("webhook build req", slog.String("err", err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("webhook send", slog.String("err", err.Error()))
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<14))
	log.Info("webhook delivered",
		slog.String("payment_id", p.ID),
		slog.Int("status", resp.StatusCode),
		slog.String("body", string(respBody)))
}

var checkoutTpl = template.Must(template.New("checkout").Parse(`<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <title>YooKassa Mock — оплата</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 480px; margin: 60px auto; padding: 20px; background: #f6f7f9; }
    .card { background: white; padding: 32px; border-radius: 12px; box-shadow: 0 4px 24px rgba(0,0,0,0.06); }
    h1 { margin: 0 0 8px; color: #1a73e8; font-size: 22px; }
    .sub { color: #666; font-size: 13px; margin-bottom: 24px; }
    .row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #eee; font-size: 14px; }
    .row:last-child { border-bottom: 0; font-weight: 600; font-size: 16px; }
    .actions { margin-top: 24px; display: flex; gap: 10px; }
    button { flex: 1; padding: 12px; border: 0; border-radius: 8px; font-size: 15px; font-weight: 500; cursor: pointer; }
    .pay { background: #1a73e8; color: white; }
    .cancel { background: #f0f0f0; color: #555; }
    .badge { display: inline-block; padding: 2px 8px; background: #fff3cd; color: #b8860b; border-radius: 4px; font-size: 11px; margin-left: 8px; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Оплата заказа <span class="badge">MOCK</span></h1>
    <div class="sub">Тестовый платёжный шлюз. Имитирует страницу YooKassa.</div>
    <div class="row"><span>Платёж ID</span><span><code>{{.ID}}</code></span></div>
    <div class="row"><span>Заказ</span><span><code>{{.OrderID}}</code></span></div>
    <div class="row"><span>Статус</span><span>{{.Status}}</span></div>
    <div class="row"><span>Сумма</span><span>{{.Amount}} {{.Currency}}</span></div>
    <form method="post" class="actions">
      <button type="submit" class="cancel" formaction="/checkout/{{.ID}}/cancel">Отменить</button>
      <button type="submit" class="pay" formaction="/checkout/{{.ID}}/pay">Оплатить</button>
    </form>
  </div>
</body>
</html>`))

func renderCheckout(w http.ResponseWriter, p *payment) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = checkoutTpl.Execute(w, map[string]string{
		"ID":       p.ID,
		"OrderID":  p.Metadata["order_id"],
		"Status":   p.Status,
		"Amount":   p.Amount.Value,
		"Currency": p.Amount.Currency,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
