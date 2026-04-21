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
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

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
	// Title для body — читать из defs не обязательно, положим в
	// тело сам key; UI рисует через i18n.
	if _, err := exec.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, 2, $3, $4)
	`, ids.New(), userID,
		fmt.Sprintf("Достижение: %s", key),
		fmt.Sprintf("Открыто новое достижение: %s.", key),
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
//                    buildings.unit_id=1 (metal_mine) level>=1.
//   FIRST_SILICON  — аналогично для unit_id=2.
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
		{"FIRST_METAL", `
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = 1 AND b.level >= 1
			)`},
		{"FIRST_SILICON", `
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = 2 AND b.level >= 1
			)`},
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
		{"FIRST_EXPEDITION", `SELECT EXISTS (
				SELECT 1 FROM fleets WHERE owner_user_id = $1 AND mission = 15
			)`},
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
		{"SPY_SUCCESS", `SELECT EXISTS (
				SELECT 1 FROM fleets WHERE owner_user_id = $1 AND mission = 11
					AND state IN ('returning','done')
			)`},
		{"RECYCLING", `SELECT EXISTS (
				SELECT 1 FROM fleets WHERE owner_user_id = $1 AND mission = 9
					AND state IN ('returning','done')
			)`},
		{"ROCKET_LAUNCH", `SELECT EXISTS (
				SELECT 1 FROM events
				WHERE user_id = $1 AND kind = 16
			)`},
		{"SCORE_1000", `SELECT (score >= 1000) FROM users WHERE id = $1`},
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

// List возвращает все defs + флаг unlocked + timestamp (если есть).
func (s *Service) List(ctx context.Context, userID string) ([]Entry, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT d.key, d.title, d.description, d.points, a.unlocked_at
		FROM achievement_defs d
		LEFT JOIN achievements_user a
		  ON a.achievement = d.key AND a.user_id = $1
		ORDER BY d.points ASC, d.key ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("achievements list: %w", err)
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.Key, &e.Title, &e.Description, &e.Points, &e.UnlockedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Entry — одна строчка достижения для UI.
type Entry struct {
	Key         string     `json:"key"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Points      int        `json:"points"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
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
