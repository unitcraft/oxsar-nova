// Package alien — упрощённый Alien AI (M5.2, §5.18 ТЗ).
//
// Источник: game/AlienAI.class.php (1127 строк).
//
// Упрощения относительно оригинала (см. docs/simplifications.md):
//   - Только один тип события: KindAlienAttack=35 (нет HALT/GRAB_CREDIT/CUSTOM).
//   - Флот инопланетян фиксирован по уровню активности игрока (3 тира).
package alien

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
	"oxsar/game-nova/pkg/rng"
)

// alienHome — фиксированные координаты «дома» инопланетян (глубокий космос).
// Галактика 99 гарантированно вне игровых галактик (1-16).
const (
	alienHomeGalaxy   = 99
	alienHomeSystem   = 500
	alienHomePosition = 8
	// Скорость флота инопланетян (условные единицы, как minSpeed у fleet).
	alienFleetSpeed = 20000
)

// alienPayload — содержимое events.payload для KindAlienAttack.
type alienPayload struct {
	PlanetID string `json:"planet_id"`
	UserID   string `json:"user_id"`
	Tier     int    `json:"tier"` // 1=слабые, 2=средние, 3=сильные
	Galaxy   int    `json:"galaxy"`
	System   int    `json:"system"`
	Position int    `json:"position"`
}

// Service — сервис инопланетян: спавн событий + обработка атаки.
type Service struct {
	db     repo.Exec
	cat    *config.Catalog
	bundle *i18n.Bundle
}

func NewService(db repo.Exec, cat *config.Catalog) *Service {
	return &Service{db: db, cat: cat}
}

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

