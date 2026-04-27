// Command resync — CLI для ручного пересчёта artefact-факторов.
//
// Использование:
//   resync --user=<uuid>            # один пользователь
//   resync --all                    # все пользователи (бережёт БД,
//                                     идёт батчами по 100)
//
// Это аналог oxsar2 CronCommand::actionResyncArtefacts +
// scripts/cron-win32/ResyncArtefacts.bat (§5.10.1 ТЗ).
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"oxsar/game-nova/internal/artefact"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "resync:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		userID = flag.String("user", "", "UUID пользователя для пересчёта")
		all    = flag.Bool("all", false, "пересчитать всех пользователей (осторожно)")
	)
	flag.Parse()

	if *userID == "" && !*all {
		return fmt.Errorf("нужен либо --user=<uuid>, либо --all")
	}

	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	catalogDir := os.Getenv("CATALOG_DIR")
	if catalogDir == "" {
		catalogDir = "../../configs"
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
	svc := artefact.NewService(db, cat)

	if *userID != "" {
		if err := svc.ResyncUser(ctx, *userID); err != nil {
			return fmt.Errorf("resync %s: %w", *userID, err)
		}
		log.InfoContext(ctx, "resync ok", slog.String("user", *userID))
		return nil
	}

	// --all: читаем всех игроков батчем. Keyset pagination.
	var lastID string
	const batchSize = 100
	totalDone := 0
	for {
		rows, err := pool.Query(ctx, `
			SELECT id FROM users
			WHERE deleted_at IS NULL AND id > $1
			ORDER BY id
			LIMIT $2
		`, lastID, batchSize)
		if err != nil {
			return fmt.Errorf("list users: %w", err)
		}
		var batch []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			batch = append(batch, id)
		}
		rows.Close()
		if len(batch) == 0 {
			break
		}
		for _, id := range batch {
			if err := svc.ResyncUser(ctx, id); err != nil {
				log.WarnContext(ctx, "resync failed",
					slog.String("user", id), slog.String("err", err.Error()))
				continue
			}
			totalDone++
		}
		lastID = batch[len(batch)-1]
	}
	log.InfoContext(ctx, "resync all done", slog.Int("users", totalDone))
	return nil
}
