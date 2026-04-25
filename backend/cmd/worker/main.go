// Command worker — фоновый обработчик event-loop.
//
// Запускается отдельным процессом от сервера — так можно масштабировать
// воркеры независимо и перезапускать без простоя API.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/achievement"
	"github.com/oxsar/nova/backend/internal/alien"
	"github.com/oxsar/nova/backend/internal/artefact"
	"github.com/oxsar/nova/backend/internal/automsg"
	"github.com/oxsar/nova/backend/internal/dailyquest"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/features"
	"github.com/oxsar/nova/backend/internal/health"
	"github.com/oxsar/nova/backend/internal/fleet"
	"github.com/oxsar/nova/backend/internal/officer"
	"github.com/oxsar/nova/backend/internal/planet"
	"github.com/oxsar/nova/backend/internal/repair"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/internal/requirements"
	"github.com/oxsar/nova/backend/internal/rocket"
	"github.com/oxsar/nova/backend/internal/scheduler"
	"github.com/oxsar/nova/backend/internal/score"
	"github.com/oxsar/nova/backend/internal/storage"
	"github.com/oxsar/nova/backend/pkg/metrics"
)

// buildVersion — версия билда. Перебить через -ldflags
// "-X main.buildVersion=1.2.3"; по умолчанию dev.
var buildVersion = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("worker exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	// Signal-контекст отменяется сразу при получении SIGINT/SIGTERM —
	// это сигнал "начать shutdown". Воркер получает его как ctx.Err(),
	// tickLoop останавливается между events. В main() ждём текущий
	// tick до GRACE_PERIOD_SEC (default 30s), потом принудительно
	// закрываем.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	catalogDir := os.Getenv("CATALOG_DIR")
	if catalogDir == "" {
		catalogDir = "../configs"
	}
	cat, err := config.LoadCatalog(catalogDir)
	if err != nil {
		return err
	}

	// Feature flags (план 31 Ф.2). Worker читает тот же набор, что и
	// server — позволяет event-handler'ам ветвиться по флагам
	// (например goal_engine для плана 30).
	featuresPath := os.Getenv("FEATURES_FILE")
	if featuresPath == "" {
		featuresPath = filepath.Join(catalogDir, "features.yaml")
	}
	featureSet, err := features.Load(featuresPath)
	if err != nil {
		log.WarnContext(ctx, "features load failed, using empty set",
			slog.String("path", featuresPath), slog.String("err", err.Error()))
		featureSet, _ = features.ParseBytes(nil)
	}
	log.InfoContext(ctx, "features loaded",
		slog.String("path", featuresPath),
		slog.Any("enabled", features.EnabledKeys(featureSet)))
	_ = featureSet // используется в event-handler'ах при появлении флагов

	pool, err := storage.OpenPostgres(ctx, cfg.DB.URL)
	if err != nil {
		return err
	}
	defer pool.Close()

	db := repo.New(pool)
	w := event.NewWorker(db, log).WithConfig(event.Config{
		Interval:    parseDurEnv("WORKER_INTERVAL", 10*time.Second),
		Batch:       parseIntEnv("WORKER_BATCH", 100),
		MaxBatch:    parseIntEnv("WORKER_MAX_BATCH", 1000),
		MaxAttempts: parseIntEnv("WORKER_MAX_ATTEMPTS", 3),
	})
	artefactSvc := artefact.NewService(db, cat)
	transportSvc := fleet.NewTransportServiceWithConfig(db, cat, cfg.Game.Speed, artefactSvc, cfg.Game.MaxPlanets, cfg.Game.ProtectionPeriod)

	// repair.Service нужен только ради DisassembleHandler; сами
	// enqueue-операции идут через server. Конструктор требует полный
	// набор зависимостей — даём их тем же способом, что и в server/main.
	planetRepo := planet.NewRepository(pool)
	planetSvc := planet.NewServiceWithFactors(db, planetRepo, cat, cfg.Game.StorageFactor, cfg.Game.EnergyProductionFactor)
	reqs := requirements.New(cat)
	repairSvc := repair.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	rocketSvc := rocket.NewService(db, cat, cfg.Game.Speed)
	officerSvc := officer.NewService(db)

	scoreSvc := score.NewServiceWithCoeffs(db, cat, cfg.Game.Points)
	achSvc := achievement.NewService(db)
	dailyQuestSvc := dailyquest.New(pool)

	// withScore оборачивает handler: после успеха пересчитывает очки
	// пользователя события (если UserID задан). Ошибка пересчёта не
	// прерывает основной handler — только логируется.
	withScore := func(h event.Handler) event.Handler {
		return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
			if err := h(ctx, tx, e); err != nil {
				return err
			}
			if e.UserID == nil {
				return nil
			}
			if err := scoreSvc.RecalcUser(ctx, *e.UserID); err != nil {
				log.WarnContext(ctx, "score_recalc_failed",
					slog.String("user_id", *e.UserID),
					slog.String("err", err.Error()))
			}
			return nil
		}
	}

	// withAchievement оборачивает handler: после успеха проверяет и
	// открывает достижения. Ошибка не прерывает основной handler.
	withAchievement := func(h event.Handler) event.Handler {
		return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
			if err := h(ctx, tx, e); err != nil {
				return err
			}
			if e.UserID == nil {
				return nil
			}
			if err := achSvc.CheckAll(ctx, *e.UserID); err != nil {
				log.WarnContext(ctx, "achievement_check_failed",
					slog.String("user_id", *e.UserID),
					slog.String("err", err.Error()))
			}
			return nil
		}
	}

	// План 17 D: withDailyQuest оборачивает handler — после успеха
	// инкрементирует прогресс quest по condition_type.
	// Ошибка не прерывает основной handler.
	withDailyQuest := func(conditionType string) func(event.Handler) event.Handler {
		return func(h event.Handler) event.Handler {
			return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
				if err := h(ctx, tx, e); err != nil {
					return err
				}
				if e.UserID == nil || dailyQuestSvc == nil {
					return nil
				}
				if err := dailyQuestSvc.IncrementProgress(ctx, *e.UserID,
					conditionType, 1, nil); err != nil {
					log.WarnContext(ctx, "daily_quest_progress_failed",
						slog.String("user_id", *e.UserID),
						slog.String("condition", conditionType),
						slog.String("err", err.Error()))
				}
				return nil
			}
		}
	}

	// Регистрация handler-ов.
	// Один handler на Kind. Domain-пакеты сами не регистрируются —
	// чтобы воркер видел весь список в одном месте и было проще
	// отслеживать, что именно обрабатывается.
	w.Register(event.KindBuildConstruction, withDailyQuest("building_done")(withAchievement(withScore(event.HandleBuildConstruction))))
	w.Register(event.KindResearch, withDailyQuest("research_done")(withAchievement(withScore(event.HandleResearch))))
	w.Register(event.KindBuildFleet, withAchievement(withScore(event.HandleBuildFleet)))
	w.Register(event.KindBuildDefense, withScore(event.HandleBuildFleet))
	w.Register(event.KindArtefactExpire, withAchievement(artefactSvc.ExpireEvent()))
	w.Register(event.KindArtefactDelay, artefactSvc.DelayEvent())
	w.Register(event.KindTransport, transportSvc.ArriveHandler())
	w.Register(event.KindPosition, transportSvc.PositionArriveHandler())
	w.Register(event.KindReturn, transportSvc.ReturnHandler())
	w.Register(event.KindAttackSingle, withAchievement(transportSvc.AttackHandler()))
	w.Register(event.KindAttackAlliance, withAchievement(transportSvc.ACSAttackHandler()))
	// План 20 Ф.6: moon destruction — те же handlers с веткой rip-roll.
	w.Register(event.KindAttackDestroyMoon, withAchievement(transportSvc.AttackHandler()))
	w.Register(event.KindAttackAllianceDestroyMoon, withAchievement(transportSvc.ACSAttackHandler()))
	w.Register(event.KindRaidWarning, transportSvc.RaidWarningHandler())
	w.Register(event.KindRecycling, withAchievement(transportSvc.RecyclingHandler()))
	w.Register(event.KindSpy, withAchievement(transportSvc.SpyHandler()))
	w.Register(event.KindColonize, withAchievement(transportSvc.ColonizeHandler()))
	w.Register(event.KindDisassemble, repairSvc.DisassembleHandler())
	w.Register(event.KindRepair, repairSvc.RepairHandler())
	w.Register(event.KindRocketAttack, withAchievement(rocketSvc.ImpactHandler()))
	w.Register(event.KindExpedition, withAchievement(transportSvc.ExpeditionHandler()))
	w.Register(event.KindOfficerExpire, officerSvc.ExpireHandler())
	w.Register(event.KindExpirePlanet, event.HandleExpirePlanet)

	alienSvc := alien.NewService(db, cat)
	w.Register(event.KindAlienAttack, alienSvc.AttackHandler())
	w.Register(event.KindAlienHalt, alienSvc.HaltHandler())
	w.Register(event.KindAlienHolding, alienSvc.HoldingHandler())
	w.Register(event.KindAlienHoldingAI, alienSvc.HoldingAIHandler())

	automsgSvc := automsg.NewService(db)

	// План 32: периодические задачи через scheduler с advisory lock.
	// При N≥2 worker'ах ровно один инстанс выполняет каждую job, остальные
	// тихо инкрементят skip-метрику.
	schedulePath := os.Getenv("SCHEDULE_FILE")
	if schedulePath == "" {
		schedulePath = filepath.Join(catalogDir, "schedule.yaml")
	}
	schedCfg, err := scheduler.LoadConfig(schedulePath)
	if err != nil {
		return fmt.Errorf("load schedule: %w", err)
	}
	sch := scheduler.New(schedCfg, pool, log)

	if err := sch.Register("alien_spawn", alienSvc.Spawn); err != nil {
		return fmt.Errorf("register alien_spawn: %w", err)
	}

	if err := sch.Register("inactivity_reminders", func(ctx context.Context) error {
		year, week := time.Now().UTC().ISOWeek()
		weekSuffix := fmt.Sprintf("%dW%02d", year, week)
		n, err := automsgSvc.SendInactivityReminders(ctx, 3, weekSuffix)
		if err != nil {
			return err
		}
		if n > 0 {
			log.InfoContext(ctx, "inactivity_reminders_sent", slog.Int("count", n))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("register inactivity_reminders: %w", err)
	}

	if err := sch.Register("expire_temp_planets", func(ctx context.Context) error {
		tag, err := pool.Exec(ctx,
			`DELETE FROM planets WHERE expires_at IS NOT NULL AND expires_at < now()`)
		if err != nil {
			return err
		}
		if tag.RowsAffected() > 0 {
			log.InfoContext(ctx, "expire_planets_deleted",
				slog.Int64("count", tag.RowsAffected()))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("register expire_temp_planets: %w", err)
	}

	if err := sch.Register("event_pruner", func(ctx context.Context) error {
		moved, err := w.PruneErrors(ctx)
		if err != nil {
			return err
		}
		if moved > 0 {
			log.InfoContext(ctx, "event_pruned",
				slog.Int64("moved", moved),
				slog.String("threshold", event.DeadLetterThreshold.String()))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("register event_pruner: %w", err)
	}

	// Пересчёт очков всех игроков — теперь через scheduler (план 32 Ф.4).
	// Handler KindScoreRecalcAll оставлен для legacy wait-events,
	// созданных до миграции (через 7 дней миграция удалит их).
	w.Register(event.KindScoreRecalcAll, scoreSvc.RecalcAllEvent())
	if err := sch.Register("score_recalc_all", scoreSvc.RecalcAllScheduled); err != nil {
		return fmt.Errorf("register score_recalc_all: %w", err)
	}

	if err := sch.Start(ctx); err != nil {
		return fmt.Errorf("scheduler start: %w", err)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = sch.Stop(stopCtx)
	}()

	// Обновлятор queue-depth / lag gauge'ов.
	go func() {
		if err := w.RunMetricsUpdater(ctx); err != nil && err != context.Canceled {
			log.ErrorContext(ctx, "event_metrics_exit", slog.String("err", err.Error()))
		}
	}()

	// /metrics HTTP endpoint для Prometheus + /api/health, /api/ready
	// для container healthcheck. План 31 Ф.1.
	healthState := health.NewState("worker", buildVersion)
	healthState.SetReady() // worker готов сразу после открытия pool

	metricsAddr := os.Getenv("WORKER_METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = ":9091"
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Register())
	mux.Handle("/api/health", healthState.HealthHandler())
	mux.Handle("/api/ready", healthState.ReadyHandler(pool))
	metricsSrv := &http.Server{Addr: metricsAddr, Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	go func() {
		log.InfoContext(ctx, "worker metrics listening", slog.String("addr", metricsAddr))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.ErrorContext(ctx, "worker metrics exit", slog.String("err", err.Error()))
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = metricsSrv.Shutdown(shutdownCtx)
	}()

	// При SIGTERM переводим healthcheck в draining-state, чтобы
	// container-orchestrator (docker healthcheck / k8s) увидел 503 на
	// /api/ready и перестал считать инстанс живым. План 31 Ф.1.
	go func() {
		<-ctx.Done()
		healthState.SetDraining()
		log.InfoContext(context.Background(), "worker draining")
	}()

	grace := parseDurEnv("WORKER_SHUTDOWN_GRACE", 30*time.Second)
	log.InfoContext(ctx, "worker started",
		slog.Duration("shutdown_grace", grace))
	if err := w.RunWithGrace(ctx, grace); err != nil && err != context.Canceled {
		return err
	}
	log.InfoContext(context.Background(), "worker stopped")
	return nil
}

func parseIntEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return n
	}
	return def
}

func parseDurEnv(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return d
	}
	return def
}