// Spawn выбирает случайных активных игроков и создаёт события
// KindAlienAttack. Вызывается из воркера раз в N часов.
//
// Логика выбора (приближена к AlienAI::findTarget из legacy):
//   - Берём до N случайных активных игроков (last_seen_at < 7 дней).
//   - С вероятностью 30% для каждого создаём событие атаки.
//   - Тир зависит от суммы очков (score): <1000 → 1, 1000..50000 → 2, >50000 → 3.
//   - Исключаем игроков с любым активным alien-событием за последние 6 дней
//     (FLY_UNKNOWN/HOLDING/ATTACK/HALT) — фильтр по user_id, не по planet.
//     Это автоматически защищает планеты в HOLDING от повторной атаки,
//     даже на других планетах того же игрока.
//   - По четвергам (UTC): ×5 кандидатов (ThursdayCandidateMultiplier).
//     Усиление силы флота идёт отдельно в calcDefPower / scaledAlienFleet
//     через ThursdayPowerMin/Max.
func (s *Service) Spawn(ctx context.Context) error {
	limit := 5
	if time.Now().UTC().Weekday() == time.Thursday {
		limit *= ThursdayCandidateMultiplier
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT u.id, u.score, p.id, p.galaxy, p.system, p.position
		FROM users u
		JOIN planets p ON p.user_id = u.id AND p.destroyed_at IS NULL AND p.is_moon = false
		WHERE u.last_seen_at > now() - interval '7 days'
		  AND u.banned_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM events e
		    JOIN planets p2 ON p2.id = e.planet_id AND p2.user_id = u.id
		    WHERE e.kind IN ($2, $3, $4, $5)
		      AND e.fire_at > now() - interval '6 days'
		  )
		ORDER BY random()
		LIMIT $1
	`, limit,
		int(event.KindAlienFlyUnknown), int(event.KindAlienHolding),
		int(event.KindAlienAttack), int(event.KindAlienHalt))
	if err != nil {
		return fmt.Errorf("alien spawn: query: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		userID   string
		planetID string
		score    int64
		galaxy   int
		system   int
		position int
	}
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.userID, &c.score, &c.planetID, &c.galaxy, &c.system, &c.position); err != nil {
			return fmt.Errorf("alien spawn: scan: %w", err)
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, c := range candidates {
		if rand.IntN(100) >= 30 { // 30% chance
			continue
		}
		tier := scoreTier(c.score)
		dist := alienDistance(alienHomeGalaxy, alienHomeSystem, alienHomePosition, c.galaxy, c.system, c.position)
		flight := alienFlightDuration(dist)
		fireAt := time.Now().Add(flight)
		pl, _ := json.Marshal(alienPayload{
			PlanetID: c.planetID, UserID: c.userID, Tier: tier,
			Galaxy: c.galaxy, System: c.system, Position: c.position,
		})
		if _, err := s.db.Pool().Exec(ctx, `
			INSERT INTO events (id, kind, planet_id, fire_at, payload)
			VALUES ($1, $2, $3, $4, $5)
		`, ids.New(), event.KindAlienAttack, c.planetID, fireAt, pl); err != nil {
			return fmt.Errorf("alien spawn: insert event: %w", err)
		}
	}
	return nil
}

// alienDistance — расстояние от координат инопланетян до планеты игрока.
// Использует те же формулы, что и galaxy.Distance.
func alienDistance(aGal, aSys, aPos, bGal, bSys, bPos int) int {
	switch {
	case aGal != bGal:
		d := aGal - bGal
		if d < 0 {
			d = -d
		}
		return 20000 * d
	case aSys != bSys:
		d := aSys - bSys
		if d < 0 {
			d = -d
		}
		return 2700 + 95*d
	case aPos != bPos:
		d := aPos - bPos
		if d < 0 {
			d = -d
		}
		return 1000 + 5*d
	default:
		return 5
	}
}

// alienFlightDuration — время полёта инопланетян. Та же формула, что в fleet/transport.go.
func alienFlightDuration(distance int) time.Duration {
	raw := 10 + 3500.0*math.Sqrt(10*float64(distance)/float64(alienFleetSpeed))/100.0
	if raw < 60 {
		raw = 60 // минимум 1 минута
	}
	return time.Duration(raw * float64(time.Second))
}

// AttackHandler — event.Handler для KindAlienAttack=35.
func (s *Service) AttackHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl alienPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("alien attack: parse: %w", err)
		}

		// Проверяем, что планета жива.
		var defUserID string
		var defMetal, defSil, defHydro float64
		err := tx.QueryRow(ctx, `
			SELECT user_id, metal, silicon, hydrogen
			FROM planets WHERE id = $1 AND destroyed_at IS NULL FOR UPDATE
		`, pl.PlanetID).Scan(&defUserID, &defMetal, &defSil, &defHydro)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil // планета уничтожена — пропускаем
			}
			return fmt.Errorf("alien attack: find planet: %w", err)
		}

		// Читаем защиту планеты.
		defShips, err := readPlanetShips(ctx, tx, pl.PlanetID)
		if err != nil {
			return fmt.Errorf("alien attack: def ships: %w", err)
		}
		defDefense, err := readPlanetDefense(ctx, tx, pl.PlanetID)
		if err != nil {
			return fmt.Errorf("alien attack: def defense: %w", err)
		}
		defTech, err := readUserTech(ctx, tx, defUserID, s.cat)
		if err != nil {
			return fmt.Errorf("alien attack: def tech: %w", err)
		}

		defUnits := stacksToBattleUnits(defShips, s.cat, false)
		defUnits = append(defUnits, stacksToBattleUnits(defDefense, s.cat, true)...)
		defSide := battle.Side{UserID: defUserID, Tech: defTech, Units: defUnits}

		seed := rng.New(fnvHash(e.ID))
		defPower := calcDefPower(defUnits)
		bonusScale := 1.0
		if time.Now().UTC().Weekday() == time.Thursday {
			// ThursdayPowerMin..Max, детерминированно от seed.
			bonusScale = ThursdayPowerMin + seed.Float64()*(ThursdayPowerMax-ThursdayPowerMin)
		}
		atkSide := battle.Side{
			UserID:   "aliens",
			Username: s.tr("alien", "raceName", nil),
			IsAliens: true,
			Units:    scaledAlienFleet(defPower, seed, s.cat, bonusScale),
		}

		var report battle.Report
		if len(defSide.Units) == 0 {
			// Нет защитников — инопланетяне грабят ресурсы.
			report = battle.Report{Winner: "attackers", Rounds: 0,
				Seed: seed.Uint64()}
		} else {
			inp := battle.Input{
				Seed:      seed.Uint64(),
				Rounds:    6,
				Attackers: []battle.Side{atkSide},
				Defenders: []battle.Side{defSide},
				Rapidfire: rapidfireToMap(s.cat),
			}
			var err error
			report, err = battle.Calculate(inp)
			if err != nil {
				return fmt.Errorf("alien attack: battle: %w", err)
			}
		}

		// Применяем потери защитника.
		if len(report.Defenders) > 0 {
			if err := applyDefenderLosses(ctx, tx, pl.PlanetID,
				defShips, defDefense, report.Defenders[0].Units); err != nil {
				return fmt.Errorf("alien attack: apply defender losses: %w", err)
			}
		}

		// Лут: при победе инопланетян — 30% ресурсов планеты.
		var grabCredit int64
		var haltSpawned bool
		if report.Winner == "attackers" {
			lootM := int64(defMetal * 0.3)
			lootS := int64(defSil * 0.3)
			lootH := int64(defHydro * 0.3)
			if lootM > 0 || lootS > 0 || lootH > 0 {
				if _, err := tx.Exec(ctx,
					`UPDATE planets SET metal = metal - $1,
					 silicon = silicon - $2, hydrogen = hydrogen - $3
					 WHERE id = $4`,
					lootM, lootS, lootH, pl.PlanetID); err != nil {
					return fmt.Errorf("alien attack: loot: %w", err)
				}
			}
			// GRAB_CREDIT: 0.08–0.1% от кредитов при победе (если >100000).
			grabCredit, err = applyGrabCredit(ctx, tx, defUserID, rng.New(fnvHash(e.ID)^0xcafebabe))
			if err != nil {
				return fmt.Errorf("alien attack: grab credit: %w", err)
			}
			// HALT: планета переходит в удержание пришельцами, если
			// их флот выжил. Используем rand с seed от event.ID для
			// детерминированности.
			if len(report.Attackers) > 0 {
				survivors := survivorsToStacks(report.Attackers[0].Units)
				r := rand.New(rand.NewPCG(fnvHash(e.ID)^0xa11e50, fnvHash(e.ID)^0xa11e51))
				if err := s.spawnHalt(ctx, tx, pl, survivors, r); err != nil {
					return fmt.Errorf("alien attack: spawn halt: %w", err)
				}
				haltSpawned = len(survivors) > 0
			}
		}

		// Artefact drop: 20% шанс при победе защитника.
		var giftCredit int64
		if report.Winner == "defenders" {
			artSeed := rng.New(fnvHash(e.ID) ^ 0xdeadbeef)
			if artSeed.IntN(100) < 20 {
				if err := grantRandomArtefact(ctx, tx, defUserID, s.cat); err != nil {
					return fmt.Errorf("alien attack: artefact drop: %w", err)
				}
			}
			// GIFT_CREDIT: 5–10% кредитов (max 500) при отражении атаки.
			giftCredit, err = applyGiftCredit(ctx, tx, defUserID, rng.New(fnvHash(e.ID)^0xbeefdead))
			if err != nil {
				return fmt.Errorf("alien attack: gift credit: %w", err)
			}
		}

		// Сообщение защитнику.
		var resultKey string
		switch report.Winner {
		case "defenders":
			resultKey = "attack.defeat"
		case "draw":
			resultKey = "attack.draw"
		default:
			resultKey = "attack.victory"
		}
		tierStr := fmt.Sprintf("%d", pl.Tier)
		body := s.tr("alien", "attack.body", map[string]string{
			"tier":   tierStr,
			"result": s.tr("alien", resultKey, nil),
		})
		if grabCredit > 0 {
			body += s.tr("alien", "attack.creditStolen", map[string]string{
				"credits": fmt.Sprintf("%d", grabCredit),
			})
		}
		if giftCredit > 0 {
			body += s.tr("alien", "attack.creditGifted", map[string]string{
				"credits": fmt.Sprintf("%d", giftCredit),
			})
		}
		if haltSpawned {
			body += s.tr("alien", "attack.haltNote", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 1, $3, $4)
		`, ids.New(), defUserID, s.tr("alien", "attack.title", nil), body); err != nil {
			return fmt.Errorf("alien attack: message: %w", err)
		}

		return nil
	}
}

