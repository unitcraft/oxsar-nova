package planet

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
	"oxsar/game-nova/pkg/rng"
)

// StartingResources — ресурсы на старте (Dominator params.php).
var StartingResources = struct {
	Metal    int64
	Silicon  int64
	Hydrogen int64
}{
	Metal:    1000,
	Silicon:  500,
	Hydrogen: 0,
}

// HomePlanetSize — число полей домашней планеты (HOME_PLANET_SIZE из legacy).
const HomePlanetSize = 18800

// starterBuildings — начальные уровни зданий (INITIAL_BUILDINGS, Dominator params.php).
// unit_id из configs/buildings.yml.
var starterBuildings = []struct {
	unitID int
	level  int
}{
	{1, 2},   // metal_mine
	{2, 2},   // silicon_lab
	{3, 2},   // hydrogen_lab
	{4, 4},   // solar_plant
	{6, 2},   // robotic_factory
	{8, 2},   // shipyard
	{12, 2},  // research_lab
	{101, 1}, // defense_factory
	{100, 1}, // repair_factory
}

// starterResearch — начальные уровни исследований (INITIAL_RESEARCHES).
var starterResearch = []struct {
	unitID int
	level  int
}{
	{14, 1}, // computer_tech
	{18, 1}, // energy_tech
	{20, 2}, // combustion_engine
}

// starterFleet — начальный флот (INITIAL_UNITS).
// unit_id из configs/ships.yml.
var starterFleet = []struct {
	unitID int
	count  int
}{
	{30, 20}, // small_transporter
	{31, 10}, // light_fighter
	{35, 10}, // recycler
	{36, 3},  // colony_ship
	{37, 10}, // espionage_probe
}

// Starter создаёт первую планету игрока сразу после регистрации
// (§2.2 ТЗ: «первая планета», §5.13 защита новичков).
//
// Выделено в отдельную службу, чтобы auth-пакет не зависел от всей
// планетарной логики, а только от одной функции.
type Starter struct {
	db          repo.Exec
	numGalaxies int // план 72.1 ч.12 — лимиты вселенной из universes.yaml
	numSystems  int
}

// NewStarter — план 72.1 ч.12: numGalaxies/numSystems задают диапазон
// случайной генерации стартовой планеты (раньше был hardcoded 1..8 / 1..500).
func NewStarter(db repo.Exec, numGalaxies, numSystems int) *Starter {
	return &Starter{db: db, numGalaxies: numGalaxies, numSystems: numSystems}
}

