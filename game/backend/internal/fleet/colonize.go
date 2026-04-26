// COLONIZE (mission=8) — основание новой колонии.
//
// Поток прибытия:
//  1. Читаем fleet + fleet_ships (ищем colony_ship id=36).
//  2. Проверяем, что координата пустая (нет planets там).
//  3. Проверяем лимит планет: count(user) <= computer_tech + 1.
//  4. Создаём planets row с ресурсами = carry, diameter/temp по
//     seed от координат (детерминированно).
//  5. Снимаем 1 colony_ship из fleet_ships (DECREMENT или DELETE).
//  6. Обнуляем carry флота, state='returning' — остальные ships
//     вернутся на src_planet, новая планета уже заселена.
//  7. Message игроку: «Колония основана в G:S:P».
//
// Ограничения M5.COLONIZE:
//  * позиция 1 и 15 тоже доступны (упрощение; legacy их запрещает
//    для старта).
//  * размер planet — стандарт (12800..14800), без учёта позиции
//    (в legacy ближе к звезде = меньше).
//  * нет выбора имени — дефолт «Colony».
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/pkg/ids"
	"github.com/oxsar/nova/backend/pkg/rng"
)

// unitColonyShip — id колониального корабля (legacy UNIT_COLONY_SHIP).
const unitColonyShip = 36

// unitComputerTech — id computer-технологии в research (legacy UNIT_COMPUTER_TECH).
const unitComputerTech = 14

// unitAstroTech — Astrophysics (id=112). План 20 Ф.7 + ADR-0005.
// Расширяет лимит колоний и даёт слоты экспедиций.
const unitAstroTech = 112

// ColonizeHandler — event.Handler для KindColonize=8.
func (s *TransportService) ColonizeHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl transportPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("colonize: parse payload: %w", err)
		}
		var (
			state                    string
			ownerUserID              string
			g, sys, pos              int
			isMoon                   bool
			cm, csil, ch             int64
		)
		err := tx.QueryRow(ctx, `
			SELECT state, owner_user_id,
			       dst_galaxy, dst_system, dst_position, dst_is_moon,
			       carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id = $1 FOR UPDATE
		`, pl.FleetID).Scan(&state, &ownerUserID, &g, &sys, &pos, &isMoon, &cm, &csil, &ch)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("colonize: read fleet: %w", err)
		}
		if state != "outbound" {
			return nil
		}
		coords := fmt.Sprintf("%d:%d:%d", g, sys, pos)
		failedSubj := s.tr("colonize", "failed.title", map[string]string{"coords": coords})

		if isMoon {
			return abortReturning(ctx, tx, pl.FleetID, ownerUserID,
				failedSubj, s.tr("colonize", "error.moonOnly", nil))
		}

		// colony_ship есть?
		var colonyCount int64
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(count, 0) FROM fleet_ships WHERE fleet_id=$1 AND unit_id=$2`,
			pl.FleetID, unitColonyShip).Scan(&colonyCount); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("colonize: read colony ship: %w", err)
		}
		if colonyCount <= 0 {
			return abortReturning(ctx, tx, pl.FleetID, ownerUserID,
				failedSubj, s.tr("colonize", "error.noColonyShip", nil))
		}

		// Координата пуста?
		var existingID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM planets
			WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=false
			  AND destroyed_at IS NULL
		`, g, sys, pos).Scan(&existingID)
		if err == nil {
			return abortReturning(ctx, tx, pl.FleetID, ownerUserID,
				failedSubj, s.tr("colonize", "error.occupied", nil))
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("colonize: check empty: %w", err)
		}

		// Лимит планет: max(computer_tech+1, astro_level/2+1).
		// План 20 Ф.7 + ADR-0005: astro_tech даёт дополнительные
		// колонии. Берём максимум, чтобы не отнять у игроков с
		// прокаченным computer_tech их текущие слоты.
		computerLvl := readComputerLevel(ctx, tx, ownerUserID)
		astroLvl := readResearchLevel(ctx, tx, ownerUserID, unitAstroTech)
		maxPlanets := computerLvl + 1
		if astroLimit := astroLvl/2 + 1; astroLimit > maxPlanets {
			maxPlanets = astroLimit
		}
		if s.maxPlanets > 0 && s.maxPlanets > maxPlanets {
			maxPlanets = s.maxPlanets
		}
		var curPlanets int
		if err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM planets WHERE user_id=$1 AND destroyed_at IS NULL AND is_moon=false`,
			ownerUserID).Scan(&curPlanets); err != nil {
			return fmt.Errorf("colonize: count planets: %w", err)
		}
		if curPlanets >= maxPlanets {
			return abortReturning(ctx, tx, pl.FleetID, ownerUserID,
				failedSubj, s.tr("colonize", "error.planetLimit", map[string]string{
					"current": strconv.Itoa(curPlanets),
					"max":     strconv.Itoa(maxPlanets),
				}))
		}

		// Создаём планету. diameter/temp по детерминированному seed от
		// (galaxy, system, position) — одинаковая координата всегда
		// даёт одинаковые параметры. Размер зависит от позиции: позиции
		// 1-3 и 13-15 — меньше; 4-12 — стандарт (как в legacy OGame).
		r := rng.New(coordsSeed(g, sys, pos))
		diameter := positionDiameter(pos, r)
		pType := planetTypeOf(pos, r)
		tempMin, tempMax := positionTemp(pos, r)

		// Имя планеты: из payload или «Colony» по умолчанию.
		colonyName := pl.ColonyName
		if colonyName == "" {
			colonyName = "Colony"
		}

		newPlanetID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO planets (id, user_id, is_moon, name, galaxy, system, position,
			                     diameter, used_fields, planet_type, temperature_min, temperature_max,
			                     metal, silicon, hydrogen)
			VALUES ($1, $2, false, $3, $4, $5, $6, $7, 0, $8, $9, $10, $11, $12, $13)
		`, newPlanetID, ownerUserID, colonyName, g, sys, pos, diameter, pType, tempMin, tempMax,
			cm, csil, ch); err != nil {
			return fmt.Errorf("colonize: insert planet: %w", err)
		}

		// Снимаем 1 colony_ship из флота (для возврата он больше
		// не нужен — корабль «разобран» для основания базы).
		if colonyCount == 1 {
			if _, err := tx.Exec(ctx,
				`DELETE FROM fleet_ships WHERE fleet_id=$1 AND unit_id=$2`,
				pl.FleetID, unitColonyShip); err != nil {
				return fmt.Errorf("colonize: delete colony ship: %w", err)
			}
		} else {
			if _, err := tx.Exec(ctx, `
				UPDATE fleet_ships SET count = count - 1
				WHERE fleet_id=$1 AND unit_id=$2
			`, pl.FleetID, unitColonyShip); err != nil {
				return fmt.Errorf("colonize: decrement colony ship: %w", err)
			}
		}

		// Carry обнуляется (уже на новой планете).
		if _, err := tx.Exec(ctx, `
			UPDATE fleets SET state='returning',
			                  carried_metal=0, carried_silicon=0, carried_hydrogen=0
			WHERE id=$1
		`, pl.FleetID); err != nil {
			return fmt.Errorf("colonize: update fleet: %w", err)
		}

		// Сообщение игроку.
		subj := s.tr("colonize", "success.title", map[string]string{"coords": coords})
		body := s.tr("colonize", "success.body", map[string]string{
			"planetName": colonyName,
			"coords":     coords,
			"metal":      strconv.FormatInt(cm, 10),
			"silicon":    strconv.FormatInt(csil, 10),
			"hydrogen":   strconv.FormatInt(ch, 10),
		})
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 2, $3, $4)
		`, ids.New(), ownerUserID, subj, body); err != nil {
			return fmt.Errorf("colonize: insert message: %w", err)
		}
		return nil
	}
}

// abortReturning — причина неудачи → message + state='returning' с
// сохранением carry (флот возвращает груз).
func abortReturning(ctx context.Context, tx pgx.Tx, fleetID, userID, subj, reason string) error {
	if _, err := tx.Exec(ctx,
		`UPDATE fleets SET state='returning' WHERE id=$1`, fleetID); err != nil {
		return fmt.Errorf("abort: update fleet: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, 2, $3, $4)
	`, ids.New(), userID, subj, reason); err != nil {
		return fmt.Errorf("abort: insert message: %w", err)
	}
	return nil
}