// applyGrabCredit снимает 0.08–0.1% кредитов (если >100000).
// Возвращает реально снятую сумму (0 если условие не выполнено).
func applyGrabCredit(ctx context.Context, tx pgx.Tx, userID string, r *rng.R) (int64, error) {
	var credit int64
	if err := tx.QueryRow(ctx, `SELECT credit FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&credit); err != nil {
		return 0, err
	}
	const minCredit = 100_000
	if credit <= minCredit {
		return 0, nil
	}
	pct := 0.0008 + r.Float64()*0.0002 // 0.08%–0.10%
	grab := int64(math.Round(float64(credit) * pct))
	if grab <= 0 {
		return 0, nil
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET credit = credit - $1 WHERE id = $2`, grab, userID); err != nil {
		return 0, err
	}
	return grab, nil
}

// applyGiftCredit добавляет 5–10% кредитов (max 500) при отражении атаки.
func applyGiftCredit(ctx context.Context, tx pgx.Tx, userID string, r *rng.R) (int64, error) {
	var credit int64
	if err := tx.QueryRow(ctx, `SELECT credit FROM users WHERE id=$1`, userID).Scan(&credit); err != nil {
		return 0, err
	}
	pct := 0.05 + r.Float64()*0.05 // 5%–10%
	gift := int64(math.Round(float64(credit) * pct))
	const maxGift = 500
	if gift > maxGift {
		gift = maxGift
	}
	if gift <= 0 {
		return 0, nil
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET credit = credit + $1 WHERE id = $2`, gift, userID); err != nil {
		return 0, err
	}
	return gift, nil
}

// grantRandomArtefact вставляет случайный артефакт из каталога игроку.
func grantRandomArtefact(ctx context.Context, tx pgx.Tx, userID string, cat *config.Catalog) error {
	if len(cat.Artefacts.Artefacts) == 0 {
		return nil
	}
	artIDs := make([]int, 0, len(cat.Artefacts.Artefacts))
	for _, spec := range cat.Artefacts.Artefacts {
		artIDs = append(artIDs, spec.ID)
	}
	artID := artIDs[rand.IntN(len(artIDs))]
	_, err := tx.Exec(ctx, `
		INSERT INTO artefacts_user (id, user_id, planet_id, unit_id, state, acquired_at)
		VALUES ($1, $2, NULL, $3, 'held', now())
	`, ids.New(), userID, artID)
	return err
}

func scoreTier(score int64) int {
	switch {
	case score >= 50000:
		return 3
	case score >= 1000:
		return 2
	default:
		return 1
	}
}

// alienFleet — флот инопланетян по тиру. Характеристики взяты из
// oxsar2-java (unit_id 200-204 не в каталоге — задаём напрямую).
func alienFleet(tier int) []battle.Unit {
	switch tier {
	case 3:
		return []battle.Unit{
			{UnitID: 202, Quantity: 50, Front: 8,
				Attack: 2000, Shell: 25000, Name: "Alien Cruiser"},
			{UnitID: 203, Quantity: 10, Front: 9,
				Attack: 8000, Shell: 80000, Name: "Alien Battleship"},
		}
	case 2:
		return []battle.Unit{
			{UnitID: 201, Quantity: 30, Front: 7,
				Attack: 800, Shell: 10000, Name: "Alien Destroyer"},
			{UnitID: 202, Quantity: 10, Front: 8,
				Attack: 2000, Shell: 25000, Name: "Alien Cruiser"},
		}
	default: // tier 1
		return []battle.Unit{
			{UnitID: 200, Quantity: 20, Front: 5,
				Attack: 150, Shell: 2000, Name: "Alien Scout"},
			{UnitID: 201, Quantity: 5, Front: 6,
				Attack: 800, Shell: 10000, Name: "Alien Destroyer"},
		}
	}
}
