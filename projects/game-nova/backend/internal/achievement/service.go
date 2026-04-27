// Package achievement — пассивные награды за прогресс.
//
// UnlockIfNew(userID, key) — идемпотентный вызов из domain-handler'ов
// («открылось первый раз — пишем; повторно — тишина»). Вставка
// ON CONFLICT DO NOTHING + message в inbox при фактическом insert'е.
//
// Подход без отдельных event'ов: прогресс считается «в момент
// действия» внутри уже существующих транзакций. Это проще, чем
// batch-проверка по крону.
package achievement

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

type Service struct {
	db     repo.Exec
	bundle *i18n.Bundle
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
}

func (s *Service) tr(group, key string, vars map[string]string) string {
	if s.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return s.bundle.Tr(i18n.LangRu, group, key, vars)
}

// UnlockIfNew открывает достижение userID. Если уже открыто —
// ничего. При открытии пишет message (folder=2).
//
// tx опционален: если nil, используется autopool. В подавляющем
// большинстве вызовов handler уже внутри транзакции — тогда передаём
// её, чтобы unlock был атомарен с основной операцией.
func (s *Service) UnlockIfNew(ctx context.Context, tx pgx.Tx, userID, key string) error {
	exec := txOrPool(s, tx)
	tag, err := exec.Exec(ctx, `
		INSERT INTO achievements_user (user_id, achievement, unlocked_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, achievement) DO NOTHING
	`, userID, key, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("unlock %s: %w", key, err)
	}
	if tag.RowsAffected() == 0 {
		return nil // уже было
	}
	// Начислить кредиты за достижение.
	if _, err := exec.Exec(ctx,
		`UPDATE users SET credit=credit+$1 WHERE id=$2`,
		economy.CreditAchievement, userID,
	); err != nil {
		return fmt.Errorf("unlock credit: %w", err)
	}
	// Title для body — читать из defs не обязательно, положим в
	// тело сам key; UI рисует через i18n.
	if _, err := exec.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, 2, $3, $4)
	`, ids.New(), userID,
		s.tr("achievement", "subject", map[string]string{"key": key}),
		s.tr("achievement", "body", map[string]string{"key": key, "credits": strconv.FormatInt(economy.CreditAchievement, 10)}),
	); err != nil {
		return fmt.Errorf("unlock message: %w", err)
	}
	return nil
}

// Unlocked — запись о полученном достижении.
type Unlocked struct {
	Key         string    `json:"key"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Points      int       `json:"points"`
	UnlockedAt  time.Time `json:"unlocked_at"`
}

// Def — каталожный элемент (может быть ещё не unlocked).
type Def struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Points      int    `json:"points"`
}

