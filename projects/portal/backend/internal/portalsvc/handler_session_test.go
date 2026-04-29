package portalsvc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"oxsar/portal/internal/universe"
)

// План 72.2: handler-тесты для CreateUniverseSession.

func newTestHandler(t *testing.T, identityURL string) *Handler {
	t.Helper()
	reg, err := universe.NewRegistryFromSlice([]universe.Universe{
		{
			ID: "uni01", Name: "Nova", Subdomain: "uni01",
			DevURL: "http://localhost:5173", Status: "active",
		},
		{
			ID: "uni99", Name: "Maintenance", Subdomain: "uni99",
			DevURL: "http://localhost:5199", Status: "maintenance",
		},
	})
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	return &Handler{
		svc:      nil, // не нужен в этом handler
		reg:      reg,
		credits:  NewBillingClient(""),
		identity: NewIdentityClient(identityURL),
	}
}

func authedRequest(method, path, body string) *http.Request {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer test-token")
	ctx := context.WithValue(r.Context(), ctxUserID, "user-1")
	return r.WithContext(ctx)
}

func unauthedRequest(method, path string) *http.Request {
	r := httptest.NewRequest(method, path, nil)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func mountSessionRoute(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/universes/{id}/session", h.CreateUniverseSession)
	return r
}

func TestCreateUniverseSession_Unauthorized(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, "http://identity.invalid")
	req := unauthedRequest(http.MethodPost, "/api/universes/uni01/session")
	rec := httptest.NewRecorder()
	mountSessionRoute(h).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestCreateUniverseSession_NotFound(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, "http://identity.invalid")
	req := authedRequest(http.MethodPost, "/api/universes/uni404/session", "{}")
	rec := httptest.NewRecorder()
	mountSessionRoute(h).ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}
}

func TestCreateUniverseSession_NotActive(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, "http://identity.invalid")
	req := authedRequest(http.MethodPost, "/api/universes/uni99/session", "{}")
	rec := httptest.NewRecorder()
	mountSessionRoute(h).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400 (universe not active)", rec.Code)
	}
}

func TestCreateUniverseSession_IdentityUnavailable(t *testing.T) {
	t.Parallel()
	// Identity URL валидный синтаксически, но никто не слушает.
	h := newTestHandler(t, "http://127.0.0.1:1")
	// Делаем httpClient timeout коротким, чтобы тест не висел.
	h.identity.httpClient.Timeout = 200 * time.Millisecond
	req := authedRequest(http.MethodPost, "/api/universes/uni01/session", "{}")
	rec := httptest.NewRecorder()
	mountSessionRoute(h).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", rec.Code)
	}
}

func TestCreateUniverseSession_Success(t *testing.T) {
	t.Parallel()
	// Mock identity-service.
	mockIdentity := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/universe-token" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var in struct {
			UniverseID string `json:"universe_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.UniverseID != "uni01" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"handoff_token":"abc-handoff-XYZ"}`))
	}))
	defer mockIdentity.Close()

	h := newTestHandler(t, mockIdentity.URL)
	req := authedRequest(http.MethodPost, "/api/universes/uni01/session", "{}")
	rec := httptest.NewRecorder()
	mountSessionRoute(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		RedirectURL  string `json:"redirect_url"`
		UniverseID   string `json:"universe_id"`
		UniverseName string `json:"universe_name"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	wantURL := "http://localhost:5173/auth/handoff?code=abc-handoff-XYZ"
	if resp.RedirectURL != wantURL {
		t.Errorf("redirect_url = %q, want %q", resp.RedirectURL, wantURL)
	}
	if resp.UniverseID != "uni01" {
		t.Errorf("universe_id = %q, want uni01", resp.UniverseID)
	}
	if resp.ExpiresIn != 30 {
		t.Errorf("expires_in = %d, want 30", resp.ExpiresIn)
	}
}
