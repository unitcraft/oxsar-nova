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

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
	"github.com/oxsar/nova/backend/pkg/rng"
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
	db  repo.Exec
	cat *config.Catalog
}

func NewService(db repo.Exec, cat *config.Catalog) *Service {
	return &Service{db: db, cat: cat}
}

// Spawn выбирает случайных активных игроков и создаёт события
// KindAlienAttack. Вызывается из воркера раз в N часов.
//
// Логика выбора (упрощена vs оригинала):
//   - Берём до 5 случайных активных игроков (last_seen_at < 7 дней).
//   - С вероятностью 30% для каждого создаём событие атаки.
//   - Тир зависит от суммы очков (score): <1000 → 1, 1000..50000 → 2, >50000 → 3.
func (s *Service) Spawn(ctx context.Context) error {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT u.id, u.score, p.id, p.galaxy, p.system, p.position
		FROM users u
		JOIN planets p ON p.user_id = u.id AND p.destroyed_at IS NULL AND p.is_moon = false
		WHERE u.last_seen_at > now() - interval '7 days'
		  AND u.banned_at IS NULL
		ORDER BY random()
		LIMIT 5
	`)
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
		defTech, err := readUserTech(ctx, tx, defUserID)
		if err != nil {
			return fmt.Errorf("alien attack: def tech: %w", err)
		}

		defUnits := stacksToBattleUnits(defShips, s.cat, false)
		defUnits = append(defUnits, stacksToBattleUnits(defDefense, s.cat, true)...)
		defSide := battle.Side{UserID: defUserID, Tech: defTech, Units: defUnits}

		atkSide := battle.Side{
			UserID:   "aliens",
			Username: "Инопланетяне",
			IsAliens: true,
			Units:    alienFleet(pl.Tier),
		}

		seed := rng.New(fnvHash(e.ID))

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
		}

		// Artefact drop: 20% шанс при победе защитника.
		if report.Winner == "defenders" {
			artSeed := rng.New(fnvHash(e.ID) ^ 0xdeadbeef)
			if artSeed.IntN(100) < 20 {
				if err := grantRandomArtefact(ctx, tx, defUserID, s.cat); err != nil {
					return fmt.Errorf("alien attack: artefact drop: %w", err)
				}
			}
		}

		// Сообщение защитнику.
		result := "атаковали и победили"
		if report.Winner == "defenders" {
			result = "атаковали, но были отбиты"
		} else if report.Winner == "draw" {
			result = "атаковали — ничья"
		}
		body := fmt.Sprintf("Инопланетяне (тир %d) %s вашу планету.", pl.Tier, result)
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 1, 'Атака инопланетян', $3)
		`, ids.New(), defUserID, body); err != nil {
			return fmt.Errorf("alien attack: message: %w", err)
		}

		return nil
	}
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
				Attack: [3]float64{2000, 0, 0}, Shell: 25000, Name: "Alien Cruiser"},
			{UnitID: 203, Quantity: 10, Front: 9,
				Attack: [3]float64{8000, 0, 0}, Shell: 80000, Name: "Alien Battleship"},
		}
	case 2:
		return []battle.Unit{
			{UnitID: 201, Quantity: 30, Front: 7,
				Attack: [3]float64{800, 0, 0}, Shell: 10000, Name: "Alien Destroyer"},
			{UnitID: 202, Quantity: 10, Front: 8,
				Attack: [3]float64{2000, 0, 0}, Shell: 25000, Name: "Alien Cruiser"},
		}
	default: // tier 1
		return []battle.Unit{
			{UnitID: 200, Quantity: 20, Front: 5,
				Attack: [3]float64{150, 0, 0}, Shell: 2000, Name: "Alien Scout"},
			{UnitID: 201, Quantity: 5, Front: 6,
				Attack: [3]float64{800, 0, 0}, Shell: 10000, Name: "Alien Destroyer"},
		}
	}
}
