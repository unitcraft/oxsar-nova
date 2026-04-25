package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_Ok(t *testing.T) {
	s := NewState("backend", "test-1.0")
	rec := httptest.NewRecorder()
	s.HealthHandler()(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !contains(body, `"status":"ok"`) {
		t.Errorf("body missing status=ok: %s", body)
	}
	if !contains(body, `"component":"backend"`) {
		t.Errorf("body missing component: %s", body)
	}
}

func TestHealthHandler_Draining(t *testing.T) {
	s := NewState("backend", "test-1.0")
	s.SetDraining()
	rec := httptest.NewRecorder()
	s.HealthHandler()(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), `"status":"draining"`) {
		t.Errorf("body: %s", rec.Body.String())
	}
}

type fakePing struct{ err error }

func (f fakePing) Ping(ctx context.Context) error { return f.err }

func TestReadyHandler_NotReady(t *testing.T) {
	s := NewState("backend", "")
	rec := httptest.NewRecorder()
	s.ReadyHandler(fakePing{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ready", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503 (not ready), got %d", rec.Code)
	}
	if !contains(rec.Body.String(), `"status":"starting"`) {
		t.Errorf("body: %s", rec.Body.String())
	}
}

func TestReadyHandler_Ok(t *testing.T) {
	s := NewState("backend", "")
	s.SetReady()
	rec := httptest.NewRecorder()
	s.ReadyHandler(fakePing{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ready", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), `"status":"ready"`) {
		t.Errorf("body: %s", rec.Body.String())
	}
}

func TestReadyHandler_DBDown(t *testing.T) {
	s := NewState("backend", "")
	s.SetReady()
	rec := httptest.NewRecorder()
	s.ReadyHandler(fakePing{err: errors.New("db connection refused")}).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ready", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503 (db down), got %d", rec.Code)
	}
	if !contains(rec.Body.String(), `"status":"db_unhealthy"`) {
		t.Errorf("body: %s", rec.Body.String())
	}
}

func TestReadyHandler_DrainingTakesPrecedence(t *testing.T) {
	s := NewState("backend", "")
	s.SetReady()
	s.SetDraining()
	rec := httptest.NewRecorder()
	// Даже при здоровой БД, draining должен ответить 503.
	s.ReadyHandler(fakePing{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ready", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503 (draining), got %d", rec.Code)
	}
	if !contains(rec.Body.String(), `"status":"draining"`) {
		t.Errorf("body: %s", rec.Body.String())
	}
}

func TestState_Flags(t *testing.T) {
	s := NewState("backend", "")
	if s.IsReady() {
		t.Error("ready should be false initially")
	}
	if s.IsDraining() {
		t.Error("draining should be false initially")
	}
	s.SetReady()
	if !s.IsReady() {
		t.Error("ready should be true after SetReady")
	}
	s.SetDraining()
	if !s.IsDraining() {
		t.Error("draining should be true after SetDraining")
	}
	// Идемпотентность
	s.SetDraining()
	if !s.IsDraining() {
		t.Error("SetDraining should be idempotent")
	}
}

// contains — простой substring-чек без strings импорта.
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
