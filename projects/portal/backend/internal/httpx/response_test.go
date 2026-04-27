package httpx

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth и oxsar/portal. При любом изменении синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/httpx/response_test.go
//   - projects/auth/backend/internal/httpx/response_test.go
//   - projects/portal/backend/internal/httpx/response_test.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorError(t *testing.T) {
	t.Parallel()
	e := &Error{Status: 400, Code: "bad_request", Message: "bad request"}
	if got := e.Error(); got != "bad_request: bad request" {
		t.Errorf("Error() = %q, want %q", got, "bad_request: bad request")
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()
	wrapped := Wrap(ErrNotFound, "planet not found")
	if wrapped.Status != http.StatusNotFound {
		t.Errorf("Wrap status = %d, want %d", wrapped.Status, http.StatusNotFound)
	}
	if wrapped.Code != "not_found" {
		t.Errorf("Wrap code = %q, want not_found", wrapped.Code)
	}
	if wrapped.Message != "planet not found" {
		t.Errorf("Wrap message = %q, want %q", wrapped.Message, "planet not found")
	}
}

func TestWriteJSON_SetsContentType(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	WriteJSON(rr, req, http.StatusOK, map[string]string{"key": "value"})
	if ct := rr.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestWriteJSON_NilBodyNoContent(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	WriteJSON(rr, req, http.StatusNoContent, nil)
	if rr.Body.Len() != 0 {
		t.Errorf("expected empty body for nil value, got %q", rr.Body.String())
	}
}

func TestWriteError_KnownError(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	WriteError(rr, req, ErrBadRequest)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
	var body map[string]map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"]["code"] != "bad_request" {
		t.Errorf("code = %q, want bad_request", body["error"]["code"])
	}
}

func TestWriteError_UnknownError(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	WriteError(rr, req, ErrInternal)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestSentinelErrors_UniqueStatus(t *testing.T) {
	t.Parallel()
	sentinels := []*Error{ErrBadRequest, ErrUnauthorized, ErrForbidden, ErrNotFound, ErrConflict, ErrInternal}
	for _, e := range sentinels {
		if e.Status == 0 {
			t.Errorf("sentinel %q has zero status", e.Code)
		}
		if e.Code == "" {
			t.Errorf("sentinel has empty code")
		}
	}
}
