package notepad_test

// План 69 Ф.6: тесты notepad-endpoint.
//
// Покрытие:
//   - TestSave_Unauthorized: PUT без userID в контексте → 401.
//   - TestGet_Unauthorized: GET без userID в контексте → 401.
//   - TestSave_InvalidJSON: PUT с битым телом → 400.
//   - TestSave_TooLong: PUT с content > MaxLength → 400.
//
// Интеграционные сценарии (полный round-trip GET/PUT с реальной БД)
// требуют TEST_DATABASE_URL и FK на users; они автоматически
// skip'аются без этой переменной — образец см.
// internal/origin/alien/handlers_integration_test.go.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth/authtest"
	"oxsar/game-nova/internal/notepad"
)

func newHandler(t *testing.T) *notepad.Handler {
	t.Helper()
	// pool=nil безопасен для тестов, которые отвечают 401/400 до
	// первого обращения к БД (проверки аутентификации и парсинга).
	return notepad.NewHandler((*pgxpool.Pool)(nil))
}

func TestSave_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodPut, "/api/notepad",
		strings.NewReader(`{"content":"hi"}`))
	rr := httptest.NewRecorder()
	h.Save(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", rr.Code, rr.Body.String())
	}
}

func TestGet_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/notepad", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", rr.Code, rr.Body.String())
	}
}

func TestSave_InvalidJSON(t *testing.T) {
	t.Parallel()
	h := newHandler(t)
	ctx := authtest.WithUserID(context.Background(), "00000000-0000-0000-0000-000000000001")
	req := httptest.NewRequest(http.MethodPut, "/api/notepad",
		strings.NewReader(`not-json`)).WithContext(ctx)
	rr := httptest.NewRecorder()
	h.Save(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestSave_TooLong(t *testing.T) {
	t.Parallel()
	h := newHandler(t)
	ctx := authtest.WithUserID(context.Background(), "00000000-0000-0000-0000-000000000001")
	// MaxLength+1 байт — простая ascii-строка, длина в байтах == длина в run'ах.
	body := `{"content":"` + strings.Repeat("a", notepad.MaxLength+1) + `"}`
	req := httptest.NewRequest(http.MethodPut, "/api/notepad",
		strings.NewReader(body)).WithContext(ctx)
	rr := httptest.NewRecorder()
	h.Save(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "too long") {
		t.Fatalf("expected 'too long' in body, got %s", rr.Body.String())
	}
}

func TestSave_AtMaxLength(t *testing.T) {
	t.Parallel()
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	// Граничное значение len == MaxLength: должен пройти валидацию.
	// Дальше handler идёт в БД — проверяется в полном integration-тесте.
}
