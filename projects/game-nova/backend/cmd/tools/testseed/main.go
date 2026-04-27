// Command testseed наполняет БД детерминированным набором данных для E2E-тестов.
//
// Сид содержит 5 игроков с фиксированными UUID, паролями и координатами:
//   - admin    (superadmin, UUID 0000…01) — для админских сценариев
//   - alice    (новичок,    UUID 0000…02) — пустые состояния
//   - bob      (прокачан,   UUID 0000…03) — полные состояния, много данных
//   - eve      (жертва,     UUID 0000…04) — слабая планета рядом с bob для атак
//   - charlie  (союзник,    UUID 0000…05) — член альянса bob'а
//
// Пароль у всех: `test-password-123`.
//
// Использование:
//   go run ./cmd/tools/testseed            — сидить поверх (INSERT … ON CONFLICT)
//   go run ./cmd/tools/testseed --reset    — сначала очистить игровые таблицы
//
// Требует DB_URL. Не трогает миграции, конфиг-справочники и служебные таблицы.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
)

// Фиксированные UUID — важны для E2E (чтобы Playwright мог подставлять ID в URL).
const (
	uidAdmin   = "00000000-0000-0000-0000-000000000001"
	uidAlice   = "00000000-0000-0000-0000-000000000002"
	uidBob     = "00000000-0000-0000-0000-000000000003"
	uidEve     = "00000000-0000-0000-0000-000000000004"
	uidCharlie = "00000000-0000-0000-0000-000000000005"

	pidAdmin   = "00000000-0000-0000-0000-0000000000a1"
	pidAlice   = "00000000-0000-0000-0000-0000000000a2"
	pidBob     = "00000000-0000-0000-0000-0000000000a3"
	pidEve     = "00000000-0000-0000-0000-0000000000a4"
	pidCharlie = "00000000-0000-0000-0000-0000000000a5"

	aidUT = "00000000-0000-0000-0000-0000000000b1" // alliance [UT]

	testPassword = "test-password-123"
)

func main() {
	reset := flag.Bool("reset", false, "truncate игровые таблицы перед сидом")
	flag.Parse()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "testseed: DB_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testseed: pool:", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Retry ping — Docker embedded DNS иногда misbehaved при пике нагрузки.
	// Полная выдержка: 0.5+1+2+4+8+16 ≈ 31s.
	{
		delay := 500 * time.Millisecond
		var pingErr error
		for attempt := 1; attempt <= 6; attempt++ {
			if pingErr = pool.Ping(ctx); pingErr == nil {
				break
			}
			slog.Warn("testseed: pg ping failed, retrying", "attempt", attempt, "err", pingErr.Error())
			time.Sleep(delay)
			delay *= 2
		}
		if pingErr != nil {
			fmt.Fprintln(os.Stderr, "testseed: ping:", pingErr)
			os.Exit(1)
		}
	}

	if *reset {
		if err := truncateGameTables(ctx, pool); err != nil {
			fmt.Fprintln(os.Stderr, "testseed: reset:", err)
			os.Exit(1)
		}
		slog.Info("testseed: reset done")
	}

	hash, err := auth.HashPassword(testPassword)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testseed: hash:", err)
		os.Exit(1)
	}

	if err := seed(ctx, pool, hash); err != nil {
		fmt.Fprintln(os.Stderr, "testseed: seed:", err)
		os.Exit(1)
	}

	fmt.Printf("testseed: ok — 5 users, 4 planets, 1 alliance [UT], 1 message\n")
	fmt.Printf("  password for all: %s\n", testPassword)
}

// truncateGameTables сбрасывает только «игровые» таблицы — миграции и
// справочники оставляем. Порядок — от зависимых к независимым; CASCADE
// добивает всё, что ссылается.
func truncateGameTables(ctx context.Context, pool *pgxpool.Pool) error {
	tables := []string{
		"chat_messages",
		"messages",
		"battle_reports",
		"espionage_reports",
		"expedition_reports",
		"fleets",
		"events",
		"ships",
		"buildings",
		"research",
		"artefacts",
		"artefact_market_offers",
		"market_lots",
		"officers",
		"repair_queue",
		"alliance_members",
		"alliance_relationships",
		"alliances",
		"credit_purchases",
		"res_log",
		"resource_transfers",
		"debris_fields",
		"notepad",
		"friends",
		"user_settings",
		"planets",
		"users",
	}
	for _, t := range tables {
		if _, err := pool.Exec(ctx, "TRUNCATE TABLE "+t+" RESTART IDENTITY CASCADE"); err != nil {
			// Таблицы могут отсутствовать в dev-базе — не падаем, просто логируем.
			slog.Warn("testseed: truncate skipped", "table", t, "err", err.Error())
		}
	}
	return nil
}

