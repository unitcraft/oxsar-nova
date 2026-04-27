// Command billing-service — кошельки, платёжные шлюзы, idempotent spend/credit.
// План 38.
//
// Архитектура: см. docs/plans/38-billing-service.md.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"oxsar/billing/internal/auth"
	"oxsar/billing/internal/httpx"
	"oxsar/billing/internal/storage"
	"oxsar/billing/pkg/metrics"
)

const drainDelay = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("billing exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := envStr("BILLING_ADDR", ":9100")
	dbURL := mustEnv("BILLING_DB_URL")
	redisURL := envStr("REDIS_URL", "redis://localhost:6379/3")
	jwksURL := mustEnv("AUTH_JWKS_URL")
	allowedOrigins := strings.Split(envStr("ALLOWED_ORIGINS",
		"http://localhost:5173,http://localhost:5174"), ",")

	log := newLogger(envStr("LOG_LEVEL", "info"))
	slog.SetDefault(log)
	log.InfoContext(ctx, "starting billing-service", slog.String("addr", addr))

	pool, err := storage.OpenPostgres(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer pool.Close()

	rdb, err := storage.OpenRedis(ctx, redisURL)
	if err != nil {
		return fmt.Errorf("open redis: %w", err)
	}
	_ = rdb // redis используется wallet/idempotency сервисами (Ф.2)

	ver, err := auth.LoadVerifier(ctx, jwksURL)
	if err != nil {
		return fmt.Errorf("load jwks: %w", err)
	}
	_ = ver // используется auth-middleware (Ф.2)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpx.TraceIDMiddleware)
	r.Use(httpx.Logger(log))
	r.Use(httpx.Recoverer(log))
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Idempotency-Key"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/api/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, "db not ready"))
			return
		}
		httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ready"})
	})
	r.Handle("/metrics", metrics.Register())

	// TODO Ф.2: /billing/wallet/{spend,credit,balance,history}
	// TODO Ф.3: /billing/{packages,orders}
	// TODO Ф.4: /billing/webhooks/{provider}

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "listening", slog.String("addr", addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.InfoContext(ctx, "shutdown requested")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	time.Sleep(drainDelay)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func newLogger(level string) *slog.Logger {
	lvl := slog.LevelInfo
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required env var not set", slog.String("key", key))
		os.Exit(1)
	}
	return v
}
