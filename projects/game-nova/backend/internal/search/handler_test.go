package search_test

// План 72.1.16: handler-тесты на auth-границы и валидацию длины запроса.
// Полные round-trip тесты (LATERAL home_planet, banned, ORDER BY)
// требуют TEST_DATABASE_URL.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth/authtest"
	"oxsar/game-nova/internal/search"
)

func newHandler() *search.Handler {
	return search.NewHandler((*pgxpool.Pool)(nil))
}

func TestSearch_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=test", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestSearch_TooShortReturnsEmpty(t *testing.T) {
	t.Parallel()
	h := newHandler()
	uid := "00000000-0000-0000-0000-0000000000aa"
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=a", nil)
	req = req.WithContext(authtest.WithUserID(req.Context(), uid))
	rr := httptest.NewRecorder()
	h.Search(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	// Должен вернуть пустой response без обращения к БД (значит и без panic
	// при nil pool).
	if !contains(body, `"players":[]`) || !contains(body, `"alliances":[]`) || !contains(body, `"planets":[]`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