// CheckAll пробегает по известным правилам и открывает всё, что
// заслужено (идемпотентно через ON CONFLICT). Вызывается перед List,
// чтобы UI всегда видел актуальный статус без инвазивных триггеров
// в domain-handler'ах.
//
// Правила (MVP):
//   FIRST_METAL    — у пользователя хоть одна планета с
//                    buildings.unit_id=IDMetalmine level>=1.
//   FIRST_SILICON  — аналогично для IDSiliconLab.
//   FIRST_ARTEFACT — есть хотя бы одна запись в artefacts_user.
//   FIRST_WIN      — есть battle_reports winner='attackers' где
//                    attacker_user_id = userID.
//   FIRST_COLONY   — у пользователя >=2 планет (стартовая + колония).
func (s *Service) CheckAll(ctx context.Context, userID string) error {
	type check struct {
		key string
		sql string
	}
	checks := []check{
		{"FIRST_METAL", fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = %d AND b.level >= 1
			)`, economy.IDMetalmine)},
		{"FIRST_SILICON", fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = %d AND b.level >= 1
			)`, economy.IDSiliconLab)},
		{"FIRST_ARTEFACT", `SELECT EXISTS (SELECT 1 FROM artefacts_user WHERE user_id = $1)`},
		{"FIRST_WIN", `SELECT EXISTS (
				SELECT 1 FROM battle_reports
				WHERE attacker_user_id = $1 AND winner = 'attackers'
			)`},
		{"FIRST_COLONY", `
			SELECT (COUNT(*) >= 2) FROM planets
			WHERE user_id = $1 AND destroyed_at IS NULL AND is_moon = false
		`},
		{"FIRST_FLEET", `SELECT EXISTS (SELECT 1 FROM fleets WHERE owner_user_id = $1)`},
		{"FIRST_EXPEDITION", fmt.Sprintf(`SELECT EXISTS (
				SELECT 1 FROM fleets WHERE owner_user_id = $1 AND mission = %d
			)`, int(event.KindExpedition))},
		{"FIRST_RESEARCH", `SELECT EXISTS (SELECT 1 FROM research WHERE user_id = $1 AND level >= 1)`},
		{"BATTLE_10", `
			SELECT (COUNT(*) >= 10) FROM battle_reports
			WHERE attacker_user_id = $1 AND winner = 'attackers'
		`},
		{"FLEET_50", `
			SELECT (COALESCE(SUM(s.count), 0) >= 50)
			FROM ships s JOIN planets p ON p.id = s.planet_id
			WHERE p.user_id = $1
		`},
		{"ARTEFACT_MARKET", `SELECT EXISTS (
				SELECT 1 FROM artefacts_user
				WHERE user_id = $1 AND acquired_at IS NOT NULL
					AND state IN ('held','active','delayed','listed')
			)`},
		{"SPY_SUCCESS", fmt.Sprintf(`SELECT EXISTS (
				SELECT 1 FROM fleets WHERE owner_user_id = $1 AND mission = %d
					AND state IN ('returning','done')
			)`, int(event.KindSpy))},
		{"RECYCLING", fmt.Sprintf(`SELECT EXISTS (
				SELECT 1 FROM fleets WHERE owner_user_id = $1 AND mission = %d
					AND state IN ('returning','done')
			)`, int(event.KindRecycling))},
		{"ROCKET_LAUNCH", fmt.Sprintf(`SELECT EXISTS (
				SELECT 1 FROM events
				WHERE user_id = $1 AND kind = %d
			)`, int(event.KindRocketAttack))},
		{"SCORE_1000", `SELECT (COALESCE(points,0) >= 1000) FROM users WHERE id = $1`},
	}
	for _, c := range checks {
		var ok bool
		if err := s.db.Pool().QueryRow(ctx, c.sql, userID).Scan(&ok); err != nil {
			return fmt.Errorf("check %s: %w", c.key, err)
		}
		if ok {
			if err := s.UnlockIfNew(ctx, nil, userID, c.key); err != nil {
				return err
			}
		}
	}
	return nil
}

// CheckAllStarter пробегает по стартовым достижениям (Tutorial-цепочка).
// Вызывается после CheckAll для завершённости.
func (s *Service) CheckAllStarter(ctx context.Context, userID string) error {
	type check struct {
		key string
		sql string
	}
	// Fix 2026-04-26: STARTER_BUILD_SOLARPLANT/METALLURGY/SHIPYARD/LAB
	// исторически проверяли unit_id 3/4/21/22, что соответствовало
	// HydrogenLab/SolarPlant/ImpulseEngine/HyperspaceEngine — не совпадало
	// с именами достижений. STARTER_BUILD_SHIPYARD и STARTER_BUILD_LAB
	// при этом никогда не разблокировались (unit_id 21/22 — research, не
	// buildings). Сверка с legacy oxsar2 (sql/tutorial.sql) показала
	// ожидаемые ID: SolarPlant=4, SiliconLab=2 (металлургический), Shipyard=8,
	// ResearchLab=12. Исправлено.
	checks := []check{
		{"STARTER_BUILD_METALMINE", fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = %d AND b.level >= 1
			)`, economy.IDMetalmine)},
		{"STARTER_BUILD_SOLARPLANT", fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = %d AND b.level >= 1
			)`, economy.IDSolarPlant)},
		{"STARTER_BUILD_METALLURGY", fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = %d AND b.level >= 1
			)`, economy.IDSiliconLab)},
		{"STARTER_BUILD_SHIPYARD", fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = %d AND b.level >= 1
			)`, economy.IDShipyard)},
		{"STARTER_BUILD_LAB", fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = %d AND b.level >= 1
			)`, economy.IDResearchLab)},
		{"STARTER_RESEARCH_TECH", `
			SELECT EXISTS (
				SELECT 1 FROM research WHERE user_id = $1 AND level >= 1
			)`},
		{"STARTER_BUILD_SHIP", `
			SELECT EXISTS (
				SELECT 1 FROM ships WHERE planet_id IN (
					SELECT id FROM planets WHERE user_id = $1
				) AND count >= 1
			)`},
		{"STARTER_SEND_MISSION", fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM fleets WHERE owner_user_id = $1 AND mission = %d
			)`, int(event.KindAttackSingle))},
	}
	for _, c := range checks {
		var ok bool
		if err := s.db.Pool().QueryRow(ctx, c.sql, userID).Scan(&ok); err != nil {
			return fmt.Errorf("check starter %s: %w", c.key, err)
		}
		if ok {
			if err := s.UnlockIfNew(ctx, nil, userID, c.key); err != nil {
				return err
			}
		}
	}
	return nil
}

