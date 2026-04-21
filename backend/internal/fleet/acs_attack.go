// ATTACK_ALLIANCE (mission=12) — ACS: несколько флотов разных игроков
// атакуют одну цель одновременно.
//
// Поток:
//  1. Читаем acs_group_id из payload.
//  2. Для каждого флота группы собираем battle.Side (корабли + тех).
//  3. Один battle.Calculate с Attackers=[]Side{s1, s2, …}.
//  4. Применяем потери к каждому fleet_ships отдельно.
//  5. Loot делим поровну между выжившими атакующими флотами.
//  6. Пишем battle_reports + сообщения.
//
// Упрощения (M5 ACS):
//   - Loot делится поровну по числу выживших флотов (не по грузоподъёмности).
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/pkg/ids"
)

type acsPayload struct {
	FleetID    string `json:"fleet_id"`
	ACSGroupID string `json:"acs_group_id"`
}

// ACSAttackHandler — event.Handler для KindAttackAlliance=12.
// Каждый флот в группе получает своё событие, но обрабатывается только
// тем событием, чей fleet является «лидером» (первый в порядке created_at).
// Остальные события группы пропускаются (fleet.state уже 'returning').
func (s *TransportService) ACSAttackHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl acsPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("acs attack: parse payload: %w", err)
		}
		if pl.ACSGroupID == "" {
			return fmt.Errorf("acs attack: missing acs_group_id in payload")
		}

		// Проверяем, что этот флот ещё 'outbound' (не обработан другим событием группы).
		var state string
		err := tx.QueryRow(ctx,
			`SELECT state FROM fleets WHERE id=$1 FOR UPDATE`, pl.FleetID).Scan(&state)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("acs attack: read fleet: %w", err)
		}
		if state != "outbound" {
			return nil // уже обработан
		}

		// Читаем все флоты группы (outbound).
		type fleetInfo struct {
			id          string
			ownerUserID string
			cm, cs, ch  int64
			isMoon      bool
			g, sys, pos int
		}
		rows, err := tx.Query(ctx, `
			SELECT id, owner_user_id, carried_metal, carried_silicon, carried_hydrogen,
			       dst_is_moon, dst_galaxy, dst_system, dst_position
			FROM fleets
			WHERE acs_group_id=$1 AND state='outbound'
			ORDER BY created_at ASC
			FOR UPDATE
		`, pl.ACSGroupID)
		if err != nil {
			return fmt.Errorf("acs attack: read group fleets: %w", err)
		}
		var fleets []fleetInfo
		for rows.Next() {
			var f fleetInfo
			if err := rows.Scan(&f.id, &f.ownerUserID, &f.cm, &f.cs, &f.ch,
				&f.isMoon, &f.g, &f.sys, &f.pos); err != nil {
				return err
			}
			fleets = append(fleets, f)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		if len(fleets) == 0 {
			return nil
		}

		// Координаты цели от первого флота.
		lead := fleets[0]
		isMoon := lead.isMoon

		// Читаем цель.
		var planetID, defenderUserID string
		var defMetal, defSil, defHydro float64
		err = tx.QueryRow(ctx, `
			SELECT id, user_id, metal, silicon, hydrogen
			FROM planets
			WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
			  AND destroyed_at IS NULL
			FOR UPDATE
		`, lead.g, lead.sys, lead.pos, isMoon).Scan(
			&planetID, &defenderUserID, &defMetal, &defSil, &defHydro)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Цель исчезла — все флоты возвращают.
				for _, f := range fleets {
					if _, uerr := tx.Exec(ctx,
						`UPDATE fleets SET state='returning' WHERE id=$1`, f.id); uerr != nil {
						return uerr
					}
				}
				return nil
			}
			return fmt.Errorf("acs attack: find target: %w", err)
		}

		// Читаем защиту цели.
		defShips, err := readPlanetShips(ctx, tx, planetID)
		if err != nil {
			return fmt.Errorf("acs attack: def ships: %w", err)
		}
		var defDefense []unitStack
		if !isMoon {
			defDefense, err = readPlanetDefense(ctx, tx, planetID)
			if err != nil {
				return fmt.Errorf("acs attack: def defense: %w", err)
			}
		}
		defTech, err := readUserTech(ctx, tx, defenderUserID)
		if err != nil {
			return fmt.Errorf("acs attack: def tech: %w", err)
		}
		defUnits := stacksToBattleUnits(defShips, s.catalog, false)
		defUnits = append(defUnits, stacksToBattleUnits(defDefense, s.catalog, true)...)
		defSide := battle.Side{UserID: defenderUserID, Tech: defTech, Units: defUnits}

		// Собираем стороны атакующих.
		type atkFleet struct {
			info  fleetInfo
			ships []unitStack
			side  battle.Side
		}
		var atkFleets []atkFleet
		for _, f := range fleets {
			ships, err := readFleetShips(ctx, tx, f.id)
			if err != nil {
				return fmt.Errorf("acs attack: read fleet_ships %s: %w", f.id, err)
			}
			tech, err := readUserTech(ctx, tx, f.ownerUserID)
			if err != nil {
				return fmt.Errorf("acs attack: read tech %s: %w", f.ownerUserID, err)
			}
			units := stacksToBattleUnits(ships, s.catalog, false)
			if len(units) == 0 {
				continue
			}
			atkFleets = append(atkFleets, atkFleet{
				info:  f,
				ships: ships,
				side:  battle.Side{UserID: f.ownerUserID, Tech: tech, Units: units},
			})
		}

		if len(atkFleets) == 0 {
			for _, f := range fleets {
				if _, uerr := tx.Exec(ctx,
					`UPDATE fleets SET state='returning' WHERE id=$1`, f.id); uerr != nil {
					return uerr
				}
			}
			return nil
		}

		// Пустая планета — бой без сражения.
		var report battle.Report
		if len(defSide.Units) == 0 {
			report = battle.Report{Winner: "attackers", Rounds: 0, Seed: deriveSeed(pl.FleetID)}
		} else {
			attackerSides := make([]battle.Side, len(atkFleets))
			for i, af := range atkFleets {
				attackerSides[i] = af.side
			}
			inp := battle.Input{
				Seed:      deriveSeed(pl.FleetID),
				Rounds:    6,
				Attackers: attackerSides,
				Defenders: []battle.Side{defSide},
				Rapidfire: rapidfireToMap(s.catalog),
				IsMoon:    isMoon,
			}
			report, err = battle.Calculate(inp)
			if err != nil {
				return fmt.Errorf("acs attack: battle: %w", err)
			}
		}

		// Применяем потери атакующих по каждому флоту.
		type survivorFleet struct {
			info      fleetInfo
			survivors []unitStack
		}
		var survivingFleets []survivorFleet
		for i, af := range atkFleets {
			var endUnits []battle.UnitResult
			if i < len(report.Attackers) {
				endUnits = report.Attackers[i].Units
			}
			survivors, err := applyAttackerLosses(ctx, tx, af.info.id, af.ships, endUnits)
			if err != nil {
				return fmt.Errorf("acs attack: apply losses fleet %s: %w", af.info.id, err)
			}
			if len(survivors) > 0 {
				survivingFleets = append(survivingFleets, survivorFleet{info: af.info, survivors: survivors})
			}
		}

		// Применяем потери защитника.
		if len(report.Defenders) > 0 {
			if err := applyDefenderLosses(ctx, tx, planetID, defShips, defDefense,
				report.Defenders[0].Units); err != nil {
				return fmt.Errorf("acs attack: apply defender losses: %w", err)
			}
		}

		// Debris + moon.
		defenseIDs := map[int]bool{}
		for _, d := range defDefense {
			defenseIDs[d.UnitID] = true
		}
		debrisM, debrisS := calcDebris(report, defenseIDs, s.catalog)
		if debrisM > 0 || debrisS > 0 {
			if _, err := tx.Exec(ctx, `
				INSERT INTO debris_fields (galaxy, system, position, is_moon, metal, silicon)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (galaxy, system, position, is_moon) DO UPDATE
				SET metal = debris_fields.metal + EXCLUDED.metal,
				    silicon = debris_fields.silicon + EXCLUDED.silicon,
				    last_update = now()
			`, lead.g, lead.sys, lead.pos, isMoon, debrisM, debrisS); err != nil {
				return fmt.Errorf("acs attack: debris: %w", err)
			}
			if !isMoon {
				if err := tryCreateMoon(ctx, tx, lead.g, lead.sys, lead.pos,
					debrisM+debrisS, report.Seed, defenderUserID, lead.ownerUserID); err != nil {
					return fmt.Errorf("acs attack: moon: %w", err)
				}
			}
		}

		// Loot при победе атакующих — делим пропорционально cargo capacity
		// выживших кораблей (суммарный Cargo по стекам).
		if report.Winner == "attackers" && len(survivingFleets) > 0 {
			totalAvailM := float64(defMetal * 0.5)
			totalAvailS := float64(defSil * 0.5)
			totalAvailH := float64(defHydro * 0.5)

			// Считаем суммарный cargo каждого флота и total.
			cargoPerFleet := make([]int64, len(survivingFleets))
			var totalCargo int64
			for i, sf := range survivingFleets {
				for _, st := range sf.survivors {
					for _, spec := range s.catalog.Ships.Ships {
						if spec.ID == st.UnitID {
							cargoPerFleet[i] += spec.Cargo * st.Count
							break
						}
					}
				}
				totalCargo += cargoPerFleet[i]
			}
			// Если cargo у всех нулевой — делим поровну (fallback).
			if totalCargo == 0 {
				for i := range cargoPerFleet {
					cargoPerFleet[i] = 1
				}
				totalCargo = int64(len(survivingFleets))
			}

			totalLootM, totalLootS, totalLootH := int64(0), int64(0), int64(0)
			for i, sf := range survivingFleets {
				ratio := float64(cargoPerFleet[i]) / float64(totalCargo)
				perFleetM := totalAvailM * ratio
				perFleetS := totalAvailS * ratio
				perFleetH := totalAvailH * ratio
				loot := grabLoot(perFleetM, perFleetS, perFleetH,
					sf.survivors, s.catalog, sf.info.cm, sf.info.cs, sf.info.ch)
				if loot.Metal > 0 || loot.Silicon > 0 || loot.Hydrogen > 0 {
					if _, err := tx.Exec(ctx, `
						UPDATE fleets SET carried_metal=$1, carried_silicon=$2, carried_hydrogen=$3
						WHERE id=$4
					`, sf.info.cm+loot.Metal, sf.info.cs+loot.Silicon, sf.info.ch+loot.Hydrogen,
						sf.info.id); err != nil {
						return fmt.Errorf("acs attack: fleet carry: %w", err)
					}
					totalLootM += loot.Metal
					totalLootS += loot.Silicon
					totalLootH += loot.Hydrogen
				}
			}
			if totalLootM > 0 || totalLootS > 0 || totalLootH > 0 {
				if _, err := tx.Exec(ctx, `
					UPDATE planets SET metal=metal-$1, silicon=silicon-$2, hydrogen=hydrogen-$3
					WHERE id=$4
				`, totalLootM, totalLootS, totalLootH, planetID); err != nil {
					return fmt.Errorf("acs attack: subtract loot: %w", err)
				}
				// Аудит: defender теряет, лидер — как представитель группы.
				if _, err := tx.Exec(ctx, `
					INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
					VALUES ($1, $2, 'loot', $3, $4, $5),
					       ($6, $7, 'loot', $8, $9, $10)
				`, lead.ownerUserID, planetID, totalLootM, totalLootS, totalLootH,
					defenderUserID, planetID, -totalLootM, -totalLootS, -totalLootH,
				); err != nil {
					return fmt.Errorf("acs attack: res_log: %w", err)
				}
			}
		}

		// battle_reports — один record от имени лидера + acs_participants.
		type acsParticipant struct {
			UserID  string `json:"user_id"`
			FleetID string `json:"fleet_id"`
		}
		participants := make([]acsParticipant, 0, len(atkFleets))
		for _, af := range atkFleets {
			participants = append(participants, acsParticipant{
				UserID:  af.info.ownerUserID,
				FleetID: af.info.id,
			})
		}
		participantsJSON, _ := json.Marshal(participants)

		reportJSON, _ := json.Marshal(report)
		reportID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO battle_reports (id, attacker_user_id, defender_user_id, planet_id,
			                            seed, winner, rounds,
			                            debris_metal, debris_silicon,
			                            loot_metal, loot_silicon, loot_hydrogen,
			                            report, acs_participants)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`, reportID, lead.ownerUserID, defenderUserID, planetID,
			int64(report.Seed), report.Winner, report.Rounds,
			debrisM, debrisS, int64(0), int64(0), int64(0),
			reportJSON, participantsJSON); err != nil {
			return fmt.Errorf("acs attack: insert report: %w", err)
		}

		// Сообщения всем атакующим + защитнику.
		subject := fmt.Sprintf("ACS боевой отчёт: %s", report.Winner)
		body := fmt.Sprintf("Раундов: %d. ACS-группа %s.", report.Rounds, pl.ACSGroupID)
		for _, af := range atkFleets {
			if _, err := tx.Exec(ctx, `
				INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, battle_report_id)
				VALUES ($1, $2, $3, 2, $4, $5, $6)
			`, ids.New(), af.info.ownerUserID, defenderUserID, subject, body, reportID); err != nil {
				return fmt.Errorf("acs attack: attacker message: %w", err)
			}
		}
		if defenderUserID != "" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, battle_report_id)
				VALUES ($1, $2, $3, 2, $4, $5, $6)
			`, ids.New(), defenderUserID, lead.ownerUserID, subject, body, reportID); err != nil {
				return fmt.Errorf("acs attack: defender message: %w", err)
			}
		}

		// Все флоты → returning.
		for _, f := range fleets {
			if _, err := tx.Exec(ctx,
				`UPDATE fleets SET state='returning' WHERE id=$1`, f.id); err != nil {
				return fmt.Errorf("acs attack: fleet state: %w", err)
			}
		}
		return nil
	}
}
