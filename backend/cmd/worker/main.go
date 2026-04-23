// Command worker — фоновый обработчик event-loop.
//
// Запускается отдельным процессом от сервера — так можно масштабировать
// воркеры независимо и перезапускать без простоя API.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/achievement"
	"github.com/oxsar/nova/backend/internal/alien"
	"github.com/oxsar/nova/backend/internal/artefact"
	"github.com/oxsar/nova/backend/internal/automsg"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/fleet"
	"github.com/oxsar/nova/backend/internal/officer"
	"github.com/oxsar/nova/backend/internal/planet"
	"github.com/oxsar/nova/backend/internal/repair"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/internal/requirements"
	"github.com/oxsar/nova/backend/internal/rocket"
	"github.com/oxsar/nova/backend/internal/score"
	"github.com/oxsar/nova/backend/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker exit", slog.String("err", err.Error()))
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

	pool, err := storage.OpenPostgres(ctx, cfg.DB.URL)
	if err != nil {
		return err
	}
	defer pool.Close()

	db := repo.New(pool)
	w := event.NewWorker(db, log)
	artefactSvc := artefact.NewService(db, cat)
	transportSvc := fleet.NewTransportService(db, cat, cfg.Game.Speed, artefactSvc)

	// repair.Service нужен только ради DisassembleHandler; сами
	// enqueue-операции идут через server. Конструктор требует полный
	// набор зависимостей — даём их тем же способом, что и в server/main.
	planetRepo := planet.NewRepository(pool)
	planetSvc := planet.NewService(db, planetRepo, cat)
	reqs := requirements.New(cat)
	repairSvc := repair.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	rocketSvc := rocket.NewService(db, cat, cfg.Game.Speed)
	officerSvc := officer.NewService(db)

	scoreSvc := score.NewServiceWithCoeffs(db, cat, cfg.Game.Points)
	achSvc := achievement.NewService(db)

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

	// Регистрация handler-ов.
	// Один handler на Kind. Domain-пакеты сами не регистрируются —
	// чтобы воркер видел весь список в одном месте и было проще
	// отслеживать, что именно обрабатывается.
	w.Register(event.KindBuildConstruction, withAchievement(withScore(event.HandleBuildConstruction)))
	w.Register(event.KindResearch, withAchievement(withScore(event.HandleResearch)))
	w.Register(event.KindBuildFleet, withAchievement(withScore(event.HandleBuildFleet)))
	w.Register(event.KindBuildDefense, withScore(event.HandleBuildFleet))
	w.Register(event.KindArtefactExpire, withAchievement(artefactSvc.ExpireEvent()))
	w.Register(event.KindArtefactDelay, artefactSvc.DelayEvent())
	w.Register(event.KindTransport, transportSvc.ArriveHandler())
	w.Register(event.KindReturn, transportSvc.ReturnHandler())
	w.Register(event.KindAttackSingle, withAchievement(transportSvc.AttackHandler()))
	w.Register(event.KindAttackAlliance, withAchievement(transportSvc.ACSAttackHandler()))
	w.Register(event.KindRaidWarning, transportSvc.RaidWarningHandler())
	w.Register(event.KindRecycling, withAchievement(transportSvc.RecyclingHandler()))
	w.Register(event.KindSpy, withAchievement(transportSvc.SpyHandler()))
	w.Register(event.KindColonize, withAchievement(transportSvc.ColonizeHandler()))
	w.Register(event.KindDisassemble, repairSvc.DisassembleHandler())
	w.Register(event.KindRepair, repairSvc.RepairHandler())
	w.Register(event.KindRocketAttack, withAchievement(rocketSvc.ImpactHandler()))
	w.Register(event.KindExpedition, withAchievement(transportSvc.ExpeditionHandler()))
	w.Register(event.KindOfficerExpire, officerSvc.ExpireHandler())

	alienSvc := alien.NewService(db, cat)
	w.Register(event.KindAlienAttack, alienSvc.AttackHandler())

	automsgSvc := automsg.NewService(db)

	// Alien AI: спавн атак раз в 6 часов.
	go func() {
		t := time.NewTicker(6 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := alienSvc.Spawn(ctx); err != nil {
					log.ErrorContext(ctx, "alien_spawn_failed", slog.String("err", err.Error()))
				}
			}
		}
	}()

	// Периодический пересчёт очков всех игроков (раз в 5 минут).
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := scoreSvc.RecalcAll(ctx, log); err != nil {
					log.ErrorContext(ctx, "score_recalc_all_failed", slog.String("err", err.Error()))
				}
			}
		}
	}()

	// Ежедневная рассылка уведомлений о неактивности.
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case tick := <-t.C:
				year, week := tick.ISOWeek()
				weekSuffix := fmt.Sprintf("%dW%02d", year, week)
				n, err := automsgSvc.SendInactivityReminders(ctx, 3, weekSuffix)
				if err != nil {
					log.ErrorContext(ctx, "inactivity_reminders_failed", slog.String("err", err.Error()))
				} else if n > 0 {
					log.InfoContext(ctx, "inactivity_reminders_sent", slog.Int("count", n))
				}
			}
		}
	}()

	// Удаление временных планет с истёкшим expires_at (раз в час).
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				tag, err := pool.Exec(ctx,
					`DELETE FROM planets WHERE expires_at IS NOT NULL AND expires_at < now()`)
				if err != nil {
					log.ErrorContext(ctx, "expire_planets_failed", slog.String("err", err.Error()))
				} else if tag.RowsAffected() > 0 {
					log.InfoContext(ctx, "expire_planets_deleted",
						slog.Int64("count", tag.RowsAffected()))
				}
			}
		}
	}()

	log.InfoContext(ctx, "worker started")
	if err := w.Run(ctx); err != nil && err != context.Canceled {
		return err
	}
	return nil
}
