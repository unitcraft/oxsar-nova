package friends_test

// План 72.1.14: тесты friends-handler.
//
// Полные round-trip тесты двустороннего accept-flow требуют
// TEST_DATABASE_URL (FK на users + INSERT/UPDATE в friends);
// здесь проверки auth-границ и валидации параметров без БД.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth/authtest"
	"oxsar/game-nova/internal/friends"
)

func newHandler() *friends.Handler {
	return friends.NewHandler((*pgxpool.Pool)(nil))
}

func TestList_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/friends", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdd_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/friends/x", nil)
	rr := httptest.NewRecorder()
	h.Add(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestAdd_SelfRejected(t *testing.T) {
	t.Parallel()
	h := newHandler()
	uid := "00000000-0000-0000-0000-0000000000aa"
	req := httptest.NewRequest(http.MethodPost, "/api/friends/"+uid, nil)
	// Сначала навешиваем route context (chi.URLParam читает оттуда),
	// затем auth — оба context'а коммулируются на req.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userId", uid)
	ctx := req.Context()
	ctx = authtest.WithUserID(ctx, uid)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.Add(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("self-add status = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
}

func TestAccept_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/friends/x/accept", nil)
	rr := httptest.NewRecorder()
	h.Accept(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestRemove_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newHandler()
	req := httptest.NewRequest(http.MethodDelete, "/api/friends/x", nil)
	rr := httptest.NewRecorder()
	h.Remove(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}
