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
	rp := httputil.NewSingleHostReverseProxy(u)
	defaultDirector := rp.Director
	rp.Director = func(req *http.Request) {
		defaultDirector(req)
		// Удаляем admin-bff-specific headers и cookies — backend их не ждёт.
		req.Header.Del("Cookie")
		req.Header.Del("X-CSRF-Token")
		// Добавляем X-Forwarded-* для backend-логирования.
		req.Header.Set("X-Forwarded-Proto", "https")
	}
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.ErrorContext(r.Context(), "upstream error",
			slog.String("upstream", name),
			slog.String("err", err.Error()))
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
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
