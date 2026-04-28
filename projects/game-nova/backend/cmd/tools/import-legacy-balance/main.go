// Command import-legacy-balance — оффлайн-импорт балансовых данных
// origin (oxsar2-classic) в configs/balance/origin.yaml + дополнения
// в дефолтные configs/units.yml, ships.yml, rapidfire.yml для алиен-
// и спец-юнитов (план 64 Ф.2; R0-исключение: алиен-юниты применимы во
// всех вселенных, см. roadmap-report «Часть I.5» 2026-04-28).
//
// Подключается к docker-mysql-1 (порт 3307 на хосте — origin-стенд,
// projects/game-origin-php/docker), читает na_construction,
// na_ship_datasheet, na_rapidfire, na_options. DSL-формулы парсятся
// в-process (см. dsl.go) и предвычисляются в таблицы стоимостей по
// уровням 1..50.
//
// Запускается один раз при импорте + регенерация если данные origin
// меняются. CLI НЕ запускается в продакшене — это импорт-инструмент.
// Сгенерированный YAML коммитится в репо.
//
// Запуск:
//   cd projects/game-nova/backend
//   go run ./cmd/tools/import-legacy-balance \
//     --mysql-dsn="root:root_pass@tcp(localhost:3307)/oxsar_db" \
//     --output-balance=../configs/balance/origin.yaml \
//     --output-default-units=../configs/units.yml \
//     --output-default-ships=../configs/ships.yml \
//     --output-default-rapidfire=../configs/rapidfire.yml
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if err := run(); err != nil {
		slog.Error("import-legacy-balance: failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	var (
		dsn         = flag.String("mysql-dsn", "root:root_pass@tcp(localhost:3307)/oxsar_db", "MySQL DSN to docker-mysql-1 (origin stand)")
		// na_rapidfire в docker-mysql-1 пуст (origin фронт использует
		// ext-перекрытия). Данные RF лежат в oxsar2-mysql-1 (legacy
		// dump из oxsar2/sql/new-for-dm/data.sql, тот же namespace
		// na_construction). Если основной DSN вернул 0 RF записей —
		// импортёр fallback'ает на rapidfire-fallback-dsn.
		rapidfireDSN = flag.String("rapidfire-fallback-dsn", "root:root@tcp(localhost:3306)/oxsar_db", "MySQL DSN to oxsar2-mysql-1 (legacy) — used when primary DSN has no rapidfire rows")
		outBalance   = flag.String("output-balance", "../configs/balance/origin.yaml", "destination origin override yaml")
		outUnits     = flag.String("output-default-units", "../configs/units.yml", "default units.yml (R0-exception: append alien/special)")
		outShips     = flag.String("output-default-ships", "../configs/ships.yml", "default ships.yml")
		outRF        = flag.String("output-default-rapidfire", "../configs/rapidfire.yml", "default rapidfire.yml")
		maxLevel     = flag.Int("max-level", 50, "max precomputed level for charge-tables")
		dryRun       = flag.Bool("dry-run", false, "print summary, do not write files")
	)
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}
	log.InfoContext(ctx, "connected to origin mysql", slog.String("dsn", redactDSN(*dsn)))

	cons, err := loadConstructions(ctx, db)
	if err != nil {
		return fmt.Errorf("load na_construction: %w", err)
	}
	log.InfoContext(ctx, "constructions loaded", slog.Int("rows", len(cons)))

	ships, err := loadShipDatasheet(ctx, db)
	if err != nil {
		return fmt.Errorf("load na_ship_datasheet: %w", err)
	}
	log.InfoContext(ctx, "ship_datasheet loaded", slog.Int("rows", len(ships)))

	rf, err := loadRapidfire(ctx, db)
	if err != nil {
		return fmt.Errorf("load na_rapidfire: %w", err)
	}
	if len(rf) == 0 && *rapidfireDSN != "" {
		log.InfoContext(ctx, "primary DSN has no rapidfire rows, trying fallback",
			slog.String("fallback_dsn", redactDSN(*rapidfireDSN)))
		fbDB, err := sql.Open("mysql", *rapidfireDSN)
		if err != nil {
			return fmt.Errorf("open rapidfire fallback: %w", err)
		}
		defer fbDB.Close()
		if err := fbDB.PingContext(ctx); err != nil {
			return fmt.Errorf("ping rapidfire fallback: %w", err)
		}
		rf, err = loadRapidfire(ctx, fbDB)
		if err != nil {
			return fmt.Errorf("load rapidfire fallback: %w", err)
		}
		log.InfoContext(ctx, "rapidfire loaded via fallback", slog.Int("rows", len(rf)))
	} else {
		log.InfoContext(ctx, "rapidfire loaded", slog.Int("rows", len(rf)))
	}

	// Преобразуем в YAML-структуры.
	overrideDoc, err := buildOverride(cons, ships, rf, *maxLevel, log)
	if err != nil {
		return fmt.Errorf("build override: %w", err)
	}

	defaultsDoc, err := buildDefaultExtensions(cons, ships, rf, log)
	if err != nil {
		return fmt.Errorf("build default extensions: %w", err)
	}

	if *dryRun {
		log.InfoContext(ctx, "dry-run: skipping writes",
			slog.Int("override_buildings", len(overrideDoc.Buildings)),
			slog.Int("default_alien_units", len(defaultsDoc.UnitsAppend)),
			slog.Int("default_alien_ships", len(defaultsDoc.ShipsAppend)),
			slog.Int("default_alien_rf", len(defaultsDoc.RapidfireAppend)),
		)
		return nil
	}

	if err := writeOriginOverride(*outBalance, overrideDoc); err != nil {
		return fmt.Errorf("write origin.yaml: %w", err)
	}
	log.InfoContext(ctx, "wrote origin.yaml", slog.String("path", *outBalance))

	if err := appendDefaultUnits(*outUnits, defaultsDoc); err != nil {
		return fmt.Errorf("update default units.yml: %w", err)
	}
	log.InfoContext(ctx, "appended alien/special to default units.yml", slog.String("path", *outUnits))

	if err := appendDefaultShips(*outShips, defaultsDoc); err != nil {
		return fmt.Errorf("update default ships.yml: %w", err)
	}
	log.InfoContext(ctx, "appended alien/special to default ships.yml", slog.String("path", *outShips))

	if err := appendDefaultRapidfire(*outRF, defaultsDoc); err != nil {
		return fmt.Errorf("update default rapidfire.yml: %w", err)
	}
	log.InfoContext(ctx, "appended alien rapidfire to default rapidfire.yml", slog.String("path", *outRF))

	log.InfoContext(ctx, "import-legacy-balance done")
	return nil
}

// redactDSN маскирует пароль для логов.
func redactDSN(dsn string) string {
	at := -1
	for i := 0; i < len(dsn); i++ {
		if dsn[i] == '@' {
			at = i
			break
		}
	}
	if at < 0 {
		return dsn
	}
	colon := -1
	for i := 0; i < at; i++ {
		if dsn[i] == ':' {
			colon = i
		}
	}
	if colon < 0 {
		return dsn
	}
	return dsn[:colon+1] + "***" + dsn[at:]
}