func readComputerLevel(ctx context.Context, tx pgx.Tx, userID string) int {
	if userID == "" {
		return 0
	}
	var lvl int
	err := tx.QueryRow(ctx,
		`SELECT level FROM research WHERE user_id=$1 AND unit_id=$2`,
		userID, unitComputerTech).Scan(&lvl)
	if err != nil {
		return 0
	}
	return lvl
}

// positionDiameter — диаметр планеты по позиции (как в legacy OGame).
// Позиции 1-3: малые (6000-10000), 4-12: стандарт (10000-15000), 13-15: большие (12000-17000).
func positionDiameter(pos int, r *rng.R) int {
	switch {
	case pos <= 3:
		return 6000 + r.IntN(4000)
	case pos >= 13:
		return 12000 + r.IntN(5000)
	default:
		return 10000 + r.IntN(5000)
	}
}

// coordsSeed — детерминированный seed от (g, sys, pos). FNV-1a от
// составного ключа, чтобы одинаковая координата у любого игрока
// давала одинаковые diameter/temp.
func coordsSeed(g, sys, pos int) uint64 {
	var h uint64 = 14695981039346656037
	for _, v := range []int{g, sys, pos} {
		for i := 0; i < 4; i++ {
			h ^= uint64(byte(v >> (i * 8)))
			h *= 1099511628211
		}
	}
	return h
}

// planetTypeOf — детерминированный тип планеты по позиции и rng.
// Логика по OGame: слоты 1-3 горячие (сухие/пустынные), 13-15 холодные
// (ледяные/газовые), середина — умеренные и водные биомы.
// RNG используется когда на позиции возможно несколько типов.
func planetTypeOf(pos int, r *rng.R) string {
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

// positionTemp — температура по позиции в системе.
// Слот 1 (ближе к звезде) ≈ +100°C, слот 15 ≈ -100°C, ±10° случайный разброс.
func positionTemp(pos int, r *rng.R) (tempMin, tempMax int) {
	base := 110 - pos*14        // слот 1 → 96, слот 15 → -100
	spread := r.IntN(20) - 10   // -10..+9
	tempMax = base + spread
	tempMin = tempMax - 40
	return
}
