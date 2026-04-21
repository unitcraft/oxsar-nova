// EXPEDITION (mission=15) — полёт в неисследованную зону.
//
// При прибытии флот выполняет 1 из 6 исходов, выбранных по seed от
// fleetID (детерминированно по uuid):
//
//   resources    (30%): +N ресурсов в carry (до cargo cap).
//   artefact      (5%): случайный артефакт в artefacts_user state=held.
//   extra_planet  (5%): новая планета создаётся на случайном свободном слоте.
//   pirates       (20%): battle против PvE-флота (5 light_fighter).
//   loss          (15%): 5-20% ship'ов теряются.
//   nothing       (25%): пустой отчёт, возврат без изменений.
//
// Поток:
//   1. Читаем fleet + fleet_ships.
//   2. RNG от fleetID → outcome (roll 0..99).
//   3. Применяем effect (мутируем БД).
//   4. INSERT expedition_reports + message с expedition_report_id.
//   5. state='returning'.
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/pkg/ids"
	"github.com/oxsar/nova/backend/pkg/rng"
)

// ExpeditionHandler — event.Handler для KindExpedition=15.
func (s *TransportService) ExpeditionHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl transportPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("expedition: parse payload: %w", err)
		}
		var (
			state        string
			ownerUserID  string
			cm, csil, ch int64
		)
		err := tx.QueryRow(ctx, `
			SELECT state, owner_user_id, carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id = $1 FOR UPDATE
		`, pl.FleetID).Scan(&state, &ownerUserID, &cm, &csil, &ch)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("expedition: read fleet: %w", err)
		}
		if state != "outbound" {
			return nil
		}

		r := rng.New(deriveSeed(pl.FleetID))
		roll := r.IntN(100)

		fleetShips, err := readFleetShips(ctx, tx, pl.FleetID)
		if err != nil {
			return fmt.Errorf("expedition: read fleet_ships: %w", err)
		}

		var (
			outcome    string
			reportData map[string]any
		)
		switch {
		case roll < 30:
			outcome = "resources"
			reportData = expResources(ctx, tx, r, pl.FleetID, fleetShips, s.catalog,
				cm, csil, ch)
		case roll < 35:
			outcome = "artefact"
			reportData = expArtefact(ctx, tx, r, ownerUserID, s.catalog)
		case roll < 40:
			outcome = "extra_planet"
			reportData, err = expExtraPlanet(ctx, tx, r, ownerUserID)
			if err != nil {
				return err
			}
		case roll < 60:
			outcome = "pirates"
			reportData, err = expPirates(ctx, tx, pl.FleetID, fleetShips, s.catalog)
			if err != nil {
				return err
			}
		case roll < 75:
			outcome = "loss"
			reportData, err = expLoss(ctx, tx, r, pl.FleetID, fleetShips)
			if err != nil {
				return err
			}
		default:
			outcome = "nothing"
			reportData = map[string]any{"message": "Ничего не нашли."}
		}

		reportJSON, _ := json.Marshal(reportData)
		reportID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO expedition_reports (id, user_id, fleet_id, outcome, report)
			VALUES ($1, $2, $3, $4, $5)
		`, reportID, ownerUserID, pl.FleetID, outcome, reportJSON); err != nil {
			return fmt.Errorf("expedition: insert report: %w", err)
		}

		subj := fmt.Sprintf("Экспедиция: %s", outcome)
		body := fmt.Sprintf("Результат: %s", outcome)
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, expedition_report_id)
			VALUES ($1, $2, NULL, 2, $3, $4, $5)
		`, ids.New(), ownerUserID, subj, body, reportID); err != nil {
			return fmt.Errorf("expedition: insert message: %w", err)
		}

		if _, err := tx.Exec(ctx,
			`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID); err != nil {
			return fmt.Errorf("expedition: update state: %w", err)
		}
		return nil
	}
}

// expExtraPlanet — создаёт новую пустую планету на случайном свободном
// слоте. Проверяет лимит (computer_tech + 1). Если слот не найден
// за 50 попыток или лимит достигнут — возвращает reportData с причиной
// без ошибки (исход всё равно "extra_planet", просто без планеты).
func expExtraPlanet(ctx context.Context, tx pgx.Tx, r *rng.R, userID string) (map[string]any, error) {
	// Лимит планет.
	computerLvl := readComputerLevel(ctx, tx, userID)
	maxPlanets := computerLvl + 1
	var curPlanets int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM planets WHERE user_id=$1 AND destroyed_at IS NULL AND is_moon=false`,
		userID).Scan(&curPlanets); err != nil {
		return nil, fmt.Errorf("expExtraPlanet: count: %w", err)
	}
	if curPlanets >= maxPlanets {
		return map[string]any{
			"message": fmt.Sprintf("Обнаружена пригодная планета, но достигнут лимит (%d/%d). Улучшите computer_tech.", curPlanets, maxPlanets),
		}, nil
	}

	// Ищем свободный слот.
	for attempt := 0; attempt < 50; attempt++ {
		g := r.IntN(8) + 1
		sys := r.IntN(500) + 1
		pos := r.IntN(13) + 2 // 2..14

		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM planets
				WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=false AND destroyed_at IS NULL
			)
		`, g, sys, pos).Scan(&exists); err != nil {
			return nil, fmt.Errorf("expExtraPlanet: check slot: %w", err)
		}
		if exists {
			continue
		}

		rCoord := rng.New(coordsSeed(g, sys, pos))
		diameter := 12800 + rCoord.IntN(2000)
		tempMax := -40 + rCoord.IntN(80)
		tempMin := tempMax - 40

		newID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO planets (id, user_id, is_moon, name, galaxy, system, position,
			                     diameter, used_fields, temperature_min, temperature_max,
			                     metal, silicon, hydrogen)
			VALUES ($1, $2, false, 'Expedition Colony', $3, $4, $5, $6, 0, $7, $8, 0, 0, 0)
		`, newID, userID, g, sys, pos, diameter, tempMin, tempMax); err != nil {
			return nil, fmt.Errorf("expExtraPlanet: insert: %w", err)
		}
		return map[string]any{
			"planet_id": newID,
			"galaxy":    g,
			"system":    sys,
			"position":  pos,
		}, nil
	}
	return map[string]any{"message": "Пригодная планета найдена, но подходящая позиция не обнаружена."}, nil
}

