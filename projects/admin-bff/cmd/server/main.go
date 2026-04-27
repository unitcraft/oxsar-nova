// Command admin-bff — Backend-for-Frontend для admin.oxsar-nova.ru.
//
// Хранит сессии в Redis, проксирует /api/* на backend-сервисы
// (identity, billing, game-nova), добавляя Authorization: Bearer <JWT>.
// Браузер видит только opaque admin_session cookie.
//
// Архитектура: см. docs/plans/53-admin-frontend.md §Auth-flow (BFF).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"oxsar/admin-bff/internal/config"
	"oxsar/admin-bff/internal/handler"
	"oxsar/admin-bff/internal/httpx"
	"oxsar/admin-bff/internal/identityclient"
	"oxsar/admin-bff/internal/proxy"
	"oxsar/admin-bff/internal/session"
)

const drainDelay = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("admin-bff exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logLevel := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(log)
	log.InfoContext(ctx, "starting admin-bff", slog.String("addr", cfg.ListenAddr))

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	defer rdb.Close()

	store := session.NewStore(rdb, cfg.IdleTimeout)
	identity := identityclient.New(cfg.IdentityURL)
	authH := handler.NewAuth(identity, store, cfg.CookieDomain, cfg.CookieSecure, cfg.IdleTimeout)

	billingUp, err := proxy.NewUpstream("billing", "/api/admin/billing", cfg.BillingURL)
	if err != nil {
		return err
	}
	// game-nova admin-namespace: events/planets/fleets/galaxy-events.
	// Префикс `/api/admin/game` зарезервирован под sub-план 53b
	// (миграция game-nova endpoints в namespaced-prefix). Сейчас
	// проксируем конкретные пути, чтобы не пересекаться с identity.
	// План 53 Ф.6 обрабатывает только events.
	gameEventsUp, err := proxy.NewUpstream("game-nova-events", "/api/admin/events", cfg.GameNovaURL)
	if err != nil {
		return err
	}
	identityUp, err := proxy.NewUpstream("identity", "/api/admin", cfg.IdentityURL)
	if err != nil {
		return err
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Public auth endpoints (login не требует сессии).
	r.Post("/auth/login", authH.Login)
	r.Post("/auth/logout", authH.Logout)

	// Protected — требует валидную сессию.
	r.Group(func(pr chi.Router) {
		pr.Use(handler.SessionLookup(store, identity, cfg.RefreshLeadTime))
		pr.Use(handler.CSRF())
		pr.Get("/auth/me", authH.Me)

		// Reverse-proxy на backend-сервисы. Порядок важен — billing/game
		// должны проверяться до identity (более специфичные prefix).
		pr.Mount(billingUp.Prefix, billingUp.Handler())
		pr.Mount(gameEventsUp.Prefix, gameEventsUp.Handler())
		pr.Mount(identityUp.Prefix, identityUp.Handler())
	})

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "http listening", slog.String("addr", cfg.ListenAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.InfoContext(ctx, "shutdown signal", slog.String("reason", ctx.Err().Error()))
	case err := <-errCh:
		return err
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), drainDelay)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.ErrorContext(shutCtx, "http shutdown error", slog.String("err", err.Error()))
	}
	return nil
}
