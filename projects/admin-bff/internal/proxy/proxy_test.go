package proxy_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"oxsar/admin-bff/internal/handler"
	"oxsar/admin-bff/internal/proxy"
	"oxsar/admin-bff/internal/session"
)

// withSession — оборачивает Handler так, чтобы тестовый запрос нёс сессию
// в контексте (без поднятия Redis/SessionLookup).
func withSession(h http.Handler, sess *session.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := handler.ContextWithSession(r.Context(), sess)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// sendThroughProxy поднимает backend-mock + admin-bff Upstream + http-сервер
// с инжектированной сессией, шлёт один запрос и возвращает ответ + заголовки,
// которые увидел backend.
func sendThroughProxy(t *testing.T, sess *session.Session, mutate func(r *http.Request)) (*http.Response, http.Header) {
	t.Helper()
	var captured http.Header
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(backend.Close)
	up, err := proxy.NewUpstream("test", "/api/admin/", backend.URL)
	if err != nil {
		t.Fatalf("NewUpstream: %v", err)
	}
	srv := httptest.NewServer(withSession(up.Handler(), sess))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/admin/users", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if mutate != nil {
		mutate(req)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp, captured
}

func testSession() *session.Session {
	return &session.Session{
		ID:          "sess-1",
		AccessToken: "real-server-side-token",
		CSRFToken:   "csrf-1",
	}
}

// 1) Клиентский X-Forwarded-For не пробрасывается; backend видит свой,
// собранный SetXForwarded на основе RemoteAddr.
func TestRewrite_StripsClientXForwardedFor(t *testing.T) {
	_, backendHeaders := sendThroughProxy(t, testSession(), func(r *http.Request) {
		r.Header.Set("X-Forwarded-For", "8.8.8.8")
	})
	got := backendHeaders.Get("X-Forwarded-For")
	if got == "" {
		t.Fatalf("X-Forwarded-For не установлен на backend")
	}
	if strings.Contains(got, "8.8.8.8") {
		t.Fatalf("backend увидел подделанный X-Forwarded-For=%q (содержит клиентский 8.8.8.8)", got)
	}
}

// 2) Клиентский Authorization игнорируется; backend получает токен сессии.
func TestRewrite_AuthorizationFromSessionWinsOverClient(t *testing.T) {
	_, backendHeaders := sendThroughProxy(t, testSession(), func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer evil-client-token")
	})
	want := "Bearer real-server-side-token"
	if got := backendHeaders.Get("Authorization"); got != want {
		t.Fatalf("Authorization: want %q, got %q", want, got)
	}
}

// 3) Cookie вырезается перед отправкой backend'у.
func TestRewrite_StripsCookie(t *testing.T) {
	_, backendHeaders := sendThroughProxy(t, testSession(), func(r *http.Request) {
		r.Header.Set("Cookie", "admin_session=abc; admin_csrf=xyz")
	})
	if got := backendHeaders.Get("Cookie"); got != "" {
		t.Fatalf("Cookie должен быть удалён, got %q", got)
	}
}

// 4) X-CSRF-Token вырезается.
func TestRewrite_StripsCSRFToken(t *testing.T) {
	_, backendHeaders := sendThroughProxy(t, testSession(), func(r *http.Request) {
		r.Header.Set("X-CSRF-Token", "csrf-1")
	})
	if got := backendHeaders.Get("X-CSRF-Token"); got != "" {
		t.Fatalf("X-CSRF-Token должен быть удалён, got %q", got)
	}
}

// 5) Smoke: запрос с правильной сессией доходит, backend видит Authorization
// и X-Forwarded-Proto=https.
func TestRewrite_SmokeWithSession(t *testing.T) {
	resp, backendHeaders := sendThroughProxy(t, testSession(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
	if got := backendHeaders.Get("Authorization"); got != "Bearer real-server-side-token" {
		t.Fatalf("Authorization: want server-side token, got %q", got)
	}
	if got := backendHeaders.Get("X-Forwarded-Proto"); got != "https" {
		t.Fatalf("X-Forwarded-Proto: want https, got %q", got)
	}
}

// 6) Без сессии Handler возвращает 401 и не зовёт backend.
func TestHandler_NoSessionReturns401(t *testing.T) {
	backendCalled := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		backendCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)
	up, err := proxy.NewUpstream("test", "/api/admin/", backend.URL)
	if err != nil {
		t.Fatalf("NewUpstream: %v", err)
	}
	srv := httptest.NewServer(up.Handler())
	t.Cleanup(srv.Close)
	resp, err := srv.Client().Get(srv.URL + "/api/admin/users")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: want 401, got %d", resp.StatusCode)
	}
	if backendCalled {
		t.Fatalf("backend не должен вызываться без сессии")
	}
}
