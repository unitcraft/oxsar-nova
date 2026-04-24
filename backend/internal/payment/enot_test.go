package payment

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestEnotGateway_BuildPayURL(t *testing.T) {
	gw := NewEnotGateway("shop42", "supersecret")
	got, err := gw.BuildPayURL(context.Background(), "order-abc", "Test pkg", 15050, "https://return.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, enotBaseURL+"?") {
		t.Fatalf("unexpected prefix: %s", got)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("m") != "shop42" {
		t.Errorf("m = %q", q.Get("m"))
	}
	if q.Get("oa") != "150.50" {
		t.Errorf("oa = %q, want 150.50", q.Get("oa"))
	}
	if q.Get("o") != "order-abc" {
		t.Errorf("o = %q", q.Get("o"))
	}
	// sig = MD5("shop42:150.50:supersecret:order-abc")
	wantSig := enotMD5("shop42", "150.50", "supersecret", "order-abc")
	if q.Get("s") != wantSig {
		t.Errorf("s = %q, want %q", q.Get("s"), wantSig)
	}
	if q.Get("cr") != "RUB" {
		t.Errorf("cr = %q", q.Get("cr"))
	}
}

func TestEnotGateway_VerifyWebhook_FormValid(t *testing.T) {
	gw := NewEnotGateway("shop42", "supersecret")
	sig := enotMD5("shop42", "150.50", "supersecret", "order-abc")
	form := url.Values{
		"merchant": {"shop42"},
		"amount":   {"150.50"},
		"order_id": {"order-abc"},
		"sign_2":   {sig},
	}
	r := &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		Body:   http.NoBody,
	}
	r.Form = form
	orderID, amountKop, err := gw.VerifyWebhook(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orderID != "order-abc" {
		t.Errorf("orderID = %q", orderID)
	}
	if amountKop != 15050 {
		t.Errorf("amountKop = %d, want 15050", amountKop)
	}
}

func TestEnotGateway_VerifyWebhook_JSONValid(t *testing.T) {
	gw := NewEnotGateway("shop42", "supersecret")
	sig := enotMD5("shop42", "150.50", "supersecret", "order-abc")
	body := `{"merchant":"shop42","amount":"150.50","order_id":"order-abc","sign_2":"` + sig + `"}`
	r := &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/json"}},
		Body:   http.NoBody,
	}
	r.Body = nopCloser{strings.NewReader(body)}
	orderID, amountKop, err := gw.VerifyWebhook(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orderID != "order-abc" || amountKop != 15050 {
		t.Errorf("orderID=%q amountKop=%d", orderID, amountKop)
	}
}

func TestEnotGateway_VerifyWebhook_WrongSig(t *testing.T) {
	gw := NewEnotGateway("shop42", "supersecret")
	form := url.Values{
		"merchant": {"shop42"},
		"amount":   {"150.50"},
		"order_id": {"order-abc"},
		"sign_2":   {"badsig"},
	}
	r := &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		Body:   http.NoBody,
	}
	r.Form = form
	if _, _, err := gw.VerifyWebhook(r); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEnotGateway_VerifyWebhook_WrongShop(t *testing.T) {
	gw := NewEnotGateway("shop42", "supersecret")
	sig := enotMD5("otherShop", "150.50", "supersecret", "order-abc")
	form := url.Values{
		"merchant": {"otherShop"}, // чужой магазин
		"amount":   {"150.50"},
		"order_id": {"order-abc"},
		"sign_2":   {sig},
	}
	r := &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		Body:   http.NoBody,
	}
	r.Form = form
	if _, _, err := gw.VerifyWebhook(r); err == nil {
		t.Fatal("expected error for wrong merchant, got nil")
	}
}

type nopCloser struct{ *strings.Reader }

func (nopCloser) Close() error { return nil }