// progressCheck описывает числовое достижение с прогресс-баром.
type progressCheck struct {
	key string
	sql string // возвращает единственный int
	max int
}

var progressChecks = []progressCheck{
	{"BATTLE_10", `SELECT COUNT(*) FROM battle_reports WHERE attacker_user_id=$1 AND winner='attackers'`, 10},
	{"FLEET_50", `SELECT COALESCE(SUM(s.count),0) FROM ships s JOIN planets p ON p.id=s.planet_id WHERE p.user_id=$1`, 50},
	{"SCORE_1000", `SELECT COALESCE(points,0) FROM users WHERE id=$1`, 1000},
}

// List возвращает все defs + флаг unlocked + timestamp (если есть) + прогресс для числовых.
func (s *Service) List(ctx context.Context, userID string) ([]Entry, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT d.key, d.title, d.description, d.points, d.category, a.unlocked_at
		FROM achievement_defs d
		LEFT JOIN achievements_user a
		  ON a.achievement = d.key AND a.user_id = $1
		ORDER BY d.category ASC, d.points ASC, d.key ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("achievements list: %w", err)
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.Key, &e.Title, &e.Description, &e.Points, &e.Category, &e.UnlockedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Подтягиваем прогресс для числовых достижений.
	progress := make(map[string]int, len(progressChecks))
	for _, pc := range progressChecks {
		var val int
		if err := s.db.Pool().QueryRow(ctx, pc.sql, userID).Scan(&val); err != nil {
			continue // не прерываем; просто нет прогресса
		}
		progress[pc.key] = val
	}
	for i, e := range out {
		for _, pc := range progressChecks {
			if e.Key == pc.key {
				val := progress[pc.key]
				max := pc.max
				out[i].Progress = &val
				out[i].ProgressMax = &max
				break
			}
		}
	}
	return out, nil
}

// Entry — одна строчка достижения для UI.
type Entry struct {
	Key         string     `json:"key"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Points      int        `json:"points"`
	Category    string     `json:"category"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
	// Progress/ProgressMax — для числовых достижений (BATTLE_10, FLEET_50, SCORE_1000).
	// Nil — для булевых (FIRST_* и т.п.).
	Progress    *int       `json:"progress,omitempty"`
	ProgressMax *int       `json:"progress_max,omitempty"`
}

// execer — либо *pgx.Tx, либо pool. Позволяет вызвать UnlockIfNew
// из-под существующей транзакции или независимо.
type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func txOrPool(s *Service, tx pgx.Tx) execer {
	if tx != nil {
		return tx
	}
	return poolAdapter{pool: s.db.Pool()}
}

// poolAdapter — тонкий wrapper, чтобы pool удовлетворял execer.
type poolAdapter struct {
	pool poolExecer
}
type poolExecer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (p poolAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, args...)
}
