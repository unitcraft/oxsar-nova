package payment

import (
	"context"
	"testing"
)

func TestMockGateway_BuildPayURL(t *testing.T) {
	gw := NewMockGateway("http://localhost:8080")

	payURL, err := gw.BuildPayURL(context.Background(), "order-42", "Starter", 10000, "http://localhost:5173/")
	if err != nil {
		t.Fatalf("BuildPayURL: %v", err)
	}

	wantPrefix := "http://localhost:8080/api/payment/mock/pay?"
	if len(payURL) < len(wantPrefix) || payURL[:len(wantPrefix)] != wantPrefix {
		t.Errorf("pay URL = %q, want prefix %q", payURL, wantPrefix)
	}
	for _, substr := range []string{"order=order-42", "amount=10000", "result=success", "return=http"} {
		if !containsSubstr(payURL, substr) {
			t.Errorf("pay URL %q missing %q", payURL, substr)
		}
	}
}

func TestMockGateway_EmptyReturnURL(t *testing.T) {
	gw := NewMockGateway("")
	payURL, err := gw.BuildPayURL(context.Background(), "order-7", "Trial", 4900, "")
	if err != nil {
		t.Fatalf("BuildPayURL: %v", err)
	}
	if containsSubstr(payURL, "return=") {
		t.Errorf("pay URL should not include return when empty, got %q", payURL)
	}
}

func TestMockGateway_IsMock(t *testing.T) {
	gw := NewMockGateway("")
	if !gw.IsMock() {
		t.Error("MockGateway.IsMock() = false, want true")
	}
}

func TestServiceIsMock(t *testing.T) {
	t.Run("mock provider", func(t *testing.T) {
		svc := &Service{gateway: NewMockGateway("")}
		if !svc.IsMock() {
			t.Error("Service.IsMock() = false, want true")
		}
	})
	t.Run("robokassa provider", func(t *testing.T) {
		svc := &Service{gateway: NewRobokassaGateway("l", "p1", "p2")}
		if svc.IsMock() {
			t.Error("Service.IsMock() = true, want false")
		}
	})
	t.Run("no gateway", func(t *testing.T) {
		svc := &Service{}
		if svc.IsMock() {
			t.Error("Service.IsMock() = true, want false")
		}
	})
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
