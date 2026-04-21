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

	"github.com/oxsar/nova/backend/internal/artefact"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/fleet"
	"github.com/oxsar/nova/backend/internal/planet"
	"github.com/oxsar/nova/backend/internal/repair"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/internal/requirements"
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
	transportSvc := fleet.NewTransportService(db, cat, cfg.Game.Speed)

	// repair.Service нужен только ради DisassembleHandler; сами
	// enqueue-операции идут через server. Конструктор требует полный
	// набор зависимостей — даём их тем же способом, что и в server/main.
	planetRepo := planet.NewRepository(pool)
	planetSvc := planet.NewService(db, planetRepo, cat)
	reqs := requirements.New(cat)
	repairSvc := repair.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)

	// Регистрация handler-ов.
	// Один handler на Kind. Domain-пакеты сами не регистрируются —
	// чтобы воркер видел весь список в одном месте и было проще
	// отслеживать, что именно обрабатывается.
	w.Register(event.KindBuildConstruction, event.HandleBuildConstruction)
	w.Register(event.KindResearch, event.HandleResearch)
	w.Register(event.KindBuildFleet, event.HandleBuildFleet)
	w.Register(event.KindBuildDefense, event.HandleBuildFleet)
	w.Register(event.KindArtefactExpire, artefactSvc.ExpireEvent())
	w.Register(event.KindTransport, transportSvc.ArriveHandler())
	w.Register(event.KindReturn, transportSvc.ReturnHandler())
	w.Register(event.KindAttackSingle, transportSvc.AttackHandler())
	w.Register(event.KindRecycling, transportSvc.RecyclingHandler())
	w.Register(event.KindDisassemble, repairSvc.DisassembleHandler())
	w.Register(event.KindRepair, repairSvc.RepairHandler())

	log.InfoContext(ctx, "worker started")
	if err := w.Run(ctx); err != nil && err != context.Canceled {
		return err
	}
	return nil
}