// Assign создаёт планету на случайной свободной позиции и делает её
// текущей для пользователя. Возвращает id созданной планеты.
//
// Алгоритм: крутим координаты, пока не найдём свободную. Для каждой
// вселенной диапазон 1..numGalaxies × 1..numSystems × 1..15. На случай
// полной вселенной — возвращаем ошибку после 100 попыток.
func (s *Starter) Assign(ctx context.Context, userID string) (string, error) {
	r := rng.New(seedFromUserID(userID))
	var planetID string

	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		for attempt := 0; attempt < 100; attempt++ {
			g := r.IntN(s.numGalaxies) + 1
			sys := r.IntN(s.numSystems) + 1
			pos := r.IntN(13) + 2 // 2..14 (1 и 15 — «крайности», оставим на колонизацию)

			taken, err := coordTaken(ctx, tx, g, sys, pos, false)
			if err != nil {
				return err
			}
			if taken {
				continue
			}

			id := ids.New()
			rCoord := rng.New(starterCoordsSeed(g, sys, pos))
			pType := starterPlanetTypeOf(pos, rCoord)
			tempMin, tempMax := starterPositionTemp(pos, rCoord)

			// План 23: used_fields = число стартовых зданий (они занимают
			// поля с момента создания планеты).
			_, err = tx.Exec(ctx, `
				INSERT INTO planets (id, user_id, is_moon, name, galaxy, system, position,
				                     diameter, used_fields, planet_type, temperature_min, temperature_max,
				                     metal, silicon, hydrogen)
				VALUES ($1, $2, false, $3, $4, $5, $6, $7, $14, $8, $9, $10, $11, $12, $13)
			`, id, userID, "Homeworld", g, sys, pos, HomePlanetSize, pType, tempMin, tempMax,
				StartingResources.Metal, StartingResources.Silicon, StartingResources.Hydrogen,
				len(starterBuildings))
			if err != nil {
				return fmt.Errorf("insert starter planet: %w", err)
			}

			if _, err := tx.Exec(ctx,
				`UPDATE users SET cur_planet_id = $1 WHERE id = $2`, id, userID); err != nil {
				return fmt.Errorf("set cur_planet: %w", err)
			}

			// Начальные здания (INITIAL_BUILDINGS из legacy params.php).
			for _, b := range starterBuildings {
				if _, err := tx.Exec(ctx, `
					INSERT INTO buildings (planet_id, unit_id, level)
					VALUES ($1, $2, $3)
					ON CONFLICT (planet_id, unit_id) DO UPDATE SET level = EXCLUDED.level
				`, id, b.unitID, b.level); err != nil {
					return fmt.Errorf("starter building %d: %w", b.unitID, err)
				}
			}

			// Начальные исследования (INITIAL_RESEARCHES).
			for _, res := range starterResearch {
				if _, err := tx.Exec(ctx, `
					INSERT INTO research (user_id, unit_id, level)
					VALUES ($1, $2, $3)
					ON CONFLICT (user_id, unit_id) DO UPDATE SET level = EXCLUDED.level
				`, userID, res.unitID, res.level); err != nil {
					return fmt.Errorf("starter research %d: %w", res.unitID, err)
				}
			}

			// Начальный флот (INITIAL_UNITS).
			for _, f := range starterFleet {
				if _, err := tx.Exec(ctx, `
					INSERT INTO ships (planet_id, unit_id, count)
					VALUES ($1, $2, $3)
					ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = EXCLUDED.count
				`, id, f.unitID, f.count); err != nil {
					return fmt.Errorf("starter fleet %d: %w", f.unitID, err)
				}
			}

			// Базовый journal entry — чтобы первая запись в res_log была «стартовый грант».
			if _, err := tx.Exec(ctx, `
				INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
				VALUES ($1, $2, 'admin_gift', $3, $4, $5)
			`, userID, id,
				StartingResources.Metal, StartingResources.Silicon, StartingResources.Hydrogen,
			); err != nil {
				return fmt.Errorf("starter res_log: %w", err)
			}

			planetID = id
			return nil
		}
		return errors.New("planet: no free coords after 100 attempts")
	})
	if err != nil {
		return "", err
	}
	return planetID, nil
}

func coordTaken(ctx context.Context, tx pgx.Tx, g, sys, pos int, isMoon bool) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM planets
			WHERE galaxy = $1 AND system = $2 AND position = $3 AND is_moon = $4
			  AND destroyed_at IS NULL
		)
	`, g, sys, pos, isMoon).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("coord check: %w", err)
	}
	return exists, nil
}

// seedFromUserID — детерминированный seed для генератора стартовых
// координат. Разные пользователи получают разные последовательности,
// но один и тот же пользователь при retry (повторный вызов Assign) —
// ту же. Это не критично, просто меньше случайности в тестах.
func starterCoordsSeed(g, sys, pos int) uint64 {
	var h uint64 = 14695981039346656037
	for _, v := range []int{g, sys, pos} {
		for i := 0; i < 4; i++ {
			h ^= uint64(byte(v >> (i * 8)))
			h *= 1099511628211
		}
	}
	return h
}

func starterPlanetTypeOf(pos int, r *rng.R) string {
	switch {
	case pos <= 2:
		return "trockenplanet"
	case pos == 3:
		if r.IntN(2) == 0 {
			return "trockenplanet"
		}
		return "wuestenplanet"
	case pos <= 5:
		return "dschjungelplanet"
	case pos <= 7:
		if r.IntN(2) == 0 {
			return "dschjungelplanet"
		}
		return "normaltempplanet"
	case pos <= 9:
		return "normaltempplanet"
	case pos <= 11:
		if r.IntN(2) == 0 {
			return "normaltempplanet"
		}
		return "wasserplanet"
	case pos <= 13:
		return "wasserplanet"
	case pos == 14:
		return "eisplanet"
	default:
		if r.IntN(2) == 0 {
			return "eisplanet"
		}
		return "gasplanet"
	}
}

func starterPositionTemp(pos int, r *rng.R) (tempMin, tempMax int) {
	base := 110 - pos*14
	spread := r.IntN(20) - 10
	tempMax = base + spread
	tempMin = tempMax - 40
	return
}

func seedFromUserID(userID string) uint64 {
	// FNV-1a, чтобы не тянуть hash/fnv в этот файл ради одной строки.
	var h uint64 = 14695981039346656037
	for i := 0; i < len(userID); i++ {
		h ^= uint64(userID[i])
		h *= 1099511628211
	}
	return h
}
