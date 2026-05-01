package profession_test

// План 72.1.15: handler-тесты на auth-границы и валидацию body.
// Полные round-trip тесты (umode-блок, same-profession no-op,
// AutoMsg отправлен) требуют TEST_DATABASE_URL и FK на users —
// см. integration_test.go-паттерн в internal/origin/alien.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"oxsar/game-nova/internal/auth/authtest"
	"oxsar/game-nova/internal/profession"
)

func newHandler() *profession.Handler {
	// nil service — допустимо для тестов 401, поскольку Change/Get
	// проверяют auth до обращения к сервису.
	return profession.NewHandler(nil)
}

func TestChange_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/professions/me",
		strings.NewReader(`{"profession":"miner"}`))
	rr := httptest.NewRecorder()
	h.Change(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", rr.Code, rr.Body.String())
	}
}

func TestGet_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/professions/me", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestChange_EmptyBodyRejected(t *testing.T) {
	t.Parallel()
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/professions/me",
		strings.NewReader(`{}`))
	req = req.WithContext(authtest.WithUserID(req.Context(),
		"00000000-0000-0000-0000-0000000000aa"))
	rr := httptest.NewRecorder()
	h.Change(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}