// expResources — бонус ресурсов в carry, ограниченный свободным cargo.
func expResources(ctx context.Context, tx pgx.Tx, r *rng.R, fleetID string,
	ships []unitStack, cat *config.Catalog, cm, csil, ch int64) map[string]any {
	var totalCap int64
	for _, s := range ships {
		for _, spec := range cat.Ships.Ships {
			if spec.ID == s.UnitID {
				totalCap += spec.Cargo * s.Count
				break
			}
		}
	}
	free := totalCap - (cm + csil + ch)
	if free <= 0 {
		return map[string]any{"bonus": "нет места в cargo"}
	}
	fraction := 0.2 + r.Float64()*0.4
	pool := int64(float64(free) * fraction)
	bonusM := pool * 60 / 100
	bonusS := pool * 30 / 100
	bonusH := pool - bonusM - bonusS
	if _, err := tx.Exec(ctx, `
		UPDATE fleets SET carried_metal=$1, carried_silicon=$2, carried_hydrogen=$3
		WHERE id=$4
	`, cm+bonusM, csil+bonusS, ch+bonusH, fleetID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{
		"metal":    bonusM,
		"silicon":  bonusS,
		"hydrogen": bonusH,
	}
}

// expArtefact — вставить случайный артефакт из каталога в state=held.
func expArtefact(ctx context.Context, tx pgx.Tx, r *rng.R, userID string,
	cat *config.Catalog) map[string]any {
	if len(cat.Artefacts.Artefacts) == 0 {
		return map[string]any{"message": "артефакты недоступны"}
	}
	artIDs := make([]int, 0, len(cat.Artefacts.Artefacts))
	for _, spec := range cat.Artefacts.Artefacts {
		artIDs = append(artIDs, spec.ID)
	}
	idx := r.IntN(len(artIDs))
	artID := artIDs[idx]
	if _, err := tx.Exec(ctx, `
		INSERT INTO artefacts_user (id, user_id, planet_id, unit_id, state, acquired_at)
		VALUES ($1, $2, NULL, $3, 'held', now())
	`, ids.New(), userID, artID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"artefact_id": artID}
}

// expPirates — PvE-битва с 5 light_fighter. Потери атакующего
// применяются через существующий applyAttackerLosses (fleet_ships).
func expPirates(ctx context.Context, tx pgx.Tx, fleetID string,
	ships []unitStack, cat *config.Catalog) (map[string]any, error) {
	atkUnits := stacksToBattleUnits(ships, cat, false)
	if len(atkUnits) == 0 {
		return map[string]any{"message": "нет ship'ов"}, nil
	}
	pirateShell, pirateAttack, pirateShield := 4000, 50, 10
	for _, spec := range cat.Ships.Ships {
		if spec.ID == 31 {
			pirateShell, pirateAttack, pirateShield = spec.Shell, spec.Attack, spec.Shield
			break
		}
	}
	pirateSide := battle.Side{
		UserID: "pirates",
		Units: []battle.Unit{{
			UnitID:   31,
			Quantity: 5,
			Front:    0,
			Attack:   [3]float64{float64(pirateAttack), 0, 0},
			Shield:   [3]float64{float64(pirateShield), 0, 0},
			Shell:    float64(pirateShell),
		}},
	}
	input := battle.Input{
		Seed:      deriveSeed(fleetID),
		Rounds:    6,
		Attackers: []battle.Side{{UserID: "expedition", Units: atkUnits}},
		Defenders: []battle.Side{pirateSide},
	}
	report, err := battle.Calculate(input)
	if err != nil {
		return nil, fmt.Errorf("expedition: battle: %w", err)
	}
	if _, err := applyAttackerLosses(ctx, tx, fleetID, ships, report.Attackers[0].Units); err != nil {
		return nil, fmt.Errorf("expedition: losses: %w", err)
	}
	return map[string]any{
		"winner":       report.Winner,
		"rounds":       report.Rounds,
		"pirate_fleet": "5 × light_fighter",
	}, nil
}

// expLoss — теряем 5..20% каждого ship-stack'а.
func expLoss(ctx context.Context, tx pgx.Tx, r *rng.R, fleetID string,
	ships []unitStack) (map[string]any, error) {
	frac := 0.05 + r.Float64()*0.15
	losses := map[int]int64{}
	for _, sh := range ships {
		lost := int64(math.Ceil(float64(sh.Count) * frac))
		if lost <= 0 {
			continue
		}
		if lost > sh.Count {
			lost = sh.Count
		}
		newCount := sh.Count - lost
		if newCount <= 0 {
			if _, err := tx.Exec(ctx,
				`DELETE FROM fleet_ships WHERE fleet_id=$1 AND unit_id=$2`,
				fleetID, sh.UnitID); err != nil {
				return nil, err
			}
		} else {
			if _, err := tx.Exec(ctx,
				`UPDATE fleet_ships SET count=$1 WHERE fleet_id=$2 AND unit_id=$3`,
				newCount, fleetID, sh.UnitID); err != nil {
				return nil, err
			}
		}
		losses[sh.UnitID] = lost
	}
	return map[string]any{"lost": losses, "fraction": frac}, nil
}
