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
	"oxsar/billing/internal/billing"
	"oxsar/billing/internal/httpx"
	"oxsar/billing/internal/payment"
	"oxsar/billing/internal/repo"
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
	// План 51: rename AUTH_* → IDENTITY_*; читаем оба имени (новое приоритетно)
	// для безопасной миграции. Старое имя AUTH_JWKS_URL планируется удалить
	// после прокатки нового конфига во всех окружениях.
	jwksURL := envStr("IDENTITY_JWKS_URL", os.Getenv("AUTH_JWKS_URL"))
	if jwksURL == "" {
		return fmt.Errorf("IDENTITY_JWKS_URL is required (legacy AUTH_JWKS_URL also accepted)")
	}
	allowedOrigins := strings.Split(envStr("ALLOWED_ORIGINS",
		"http://localhost:5173,http://localhost:5174"), ",")
	// Платёжный провайдер. План 42: yookassa добавлен.
	// PAYMENT_PROVIDER либо BILLING_PRIMARY_PROVIDER (план 42 §3) — синонимы;
	// второе предпочтительнее, первое осталось для обратной совместимости.
	provider := envStr("BILLING_PRIMARY_PROVIDER", envStr("PAYMENT_PROVIDER", "mock"))
	returnURL := envStr("PAYMENT_RETURN_URL", "http://localhost:5173/")
	gwCfg := payment.FactoryConfig{
		ReturnURL:                  returnURL,
		MockBaseURL:                envStr("PAYMENT_MOCK_BASE_URL", "http://localhost:9100"),
		MockSecret:                 envStr("PAYMENT_MOCK_SECRET", ""),
		YooKassaShopID:             envStr("YOOKASSA_SHOP_ID", ""),
		YooKassaSecretKey:          envStr("YOOKASSA_SECRET_KEY", ""),
		YooKassaAPIURL:             envStr("YOOKASSA_API_URL", ""),
		YooKassaDisableIPAllowlist: envStr("YOOKASSA_DISABLE_IP_ALLOWLIST", "0") == "1",
	}
	reconcileInterval := envDur("BILLING_RECONCILE_INTERVAL", time.Hour)

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
	_ = rdb // зарезервировано для будущего payment-кэша / rate-limiter

	ver, err := auth.LoadVerifier(ctx, jwksURL)
	if err != nil {
		return fmt.Errorf("load jwks: %w", err)
	}

	db := repo.New(pool)
	svc := billing.New(db)

	// Платёжный шлюз. План 42: factory выбирает по BILLING_PRIMARY_PROVIDER
	// (mock | yookassa). Robokassa/Enot — отложены до тестового аккаунта.
	gw, err := payment.NewGateway(provider, gwCfg)
	if err != nil {
		return fmt.Errorf("init payment gateway: %w", err)
	}
	log.InfoContext(ctx, "payment provider initialized", slog.String("provider", gw.Name()))

	h := billing.NewHandler(svc, gw, returnURL)
	wh := billing.NewWebhookHandler(svc, gw)
	authMW := billing.AuthMiddleware(ver)
	idemMW := billing.NewIdempotencyMiddleware(pool).Handler(billing.UserIDFromCtx)

	// Reconcile cron — план 38 §Reconciliation. Цикл сверки SUM(transactions.delta)
	// с wallet.balance. Расхождение → freeze + Prometheus alert.
	reconciler := billing.NewReconciler(svc, reconcileInterval)
	go reconciler.Run(ctx)
	log.InfoContext(ctx, "reconciler started", slog.Duration("interval", reconcileInterval))

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

	// Публичный каталог пакетов кредитов.
	r.Get("/billing/packages", h.Packages)

	// Защищённые wallet-эндпоинты. AuthMiddleware кладёт user_id в context.
	// IdempotencyMiddleware кэширует ответы по Idempotency-Key (Stripe-style).
	r.Group(func(pr chi.Router) {
		pr.Use(authMW)
		pr.Get("/billing/wallet/balance", h.Balance)
		pr.Get("/billing/wallet/history", h.History)
	})
	// Spend/Credit/CreateOrder с idempotency. Idempotency-Key читается из header.
	r.Group(func(pr chi.Router) {
		pr.Use(authMW)
		pr.Use(idemMW)
		pr.Post("/billing/wallet/spend", h.Spend)
		pr.Post("/billing/wallet/credit", h.Credit)
		pr.Post("/billing/orders", h.CreateOrder)
	})

	// Webhook от платёжного шлюза. Публичный (защищён HMAC-подписью).
	// Путь зависит от провайдера: /billing/webhooks/mock, /billing/webhooks/robokassa и т.д.
	r.Post("/billing/webhooks/"+gw.Name(), wh.Handle)

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