func seed(ctx context.Context, pool *pgxpool.Pool, pwHash string) error {
	// --- users ---
	users := []struct {
		id, username, email, role string
		credit                    float64
	}{
		{uidAdmin, "admin", "admin@test.local", "superadmin", 10000},
		{uidAlice, "alice", "alice@test.local", "player", 100},
		// bob — superadmin для удобства ручного тестирования:
		// один аккаунт со всеми фичами (прокачка + админка).
		{uidBob, "bob", "bob@test.local", "superadmin", 5000},
		{uidEve, "eve", "eve@test.local", "player", 50},
		{uidCharlie, "charlie", "charlie@test.local", "player", 1000},
	}
	for _, u := range users {
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, role, credit, language)
			VALUES ($1, $2, $3, $4, $5::user_role, $6, 'ru')
			ON CONFLICT (id) DO UPDATE
			  SET password_hash = EXCLUDED.password_hash,
			      role          = EXCLUDED.role,
			      credit        = EXCLUDED.credit
		`, u.id, u.username, u.email, pwHash, u.role, u.credit); err != nil {
			return fmt.Errorf("insert user %s: %w", u.username, err)
		}
	}

	// --- planets ---
	// Фиксированные координаты: alice/bob в одной системе, eve рядом с bob,
	// charlie в соседней системе.
	planets := []struct {
		id, userID, name  string
		g, sys, pos       int
		metal, si, hy     int64
		diameter          int
	}{
		{pidAdmin, uidAdmin, "Admin-Home", 1, 1, 3, 1000000, 500000, 100000, 18800},
		{pidAlice, uidAlice, "Alice-Home", 1, 1, 5, 1000, 500, 0, 18800},
		{pidBob, uidBob, "Bob-Home", 1, 1, 7, 9000000, 5000000, 2000000, 18800},
		{pidEve, uidEve, "Eve-Home", 1, 1, 9, 500, 200, 0, 12000},
		{pidCharlie, uidCharlie, "Charlie-Home", 1, 2, 7, 500000, 300000, 100000, 18800},
	}
	for _, p := range planets {
		if _, err := pool.Exec(ctx, `
			INSERT INTO planets (id, user_id, is_moon, name, galaxy, system, position,
			                     diameter, used_fields, planet_type, temperature_min, temperature_max,
			                     metal, silicon, hydrogen)
			VALUES ($1, $2, false, $3, $4, $5, $6, $7, 0, 'normaltempplanet', -20, 40, $8, $9, $10)
			ON CONFLICT (id) DO UPDATE
			  SET metal    = EXCLUDED.metal,
			      silicon  = EXCLUDED.silicon,
			      hydrogen = EXCLUDED.hydrogen
		`, p.id, p.userID, p.name, p.g, p.sys, p.pos, p.diameter, p.metal, p.si, p.hy); err != nil {
			return fmt.Errorf("insert planet %s: %w", p.name, err)
		}
		if _, err := pool.Exec(ctx,
			`UPDATE users SET cur_planet_id=$1 WHERE id=$2 AND cur_planet_id IS NULL`,
			p.id, p.userID); err != nil {
			return fmt.Errorf("set cur_planet %s: %w", p.userID, err)
		}
	}

	// --- buildings ---
	// alice — минимум (старт + 1 metal_mine); bob — качественно прокачан.
	type bld struct{ planetID string; unitID, level int }
	buildings := []bld{
		// alice
		{pidAlice, 1, 1},   // metal_mine
		{pidAlice, 2, 1},   // silicon_lab
		{pidAlice, 4, 1},   // solar_plant
		// bob — full rig
		{pidBob, 1, 20}, {pidBob, 2, 18}, {pidBob, 3, 15},
		{pidBob, 4, 20}, {pidBob, 6, 10}, {pidBob, 8, 10},
		{pidBob, 12, 10}, {pidBob, 100, 5}, {pidBob, 101, 5},
		// eve — слабая
		{pidEve, 1, 3}, {pidEve, 4, 3},
		// charlie
		{pidCharlie, 1, 10}, {pidCharlie, 4, 10}, {pidCharlie, 8, 5},
	}
	for _, b := range buildings {
		if _, err := pool.Exec(ctx, `
			INSERT INTO buildings (planet_id, unit_id, level) VALUES ($1,$2,$3)
			ON CONFLICT (planet_id, unit_id) DO UPDATE SET level=EXCLUDED.level
		`, b.planetID, b.unitID, b.level); err != nil {
			return fmt.Errorf("insert building: %w", err)
		}
	}

	// План 23: синхронизируем used_fields = число записей в buildings.
	// solar_satellite (id=39) не живёт в buildings, поэтому COUNT(*) норм.
	if _, err := pool.Exec(ctx, `
		UPDATE planets p
		SET used_fields = COALESCE(
			(SELECT COUNT(*) FROM buildings b WHERE b.planet_id = p.id), 0)
	`); err != nil {
		return fmt.Errorf("resync used_fields: %w", err)
	}

	// --- research (у bob — до гипердвигателя 2) ---
	type res struct{ userID string; unitID, level int }
	researches := []res{
		{uidBob, 14, 10},  // computer_tech
		{uidBob, 18, 8},   // energy_tech
		{uidBob, 20, 6},   // combustion_engine
		{uidBob, 21, 4},   // impulse_engine
		{uidBob, 22, 2},   // hyperspace_engine
		{uidCharlie, 14, 3},
	}
	for _, r := range researches {
		if _, err := pool.Exec(ctx, `
			INSERT INTO research (user_id, unit_id, level) VALUES ($1,$2,$3)
			ON CONFLICT (user_id, unit_id) DO UPDATE SET level=EXCLUDED.level
		`, r.userID, r.unitID, r.level); err != nil {
			return fmt.Errorf("insert research: %w", err)
		}
	}

	// --- ships (bob — флот; eve — минимум) ---
	type ship struct{ planetID string; unitID, count int }
	ships := []ship{
		{pidBob, 30, 50}, {pidBob, 31, 100}, {pidBob, 35, 20},
		{pidBob, 36, 5}, {pidBob, 37, 20},
		{pidEve, 31, 5},
		{pidCharlie, 30, 10},
	}
	for _, s := range ships {
		if _, err := pool.Exec(ctx, `
			INSERT INTO ships (planet_id, unit_id, count) VALUES ($1,$2,$3)
			ON CONFLICT (planet_id, unit_id) DO UPDATE SET count=EXCLUDED.count
		`, s.planetID, s.unitID, s.count); err != nil {
			return fmt.Errorf("insert ship: %w", err)
		}
	}

	// --- alliance [UT] — charlie leader + bob member ---
	if _, err := pool.Exec(ctx, `
		INSERT INTO alliances (id, tag, name, description, owner_id)
		VALUES ($1, 'UT', 'UI Testers', 'E2E test alliance', $2)
		ON CONFLICT (id) DO UPDATE SET tag=EXCLUDED.tag, name=EXCLUDED.name
	`, aidUT, uidCharlie); err != nil {
		return fmt.Errorf("insert alliance: %w", err)
	}
	for _, m := range []struct{ uid, rank string }{
		{uidCharlie, "owner"},
		{uidBob, "member"},
	} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO alliance_members (alliance_id, user_id, rank)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id) DO UPDATE SET alliance_id=EXCLUDED.alliance_id, rank=EXCLUDED.rank
		`, aidUT, m.uid, m.rank); err != nil {
			return fmt.Errorf("insert alliance_member: %w", err)
		}
		if _, err := pool.Exec(ctx,
			`UPDATE users SET alliance_id=$1 WHERE id=$2`, aidUT, m.uid); err != nil {
			return fmt.Errorf("set user.alliance_id: %w", err)
		}
	}

	// --- welcome-сообщение alice ---
	if _, err := pool.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, 0, 'Добро пожаловать', 'Это тестовое сообщение.')
		ON CONFLICT (id) DO NOTHING
	`, "00000000-0000-0000-0000-0000000000c1", uidAlice); err != nil {
		return fmt.Errorf("insert welcome message: %w", err)
	}

	return nil
}
