package payment

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestRobokassaGateway_VerifyWebhook(t *testing.T) {
	gw := NewRobokassaGateway("shop", "pass1", "pass2")

	t.Run("valid signature", func(t *testing.T) {
		// MD5("100.00:42:pass2") = ожидаемая подпись
		sig := robokassaMD5("100.00", "42", "pass2")

		form := url.Values{
			"OutSum":         {"100.00"},
			"InvId":          {"42"},
			"SignatureValue": {sig},
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
		if orderID != "42" {
			t.Errorf("orderID = %q, want %q", orderID, "42")
		}
		if amountKop != 10000 {
			t.Errorf("amountKop = %d, want 10000", amountKop)
		}
	})

	t.Run("wrong signature", func(t *testing.T) {
		form := url.Values{
			"OutSum":         {"100.00"},
			"InvId":          {"42"},
			"SignatureValue": {"badsig"},
		}
		r := &http.Request{
			Method: http.MethodPost,
			Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			Body:   http.NoBody,
		}
		r.Form = form

		_, _, err := gw.VerifyWebhook(r)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		r := &http.Request{
			Method: http.MethodPost,
			Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			Body:   http.NoBody,
		}
		r.Form = url.Values{}

		_, _, err := gw.VerifyWebhook(r)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("case insensitive sig", func(t *testing.T) {
		sig := strings.ToUpper(robokassaMD5("50.00", "7", "pass2"))

		form := url.Values{
			"OutSum":         {"50.00"},
			"InvId":          {"7"},
			"SignatureValue": {sig},
		}
		r := &http.Request{
			Method: http.MethodPost,
			Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			Body:   http.NoBody,
		}
		r.Form = form

		_, _, err := gw.VerifyWebhook(r)
		if err != nil {
			t.Fatalf("uppercase sig should be accepted: %v", err)
		}
	})
}

func TestPackageByKey(t *testing.T) {
	pkg, ok := PackageByKey("medium")
	if !ok {
		t.Fatal("expected to find 'medium' package")
	}
	if pkg.TotalCredits() != 3200 {
		t.Errorf("TotalCredits = %d, want 3200", pkg.TotalCredits())
	}
	if pkg.PriceRub() != 250.0 {
		t.Errorf("PriceRub = %f, want 250.0", pkg.PriceRub())
	}

	_, ok = PackageByKey("nonexistent")
	if ok {
		t.Error("expected not found for unknown key")
	}
}
