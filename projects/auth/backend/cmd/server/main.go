// Command auth-service — единая аутентификация для всех вселенных oxsar-nova.
// Выпускает RSA-256 JWT, обслуживает OAuth, управляет global credits.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"oxsar/auth/internal/authsvc"
	"oxsar/auth/internal/httpx"
	"oxsar/auth/internal/storage"
	"oxsar/auth/pkg/jwtrs"
	"oxsar/auth/pkg/metrics"
)

const drainDelay = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("auth-service exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := envStr("AUTH_ADDR", ":9000")
	dbURL := mustEnv("AUTH_DB_URL")
	redisURL := envStr("REDIS_URL", "redis://localhost:6379/0")
	keyPath := envStr("RSA_KEY_PATH", "/run/secrets/auth_rsa_key.pem")
	accessTTL := envDur("JWT_ACCESS_TTL", 60*time.Minute)
	refreshTTL := envDur("JWT_REFRESH_TTL", 30*24*time.Hour)
	allowedOrigins := strings.Split(envStr("ALLOWED_ORIGINS",
		"http://localhost:5173,http://localhost:3000"), ",")

	log := newLogger(envStr("LOG_LEVEL", "info"))
	slog.SetDefault(log)
	log.InfoContext(ctx, "starting auth-service", slog.String("addr", addr))

	pool, err := storage.OpenPostgres(ctx, dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	rdb, err := storage.OpenRedis(ctx, redisURL)
	if err != nil {
		log.WarnContext(ctx, "redis unavailable", slog.String("err", err.Error()))
	}

	rsaKey, err := jwtrs.LoadOrGenerateKey(keyPath)
	if err != nil {
		return err
	}
	iss := jwtrs.NewIssuer(rsaKey, accessTTL, refreshTTL)
	ver := jwtrs.NewVerifierFromKey(iss.PublicKey())

	svc := authsvc.New(pool, iss)
	h := authsvc.NewHandler(svc, iss, rdb)

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
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Публичные endpoints
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/.well-known/jwks.json", h.JWKS)

	// Auth endpoints (публичные + rate-limited)
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)

	// Обмен handoff-токена → JWT (вызывается игровым сервером)
	r.Post("/auth/token/exchange", h.TokenExchange)

	// Внутренние endpoints (между сервисами, закрыты firewall-ом)
	r.Post("/auth/credits/spend", h.SpendCredits)
	r.Post("/auth/universes/register", h.RegisterUniverse)

	// Защищённые endpoints (требуют JWT)
	r.Group(func(pr chi.Router) {
		pr.Use(authsvc.Middleware(ver))
		pr.Get("/auth/me", h.Me)
		pr.Get("/auth/credits/balance", h.CreditBalance)
		pr.Get("/auth/credits/history", h.CreditHistory)
		pr.Post("/auth/universe-token", h.UniverseToken)
	})

	// Prometheus metrics
	r.Handle("/metrics", metrics.Register())

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

func envDur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

var _ = envInt // used by future OAuth handlers
