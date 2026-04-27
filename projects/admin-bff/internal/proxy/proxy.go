// Package proxy — reverse proxy на upstream-сервисы (identity, billing, game-nova).
//
// Роутинг по префиксу URL path:
//
//	/api/admin/billing/*   → BillingURL
//	/api/admin/game/*      → GameNovaURL  (или /api/admin/* без billing-префикса)
//	/api/admin/*           → IdentityURL  (default — RBAC, audit, users)
//
// Каждый запрос обогащается Authorization: Bearer <access_token> из сессии.
// Hop-by-hop headers (Connection, Keep-Alive, ...) удаляются согласно RFC 7230.
package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"oxsar/admin-bff/internal/handler"
)

type Upstream struct {
	Name   string
	Prefix string // path-prefix для матчинга (например "/api/admin/billing")
	URL    *url.URL
	proxy  *httputil.ReverseProxy
}

// NewUpstream — собирает ReverseProxy для одного backend-сервиса.
func NewUpstream(name, prefix, target string) (*Upstream, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(u)
			// SetXForwarded пересоздаёт X-Forwarded-{For,Host,Proto} на основе
			// In.RemoteAddr/Host/URL.Scheme — клиентские X-Forwarded-* отбрасываются.
			r.SetXForwarded()
			// Backend-сервисы за admin-bff всегда на TLS со стороны клиента,
			// даже если admin-bff локально слушает HTTP.
			r.Out.Header.Set("X-Forwarded-Proto", "https")
			// Чистим admin-bff-specific заголовки — backend их не ждёт.
			r.Out.Header.Del("Cookie")
			r.Out.Header.Del("X-CSRF-Token")
			// Authorization: Rewrite по умолчанию не пробрасывает его из In в Out.
			// Handler() кладёт server-side токен в r.In.Header — копируем явно.
			if auth := r.In.Header.Get("Authorization"); auth != "" {
				r.Out.Header.Set("Authorization", auth)
			} else {
				r.Out.Header.Del("Authorization")
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.ErrorContext(r.Context(), "upstream error",
				slog.String("upstream", name),
				slog.String("err", err.Error()))
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
		},
	}
	return &Upstream{Name: name, Prefix: prefix, URL: u, proxy: rp}, nil
}

// Handler — http.Handler, который проксирует запрос на upstream с
// инжектированным Authorization-header из сессии. Должен быть обёрнут
// в SessionLookup middleware.
func (u *Upstream) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := handler.SessionFromContext(r.Context())
		if !ok {
			http.Error(w, "no session", http.StatusUnauthorized)
			return
		}
		r.Header.Set("Authorization", "Bearer "+sess.AccessToken)
		u.proxy.ServeHTTP(w, r)
	})
}

// MatchPrefix — true если request path начинается с u.Prefix.
func (u *Upstream) MatchPrefix(path string) bool {
	return strings.HasPrefix(path, u.Prefix)
}
