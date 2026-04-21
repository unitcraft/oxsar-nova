package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestIssuer() *JWTIssuer {
	return NewJWTIssuer("middleware-test-secret", time.Minute, time.Hour)
}

func okHandler(t *testing.T, wantUID string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserID(r.Context())
		if !ok {
			t.Errorf("userID not in context")
		}
		if uid != wantUID {
			t.Errorf("userID = %q, want %q", uid, wantUID)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_BearerHeader(t *testing.T) {
	t.Parallel()
	j := newTestIssuer()
	toks, _ := j.Issue("u-1")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+toks.Access)
	rr := httptest.NewRecorder()

	Middleware(j)(okHandler(t, "u-1")).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_QueryParamToken(t *testing.T) {
	t.Parallel()
	j := newTestIssuer()
	toks, _ := j.Issue("u-ws")

	req := httptest.NewRequest(http.MethodGet, "/?token="+toks.Access, nil)
	rr := httptest.NewRecorder()

	Middleware(j)(okHandler(t, "u-ws")).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (WS query param path)", rr.Code)
	}
}

func TestMiddleware_NoToken_Returns401(t *testing.T) {
	t.Parallel()
	j := newTestIssuer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true })
	Middleware(j)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if reached {
		t.Fatal("next handler must not be called without a token")
	}
}

func TestMiddleware_BadToken_Returns401(t *testing.T) {
	t.Parallel()
	j := newTestIssuer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer totally.invalid.token")
	rr := httptest.NewRecorder()

	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true })
	Middleware(j)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if reached {
		t.Fatal("next handler must not be called with invalid token")
	}
}

func TestMiddleware_RefreshToken_Rejected(t *testing.T) {
	t.Parallel()
	j := newTestIssuer()
	toks, _ := j.Issue("u-2")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+toks.Refresh)
	rr := httptest.NewRecorder()

	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true })
	Middleware(j)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for refresh token used as access, got %d", rr.Code)
	}
	if reached {
		t.Fatal("next handler must not be called with refresh token")
	}
}

func TestUserID_NoMiddleware(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	uid, ok := UserID(req.Context())
	if ok || uid != "" {
		t.Fatalf("UserID should return empty/false without middleware, got %q %v", uid, ok)
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	t.Parallel()
	j := NewJWTIssuer("sec", -time.Second, time.Hour) // negative TTL → already expired
	toks, err := j.Issue("u-exp")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := j.Parse(toks.Access, "access"); err == nil {
		t.Fatal("expired token must be rejected")
	}
}

func TestJWT_RefreshRoundtrip(t *testing.T) {
	t.Parallel()
	j := NewJWTIssuer("sec", time.Minute, time.Hour)
	toks, _ := j.Issue("u-r")
	uid, err := j.Parse(toks.Refresh, "refresh")
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	if uid != "u-r" {
		t.Fatalf("expected u-r, got %s", uid)
	}
}

func TestJWT_ExpiresFieldSet(t *testing.T) {
	t.Parallel()
	j := NewJWTIssuer("sec", 5*time.Minute, time.Hour)
	before := time.Now().UTC()
	toks, _ := j.Issue("u-exp-field")
	after := time.Now().UTC()
	if toks.Expires.Before(before.Add(4*time.Minute)) || toks.Expires.After(after.Add(6*time.Minute)) {
		t.Fatalf("Expires %v not in expected range", toks.Expires)
	}
}
