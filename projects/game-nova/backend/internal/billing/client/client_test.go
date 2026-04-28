package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newClientFor(srvURL string) *Client {
	return &Client{
		billingURL: srvURL,
		httpClient: &http.Client{Timeout: 2 * time.Second},
	}
}

func TestSpend_OK(t *testing.T) {
	var gotPath, gotIdem, gotAuth string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotIdem = r.Header.Get("Idempotency-Key")
		gotAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClientFor(srv.URL)
	err := c.Spend(context.Background(), SpendInput{
		UserToken:      "test.jwt",
		Amount:         100,
		Reason:         "teleport_planet",
		RefID:          "planet-42",
		ToAccount:      "system:teleport",
		IdempotencyKey: "user-1:teleport:planet-42",
	})
	if err != nil {
		t.Fatalf("Spend: %v", err)
	}
	if gotPath != "/billing/wallet/spend" {
		t.Errorf("path = %q, want /billing/wallet/spend", gotPath)
	}
	if gotIdem != "user-1:teleport:planet-42" {
		t.Errorf("Idempotency-Key = %q", gotIdem)
	}
	if gotAuth != "Bearer test.jwt" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotBody["amount"].(float64) != 100 {
		t.Errorf("amount = %v", gotBody["amount"])
	}
	if gotBody["to_account"] != "system:teleport" {
		t.Errorf("to_account = %v", gotBody["to_account"])
	}
}

func TestSpend_InsufficientOxsar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
	}))
	defer srv.Close()

	c := newClientFor(srv.URL)
	err := c.Spend(context.Background(), SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k1",
	})
	if !errors.Is(err, ErrInsufficientOxsar) {
		t.Fatalf("got %v, want ErrInsufficientOxsar", err)
	}
}

func TestSpend_IdempotencyConflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()

	c := newClientFor(srv.URL)
	err := c.Spend(context.Background(), SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k2",
	})
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("got %v, want ErrIdempotencyConflict", err)
	}
}

func TestSpend_FrozenWallet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusLocked)
	}))
	defer srv.Close()

	c := newClientFor(srv.URL)
	err := c.Spend(context.Background(), SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k3",
	})
	if !errors.Is(err, ErrFrozenWallet) {
		t.Fatalf("got %v, want ErrFrozenWallet", err)
	}
}

func TestSpend_Generic500NoRetry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newClientFor(srv.URL)
	err := c.Spend(context.Background(), SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k4",
	})
	if err == nil {
		t.Fatalf("want error")
	}
	// 500 — не транзиентный по нашей политике (только 502/503/504),
	// retry не делается.
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1 (no retry on 500)", got)
	}
	if errors.Is(err, ErrBillingUnavailable) {
		t.Errorf("got ErrBillingUnavailable, want generic")
	}
}

func TestSpend_RetryOn503ThenOK(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClientFor(srv.URL)
	err := c.Spend(context.Background(), SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k5",
	})
	if err != nil {
		t.Fatalf("Spend: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2 (1 retry on 503)", got)
	}
}

func TestSpend_RetryOn504ThenStillFails(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer srv.Close()

	c := newClientFor(srv.URL)
	err := c.Spend(context.Background(), SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k6",
	})
	if !errors.Is(err, ErrBillingUnavailable) {
		t.Fatalf("got %v, want ErrBillingUnavailable", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2 (1 retry on 504)", got)
	}
}

func TestSpend_TimeoutAsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{
		billingURL: srv.URL,
		httpClient: &http.Client{Timeout: 50 * time.Millisecond},
	}
	err := c.Spend(context.Background(), SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k7",
	})
	if !errors.Is(err, ErrBillingUnavailable) {
		t.Fatalf("got %v, want ErrBillingUnavailable", err)
	}
}

func TestSpend_NotConfigured(t *testing.T) {
	c := New("")
	err := c.Spend(context.Background(), SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k8",
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("got %v, want ErrNotConfigured", err)
	}
}

func TestRefund_PostsCreditEndpoint(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClientFor(srv.URL)
	err := c.Refund(context.Background(), SpendInput{
		Amount:         100,
		Reason:         "teleport_cancelled",
		RefID:          "planet-42",
		ToAccount:      "system:teleport",
		IdempotencyKey: "user-1:teleport:planet-42:refund",
	})
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if gotPath != "/billing/wallet/credit" {
		t.Errorf("path = %q, want /billing/wallet/credit", gotPath)
	}
	if gotBody["from_account"] != "system:teleport" {
		t.Errorf("from_account = %v", gotBody["from_account"])
	}
	if _, ok := gotBody["to_account"]; ok {
		t.Errorf("Refund body must not contain to_account")
	}
}

func TestSpend_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := newClientFor(srv.URL)
	err := c.Spend(ctx, SpendInput{
		Amount: 100, Reason: "x", ToAccount: "y", IdempotencyKey: "k9",
	})
	if err == nil {
		t.Fatalf("want error from canceled context")
	}
}
