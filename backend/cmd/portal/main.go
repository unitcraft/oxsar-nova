// Command portal — HTTP-сервер портала oxsar-nova.ru.
// Обслуживает список вселенных, новости и систему предложений (feedback).
package main

import (
	"context"
	"errors"
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

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/portalsvc"
	"github.com/oxsar/nova/backend/internal/storage"
	"github.com/oxsar/nova/backend/internal/universe"
	"github.com/oxsar/nova/backend/pkg/jwtrs"
	"github.com/oxsar/nova/backend/pkg/metrics"
)

const drainDelay = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("portal exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := envStr("PORTAL_ADDR", ":8090")
	dbURL := mustEnv("PORTAL_DB_URL")
	jwksURL := envStr("AUTH_JWKS_URL", "")
	authServiceURL := envStr("AUTH_SERVICE_URL", "")
	universesPath := envStr("UNIVERSES_CONFIG", "configs/universes.yaml")
	allowedOrigins := strings.Split(envStr("ALLOWED_ORIGINS",
		"http://localhost:5174,http://localhost:3000"), ",")

	log := newLogger(envStr("LOG_LEVEL", "info"))
	slog.SetDefault(log)
	log.InfoContext(ctx, "starting portal", slog.String("addr", addr))

	pool, err := storage.OpenPostgres(ctx, dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	var ver *jwtrs.Verifier
	if jwksURL != "" {
		if v, verErr := auth.LoadVerifier(ctx, jwksURL); verErr != nil {
			log.WarnContext(ctx, "JWKS not loaded — protected endpoints will reject all requests",
				slog.String("err", verErr.Error()))
		} else {
			ver = v
		}
	} else {
		log.WarnContext(ctx, "AUTH_JWKS_URL not set — auth middleware disabled (dev mode)")
	}

	reg, err := universe.NewRegistry(universesPath)
	if err != nil {
		log.WarnContext(ctx, "universes config not loaded", slog.String("err", err.Error()))
		reg, _ = universe.NewRegistryFromSlice(nil)
	}

	svc := portalsvc.New(pool)
	h := portalsvc.NewHandlerWithCredits(svc, reg, authServiceURL)

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
	r.Get("/api/universes", h.ListUniverses)
	r.Get("/api/news", h.ListNews)
	r.Get("/api/news/{id}", h.GetNews)
	r.Get("/api/feedback", h.ListFeedback)
	r.Get("/api/feedback/{id}", h.GetFeedback)
	r.Get("/api/feedback/{id}/comments", h.ListComments)

	// Защищённые endpoints (требуют JWT от auth-service).
	// В dev-режиме без JWKS middleware пропускает все запросы (ver == nil).
	r.Group(func(pr chi.Router) {
		if ver != nil {
			pr.Use(portalsvc.Middleware(ver))
		}
		pr.Post("/api/feedback", h.CreateFeedback)
		pr.Post("/api/feedback/{id}/vote", h.VoteFeedback)
		pr.Post("/api/feedback/{id}/comments", h.AddComment)

		// Admin-only
		pr.Group(func(ar chi.Router) {
			ar.Use(portalsvc.AdminMiddleware)
			ar.Post("/api/news", h.CreateNews)
			ar.Patch("/api/feedback/{id}/status", h.ModerateFeedback)
		})
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
