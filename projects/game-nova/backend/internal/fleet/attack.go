// ATTACK_SINGLE (mission=10) — живой бой с планетой-целью.
//
// Поток прибытия:
//  1. Читаем fleet + fleet_ships (атакующие юниты).
//  2. Находим планету-цель. Нет цели (разрушена/свободна) →
//     state='returning', без боя.
//  3. Собираем defenders: ships + defense (если не moon) + tech хозяина.
//  4. Собираем attackers: fleet_ships + tech владельца флота.
//  5. battle.Calculate(...) — результат боя.
//  6. Применяем потери:
//     * атакующему → fleet_ships row-wise (count - lost).
//     * защитнику → ships/defense (count - lost).
//  7. При победе атакующего — loot: 50% доступных ресурсов цели,
//     ограниченный свободным cargo флота.
//  8. Пишем battle_reports + 2 messages (attacker & defender).
//  9. fleet.state='returning', carry += loot.
//
// KindReturn=20 остаётся как был — возвращает корабли и carry на
// src_planet. Если attacker проиграл, выживших в fleet_ships не
// остаётся, handler KindReturn завершит с нулями.
//
// Осознанные упрощения M4.4a:
//   * только single attacker (ACS — в M5).
//   * debris целиком в loot (упрощение). Отдельная миссия
//     RECYCLING (kind=9) придёт позже.
//   * moon-chance: реализован (min(20, debris/100000)%). Создаётся
//     при первом бое с достаточным полем обломков.
//   * report-тексты простые.
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/alien"
	"oxsar/game-nova/internal/artefact"
	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/internal/battlestats"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/pkg/ids"
	"oxsar/game-nova/pkg/rng"
)

// unitStack — «плоская» запись в ships/defense/fleet_ships.
type unitStack struct {
	UnitID       int
	Count        int64
	Damaged      int64
	ShellPercent float64
}

// AttackHandler — event.Handler для KindAttackSingle=10.
func (s *TransportService) AttackHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl transportPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("attack: parse payload: %w", err)
		}

		var (
			state                     string
			attackerUserID, srcPlanet string
			g, sys, pos               int
			isMoon                    bool
			cm, csil, ch              int64
		)
		err := tx.QueryRow(ctx, `
			SELECT state, owner_user_id, src_planet_id,
			       dst_galaxy, dst_system, dst_position, dst_is_moon,
			       carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id = $1 FOR UPDATE
		`, pl.FleetID).Scan(&state, &attackerUserID, &srcPlanet,
			&g, &sys, &pos, &isMoon, &cm, &csil, &ch)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("attack: read fleet: %w", err)
		}
		if state != "outbound" {
			return nil
		}
		_ = srcPlanet

		attackerShips, err := readFleetShips(ctx, tx, pl.FleetID)
		if err != nil {
			return fmt.Errorf("attack: read fleet_ships: %w", err)
		}
		attackerTech, err := readUserTech(ctx, tx, attackerUserID, s.catalog)
		if err != nil {
			return fmt.Errorf("attack: read attacker tech: %w", err)
		}

		var (
			planetID                   string
			defenderUserID             string
			defMetal, defSil, defHydro float64
		)
		err = tx.QueryRow(ctx, `
			SELECT id, user_id, metal, silicon, hydrogen
			FROM planets
			WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
			  AND destroyed_at IS NULL
			FOR UPDATE
		`, g, sys, pos, isMoon).Scan(&planetID, &defenderUserID,
			&defMetal, &defSil, &defHydro)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_, uerr := tx.Exec(ctx,
					`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID)
				return uerr
			}
			return fmt.Errorf("attack: find target: %w", err)
		}

		// Если защитник в режиме отпуска — флот возвращается без боя.
		var vacSince *int64 // просто проверяем IS NOT NULL
		_ = tx.QueryRow(ctx,
			`SELECT 1 FROM users WHERE id=$1 AND vacation_since IS NOT NULL`,
			defenderUserID).Scan(&vacSince)
		if vacSince != nil {
			_, uerr := tx.Exec(ctx,
				`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID)
			return uerr
		}

		// Защита защитника от атаки. Срабатывает если активен любой
		// из трёх механизмов:
		//  - global protectionPeriod (новички — created_at > NOW() - X);
		//  - per-user protected_until_at > NOW() (план 69 D-004);
		//  - is_observer = true (план 69 D-005, наблюдатель не участвует
		//    в боях).
		var protected bool
		_ = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM users
				WHERE id = $1
				  AND (
				    ($2 > 0 AND created_at > NOW() - ($2 || ' seconds')::interval)
				    OR (protected_until_at IS NOT NULL AND protected_until_at > NOW())
				    OR is_observer = true
				  )
			)
		`, defenderUserID, s.protectionPeriod).Scan(&protected)
		if protected {
			_, uerr := tx.Exec(ctx,
				`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID)
			return uerr
		}

		defenderShips, err := readPlanetShips(ctx, tx, planetID)
		if err != nil {
			return fmt.Errorf("attack: defender ships: %w", err)
		}
		var defenderDefense []unitStack
		if !isMoon {
			defenderDefense, err = readPlanetDefense(ctx, tx, planetID)
			if err != nil {
				return fmt.Errorf("attack: defender defense: %w", err)
			}
		}
		defenderTech, err := readUserTech(ctx, tx, defenderUserID, s.catalog)
		if err != nil {
			return fmt.Errorf("attack: defender tech: %w", err)
		}

		battleMod, err := s.artefact.ActiveBattleModifiers(ctx, tx, attackerUserID)
		if err != nil {
			return fmt.Errorf("attack: battle modifiers: %w", err)
		}

		atkUnits := stacksToBattleUnits(attackerShips, s.catalog, false)
		atkUnits = applyBattleMod(atkUnits, battleMod)

		atkSide := battle.Side{
			UserID: attackerUserID,
			Tech:   attackerTech,
			Units:  atkUnits,
		}
		defUnits := stacksToBattleUnits(defenderShips, s.catalog, false)
		defUnits = append(defUnits, stacksToBattleUnits(defenderDefense, s.catalog, true)...)
		defSide := battle.Side{
			UserID: defenderUserID,
			Tech:   defenderTech,
			Units:  defUnits,
		}

		// Alien HOLDING: пришельцы на орбите защищают планету как
		// дополнительный defender (план 15, Assault::loadDefenders в legacy).
		// Если планета не в HOLDING — holdingDef == nil.
		holdingDef, err := alien.LoadHoldingDefender(ctx, tx, planetID, s.catalog)
		if err != nil {
			return fmt.Errorf("attack: load alien holding: %w", err)
		}
		defenders := []battle.Side{defSide}
		if holdingDef != nil {
			defenders = append(defenders, holdingDef.Side)
		}

		if len(atkSide.Units) == 0 {
			_, uerr := tx.Exec(ctx,
				`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID)
			return uerr
		}

		// Пустая планета + нет alien-защиты — без боя, сразу loot.
		// Debris=0 (ship'ов не было, уничтожать нечего).
		if len(defSide.Units) == 0 && holdingDef == nil {
			loot := grabLoot(defMetal, defSil, defHydro, attackerShips, s.catalog, cm, csil, ch)
			rep := battle.Report{Winner: "attackers", Rounds: 0, Seed: deriveSeed(pl.FleetID)}
			return finalizeAttack(ctx, tx, s.bundle, pl.FleetID, attackerUserID, defenderUserID, planetID,
				rep, loot, 0, 0, cm, csil, ch, 0, 0)
		}

		atkPower := sidePower(atkSide.Units)
		defPower := sidePower(defSide.Units)
		if holdingDef != nil {
			defPower += sidePower(holdingDef.Side.Units)
		}

		input := battle.Input{
			Seed:      deriveSeed(pl.FleetID),
			Rounds:    6,
			Attackers: []battle.Side{atkSide},
			Defenders: defenders,
			Rapidfire: rapidfireToMap(s.catalog),
			IsMoon:    isMoon,
			HasPlanet: true, // атака на планету или луну → опыт без штрафа (BA-007).
		}
		report, err := battle.Calculate(input)
		if err != nil {
			return fmt.Errorf("attack: battle: %w", err)
		}

		atkSurvivors, err := applyAttackerLosses(ctx, tx, pl.FleetID, attackerShips, report.Attackers[0].Units)
		if err != nil {
			return fmt.Errorf("attack: apply attacker losses: %w", err)
		}
		if err := applyDefenderLosses(ctx, tx, planetID, defenderShips, defenderDefense,
			report.Defenders[0].Units); err != nil {
			return fmt.Errorf("attack: apply defender losses: %w", err)
		}

		// Alien HOLDING: обновить payload (actual survivors) или закрыть
		// событие, если alien-флот полностью разбит. Defender index 1 —
		// именно alien-сторона (append порядок выше).
		if holdingDef != nil && len(report.Defenders) > 1 {
			if err := alien.CloseHoldingIfWiped(ctx, tx, holdingDef.EventID,
				report.Defenders[1].Units); err != nil {
				return fmt.Errorf("attack: close alien holding: %w", err)
			}
		}

		// Debris: 30% metal+silicon от стоимости уничтоженных SHIPS
		// (defense не переходит в debris — OGame-правило). Считаем
		// по UnitResult обеих сторон, защитные юниты defenderDefense
		// исключаем по unit_id.
		defenseIDs := map[int]bool{}
		for _, d := range defenderDefense {
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
			`, g, sys, pos, isMoon, debrisM, debrisS); err != nil {
				return fmt.Errorf("attack: write debris: %w", err)
			}
			// Moon-chance: min(20, total_debris/100000)%.
			if !isMoon {
				if err := tryCreateMoon(ctx, tx, s.bundle, g, sys, pos, debrisM+debrisS,
					report.Seed, defenderUserID, attackerUserID); err != nil {
					return fmt.Errorf("attack: moon: %w", err)
				}
			}
		}

		var loot lootAmount
		if report.Winner == "attackers" && len(atkSurvivors) > 0 {
			loot = grabLoot(defMetal, defSil, defHydro, atkSurvivors, s.catalog, cm, csil, ch)
		}
		// План 20 Ф.6: Moon Destruction (kind=25/27). Только если цель —
		// луна и атакующий выжил с DS. Roll по OGame-формуле, при успехе
		// помечаем луну destroyed_at и шлём сообщение обоим игрокам.
		if (e.Kind == event.KindAttackDestroyMoon ||
			e.Kind == event.KindAttackAllianceDestroyMoon) && isMoon {
			if err := s.tryDestroyMoon(ctx, tx, pl.FleetID, planetID,
				attackerUserID, defenderUserID, atkSurvivors, report.Seed); err != nil {
				return fmt.Errorf("attack: moon destroy: %w", err)
			}
		}
		// План 65 Ф.3 (D-037): Building Destruction. Только если цель —
		// планета (не луна) и атакующий выжил. Понижает уровень здания
		// на 1 (или удаляет, если level=1→0).
		if e.Kind == event.KindAttackDestroyBuilding {
			unitID, lvlFrom, lvlTo, ok, err := tryDestroyBuilding(ctx, tx,
				planetID, isMoon, report.Winner, pl.TargetBuildingID, report.Seed)
			if err != nil {
				return fmt.Errorf("attack: building destroy: %w", err)
			}
			if ok {
				if err := sendBuildingDestroyedMessages(ctx, tx, s.tr,
					defenderUserID, attackerUserID, unitID, lvlFrom, lvlTo, e.ID); err != nil {
					return fmt.Errorf("attack: building destroy msg: %w", err)
				}
			}
		}
		return finalizeAttack(ctx, tx, s.bundle, pl.FleetID, attackerUserID, defenderUserID, planetID,
			report, loot, debrisM, debrisS, cm, csil, ch, atkPower, defPower)
	}
}

// calcDebris — 30% (metal+silicon) от стоимости ships, погибших в
// бою. defenseIDs — идентификаторы defensive-юнитов (чтобы исключить
// их из debris). Cost per-unit берём из каталога по UnitResult.UnitID.
func calcDebris(rep battle.Report, defenseIDs map[int]bool, cat *config.Catalog) (int64, int64) {
	var m, s int64
	sides := append([]battle.SideResult{}, rep.Attackers...)
	sides = append(sides, rep.Defenders...)
	for _, side := range sides {
		for _, u := range side.Units {
			if defenseIDs[u.UnitID] {
				continue
			}
			lost := u.QuantityStart - u.QuantityEnd
			if lost <= 0 {
				continue
			}
			// ищем cost в Ships (все атакующие — ships, defenders без
			// defense тоже ships).
			for _, spec := range cat.Ships.Ships {
				if spec.ID == u.UnitID {
					m += lost * spec.Cost.Metal * 30 / 100
					s += lost * spec.Cost.Silicon * 30 / 100
					break
				}
			}
		}
	}
	return m, s
}

// -----------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------

func readFleetShips(ctx context.Context, tx pgx.Tx, fleetID string) ([]unitStack, error) {
	rows, err := tx.Query(ctx,
		`SELECT unit_id, count, damaged_count FROM fleet_ships WHERE fleet_id=$1`, fleetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []unitStack
	for rows.Next() {
		var s unitStack
		if err := rows.Scan(&s.UnitID, &s.Count, &s.Damaged); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func readPlanetShips(ctx context.Context, tx pgx.Tx, planetID string) ([]unitStack, error) {
	rows, err := tx.Query(ctx, `
		SELECT unit_id, count, damaged_count, shell_percent
		FROM ships WHERE planet_id=$1 AND count > 0
		FOR UPDATE
	`, planetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []unitStack
	for rows.Next() {
		var s unitStack
		if err := rows.Scan(&s.UnitID, &s.Count, &s.Damaged, &s.ShellPercent); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func readPlanetDefense(ctx context.Context, tx pgx.Tx, planetID string) ([]unitStack, error) {
	rows, err := tx.Query(ctx,
		`SELECT unit_id, count FROM defense WHERE planet_id=$1 AND count > 0 FOR UPDATE`,
		planetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []unitStack
	for rows.Next() {
		var s unitStack
		if err := rows.Scan(&s.UnitID, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func readUserTech(ctx context.Context, tx pgx.Tx, userID string, cat *config.Catalog) (battle.Tech, error) {
	rows, err := tx.Query(ctx,
		`SELECT unit_id, level FROM research WHERE user_id=$1`, userID)
	if err != nil {
		return battle.Tech{}, err
	}
	defer rows.Close()
	levels := map[int]int{}
	for rows.Next() {
		var id, lvl int
		if err := rows.Scan(&id, &lvl); err != nil {
			return battle.Tech{}, err
		}
		levels[id] = lvl
	}
	if err := rows.Err(); err != nil {
		return battle.Tech{}, err
	}

	// Применить бонусы профессии (виртуальные уровни).
	var prof string
	_ = tx.QueryRow(ctx, `SELECT profession FROM users WHERE id=$1`, userID).Scan(&prof)
	if prof != "" && prof != "none" {
		if spec, ok := cat.Professions.Professions[prof]; ok {
			for k, v := range spec.Bonus {
				if id, ok := economy.ProfessionKeyToID[k]; ok {
					levels[id] += v
				}
			}
			for k, v := range spec.Malus {
				if id, ok := economy.ProfessionKeyToID[k]; ok {
					levels[id] += v
				}
			}
		}
	}

	return battle.Tech{
		Gun:        levels[economy.IDTechGun],
		Shield:     levels[economy.IDTechShield],
		Shell:      levels[economy.IDTechShell],
		Laser:      levels[economy.IDTechLaser],
		Ion:        levels[economy.IDTechSilicon],
		Plasma:     levels[economy.IDTechHydrogen],
		Ballistics: levels[economy.IDTechBallistics],
		Masking:    levels[economy.IDTechMasking],
	}, nil
}

// stacksToBattleUnits — unitStack[] → battle.Unit[] через каталог.
// Юниты без каталожной записи пропускаются (устаревший unit_id).
// isDefense → ищем в Defense-каталоге, иначе в Ships.
func stacksToBattleUnits(stacks []unitStack, cat *config.Catalog, isDefense bool) []battle.Unit {
	out := make([]battle.Unit, 0, len(stacks))
	for _, s := range stacks {
		if s.Count <= 0 {
			continue
		}
		var (
			attack, shell    int
			cost             config.ResCost
			cargo            int64
			speed, fuel      int
			shieldVal        int
			found            bool
		)
		if isDefense {
			for _, spec := range cat.Defense.Defense {
				if spec.ID == s.UnitID {
					attack, shell, cost, shieldVal, found = spec.Attack, spec.Shell, spec.Cost, spec.Shield, true
					break
				}
			}
		} else {
			for _, spec := range cat.Ships.Ships {
				if spec.ID == s.UnitID {
					attack, shell, cost = spec.Attack, spec.Shell, spec.Cost
					cargo, speed, fuel = spec.Cargo, spec.Speed, spec.Fuel
					shieldVal = spec.Shield
					found = true
					break
				}
			}
		}
		if !found {
			continue
		}
		_ = cargo
		_ = speed
		_ = fuel
		out = append(out, battle.Unit{
			UnitID:       s.UnitID,
			Quantity:     s.Count,
			Damaged:      s.Damaged,
			ShellPercent: s.ShellPercent,
			Front:        0,
			Attack:       float64(attack),
			Shield:       float64(shieldVal),
			Shell:        float64(shell),
			Cost:         battle.UnitCost{Metal: cost.Metal, Silicon: cost.Silicon, Hydrogen: cost.Hydrogen},
		})
	}
	return out
}

// deriveSeed — детерминированный seed из fleetID (FNV-1a на первых
// байтах UUID). Не полагаемся на math/rand.
func deriveSeed(fleetID string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(fleetID); i++ {
		h ^= uint64(fleetID[i])
		h *= 1099511628211
	}
	return h
}

// rapidfireToMap — таблица rapidfire из каталога. В нашем YAML
// (configs/rapidfire.yml) ключи изначально unit_id, поэтому конвертер
// тривиальный — возвращаем as is. nil-map легальна (engine читает
// как rf=1 для всех пар).
func rapidfireToMap(cat *config.Catalog) map[int]map[int]int {
	if cat == nil {
		return nil
	}
	return cat.Rapidfire.Rapidfire
}

// grabLoot — 50% metal/silicon/hydrogen цели, зажатое свободным
// cargo флота (после карго уже существующего carry).
func grabLoot(m, si, h float64, survivors []unitStack, cat *config.Catalog,
	cm, cs, ch int64) lootAmount {
	var totalCap int64
	for _, s := range survivors {
		for _, spec := range cat.Ships.Ships {
			if spec.ID == s.UnitID {
				totalCap += spec.Cargo * s.Count
				break
			}
		}
	}
	free := totalCap - (cm + cs + ch)
	if free <= 0 {
		return lootAmount{}
	}
	want := lootAmount{
		Metal:    int64(m * 0.5),
		Silicon:  int64(si * 0.5),
		Hydrogen: int64(h * 0.5),
	}
	total := want.Metal + want.Silicon + want.Hydrogen
	if total > free && total > 0 {
		k := float64(free) / float64(total)
		want.Metal = int64(float64(want.Metal) * k)
		want.Silicon = int64(float64(want.Silicon) * k)
		want.Hydrogen = int64(float64(want.Hydrogen) * k)
	}
	return want
}

type lootAmount struct {
	Metal    int64
	Silicon  int64
	Hydrogen int64
}

// applyAttackerLosses — апдейт fleet_ships по результатам боя.
// Возвращает выживших (для loot-пересчёта).
func applyAttackerLosses(ctx context.Context, tx pgx.Tx, fleetID string,
	start []unitStack, end []battle.UnitResult) ([]unitStack, error) {
	endByID := map[int]battle.UnitResult{}
	for _, r := range end {
		endByID[r.UnitID] = r
	}
	var survivors []unitStack
	for _, s := range start {
		r, ok := endByID[s.UnitID]
		if !ok {
			// в report нет записи → ничего не меняем (не должно случаться)
			survivors = append(survivors, s)
			continue
		}
		if r.QuantityEnd == 0 {
			if _, err := tx.Exec(ctx,
				`DELETE FROM fleet_ships WHERE fleet_id=$1 AND unit_id=$2`,
				fleetID, s.UnitID); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := tx.Exec(ctx, `
			UPDATE fleet_ships
			SET count = $1, damaged_count = $2
			WHERE fleet_id=$3 AND unit_id=$4
		`, r.QuantityEnd, r.DamagedEnd, fleetID, s.UnitID); err != nil {
			return nil, err
		}
		survivors = append(survivors, unitStack{
			UnitID: s.UnitID, Count: r.QuantityEnd, Damaged: r.DamagedEnd,
			ShellPercent: r.ShellPercentEnd,
		})
	}
	return survivors, nil
}

// applyDefenderLosses — ships + defense на планете. end — один
// сплошной массив UnitResult (сначала ships-позиции, потом defense),
// в том же порядке, в котором их скормили в battle.Side.
func applyDefenderLosses(ctx context.Context, tx pgx.Tx, planetID string,
	startShips, startDefense []unitStack, end []battle.UnitResult) error {
	// Сопоставление by UnitID. Одинаковых ID в ships и defense быть
	// не должно (ships > 200, defense > 300 в legacy).
	endByID := map[int]battle.UnitResult{}
	for _, r := range end {
		endByID[r.UnitID] = r
	}
	apply := func(table string, stacks []unitStack) error {
		for _, s := range stacks {
			r, ok := endByID[s.UnitID]
			if !ok {
				continue
			}
			if r.QuantityEnd == 0 {
				if _, err := tx.Exec(ctx,
					`UPDATE `+table+` SET count=0, damaged_count=0, shell_percent=0
					 WHERE planet_id=$1 AND unit_id=$2`,
					planetID, s.UnitID); err != nil {
					return err
				}
				continue
			}
			if _, err := tx.Exec(ctx, `
				UPDATE `+table+`
				SET count=$1, damaged_count=$2, shell_percent=$3
				WHERE planet_id=$4 AND unit_id=$5
			`, r.QuantityEnd, r.DamagedEnd, r.ShellPercentEnd,
				planetID, s.UnitID); err != nil {
				return err
			}
		}
		return nil
	}
	if err := apply("ships", startShips); err != nil {
		return err
	}
	return apply("defense", startDefense)
}

// applyBattleMod применяет боевые модификаторы к юнитам.
// Множители умножаются на Attack/Shield/Shell каждого юнита.
func applyBattleMod(units []battle.Unit, m artefact.BattleModifier) []battle.Unit {
	for i := range units {
		units[i].Attack *= m.AttackMul
		units[i].Shield *= m.ShieldMul
		units[i].Shell *= m.ShellMul
	}
	return units
}

// finalizeAttack — запись battle_reports + 2 messages + списание
// ресурсов с планеты (loot) + обновление fleet.state + carry.
//
// _unusedSurvivors оставлен для API-симметрии; при победе loot уже
// посчитан в grabLoot.
// sidePower — суммарная атака стороны в первом раунде (Java: startBattleAtterPower).
func sidePower(units []battle.Unit) float64 {
	var total float64
	for _, u := range units {
		total += u.Attack * float64(u.Quantity)
	}
	return total
}

// calcExperience — удалён в плане 72.1.1 ч.2: формула опыта теперь
// внутри battle.Calculate (battle.computeExperience), результат —
// в Report.AttackerExp/DefenderExp. Зачисление в users — через
// battlestats.ApplyBattleResult, который читает поля Report.
//
// Старый порт здесь имел тонкую расходящую семантику (isMoon vs
// hasPlanet); единый источник истины устранил риск рассинхрона.

func finalizeAttack(ctx context.Context, tx pgx.Tx, b *i18n.Bundle,
	fleetID, attUID, defUID, planetID string,
	rep battle.Report, loot lootAmount,
	debrisM, debrisS int64,
	prevM, prevS, prevH int64,
	atkPower, defPower float64) error {

	// План 72.1.1: e_points/be_points/battles + списание points/u_points/
	// u_count + лог user_experience — теперь делает battlestats.ApplyBattleResult,
	// вызывается ниже после INSERT в battle_reports (нужен reportID для
	// UNIQUE-ключа idempotency).

	// Начисление кредитов победителю пропорционально мощи противника.
	if rep.Winner == "attackers" && attUID != "" {
		cr := economy.BattleWinCredits(defPower)
		if cr > 0 {
			if _, err := tx.Exec(ctx,
				`UPDATE users SET credit=credit+$1 WHERE id=$2`, cr, attUID,
			); err != nil {
				return fmt.Errorf("finalize: attacker credit: %w", err)
			}
		}
	} else if rep.Winner == "defenders" && defUID != "" {
		cr := economy.BattleWinCredits(atkPower)
		if cr > 0 {
			if _, err := tx.Exec(ctx,
				`UPDATE users SET credit=credit+$1 WHERE id=$2`, cr, defUID,
			); err != nil {
				return fmt.Errorf("finalize: defender credit: %w", err)
			}
		}
	}

	// battle_reports
	reportJSON, err := json.Marshal(rep)
	if err != nil {
		return fmt.Errorf("finalize: marshal report: %w", err)
	}
	reportID := ids.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO battle_reports (id, attacker_user_id, defender_user_id, planet_id,
		                            seed, winner, rounds,
		                            debris_metal, debris_silicon,
		                            loot_metal, loot_silicon, loot_hydrogen,
		                            report)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, reportID, attUID, defUID, planetID,
		int64(rep.Seed), rep.Winner, rep.Rounds,
		debrisM, debrisS,
		loot.Metal, loot.Silicon, loot.Hydrogen,
		reportJSON,
	); err != nil {
		return fmt.Errorf("finalize: insert report: %w", err)
	}

	// План 72.1.1: зачислить опыт/потери юзерам (e_points/be_points/
	// battles +, points/u_points/u_count -). Idempotent через UNIQUE
	// (battle_id, user_id, is_atter) в user_experience — повторный
	// вызов вернёт ErrAlreadyApplied, что мы трактуем как ОК
	// (event re-process).
	if err := battlestats.ApplyBattleResult(ctx, tx, rep, reportID); err != nil &&
		!errors.Is(err, battlestats.ErrAlreadyApplied) {
		return fmt.Errorf("finalize: apply battle result: %w", err)
	}

	// Списываем loot с цели, добавляем к carry флота.
	if loot.Metal > 0 || loot.Silicon > 0 || loot.Hydrogen > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal=metal-$1, silicon=silicon-$2, hydrogen=hydrogen-$3
			WHERE id=$4
		`, loot.Metal, loot.Silicon, loot.Hydrogen, planetID); err != nil {
			return fmt.Errorf("finalize: subtract loot: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE fleets SET carried_metal=$1, carried_silicon=$2, carried_hydrogen=$3
			WHERE id=$4
		`, prevM+loot.Metal, prevS+loot.Silicon, prevH+loot.Hydrogen, fleetID); err != nil {
			return fmt.Errorf("finalize: add carry: %w", err)
		}
		// Аудит: attacker получает ресурсы, defender теряет.
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'loot', $3, $4, $5),
			       ($6, $7, 'loot', $8, $9, $10)
		`, attUID, planetID, loot.Metal, loot.Silicon, loot.Hydrogen,
			defUID, planetID, -loot.Metal, -loot.Silicon, -loot.Hydrogen,
		); err != nil {
			return fmt.Errorf("finalize: res_log: %w", err)
		}
	}

	// Messages для обеих сторон. Folder 2 = inbox/battle в legacy.
	btr := bundleTr(b)
	subject := btr("assaultReport", "subject", map[string]string{"winner": rep.Winner})
	body := btr("assaultReport", "body", map[string]string{
		"rounds":   strconv.Itoa(rep.Rounds),
		"metal":    strconv.FormatInt(loot.Metal, 10),
		"silicon":  strconv.FormatInt(loot.Silicon, 10),
		"hydrogen": strconv.FormatInt(loot.Hydrogen, 10),
	})
	if _, err := tx.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, battle_report_id)
		VALUES ($1, $2, $3, 2, $4, $5, $6)
	`, ids.New(), attUID, defUID, subject, body, reportID); err != nil {
		return fmt.Errorf("finalize: attacker message: %w", err)
	}
	if defUID != "" && defUID != attUID {
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, battle_report_id)
			VALUES ($1, $2, $3, 2, $4, $5, $6)
		`, ids.New(), defUID, attUID, subject, body, reportID); err != nil {
			return fmt.Errorf("finalize: defender message: %w", err)
		}
	}

	// Флот → возврат.
	if _, err := tx.Exec(ctx,
		`UPDATE fleets SET state='returning' WHERE id=$1`, fleetID); err != nil {
		return fmt.Errorf("finalize: fleet state: %w", err)
	}
	return nil
}

// tryCreateMoon проверяет шанс создания луны по формуле OGame:
// chance = min(20, debrisTotal/100000)%. Если луна уже есть — пропуск.
// seed берётся из battle.Report.Seed для детерминированности.
func tryCreateMoon(ctx context.Context, tx pgx.Tx, b *i18n.Bundle, g, sys, pos int,
	debrisTotal int64, battleSeed uint64, defUserID, attUserID string) error {
	chance := int(debrisTotal / 100000)
	if chance > 20 {
		chance = 20
	}
	if chance <= 0 {
		return nil
	}
	r := rng.New(battleSeed ^ uint64(g)<<32 ^ uint64(sys)<<16 ^ uint64(pos))
	if r.IntN(100) >= chance {
		return nil // не повезло
	}
	// Луна уже есть?
	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM planets
			WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=true AND destroyed_at IS NULL
		)
	`, g, sys, pos).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	// Размер луны — 2000..6800 (OGame-диапазон для мун).
	diameter := 2000 + r.IntN(4800)
	moonID := ids.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO planets (id, user_id, is_moon, name, galaxy, system, position,
		                     diameter, used_fields, planet_type, temperature_min, temperature_max,
		                     metal, silicon, hydrogen)
		VALUES ($1, $2, true, 'Moon', $3, $4, $5, $6, 0, 'moon', -100, -60, 0, 0, 0)
	`, moonID, defUserID, g, sys, pos, diameter); err != nil {
		return fmt.Errorf("insert moon: %w", err)
	}
	// Сообщения обеим сторонам.
	btr := bundleTr(b)
	gs, ss, ps, ds := strconv.Itoa(g), strconv.Itoa(sys), strconv.Itoa(pos), strconv.Itoa(diameter)
	subj := btr("assaultReport", "moonCreatedSubject", map[string]string{"g": gs, "s": ss, "pos": ps})
	body := btr("assaultReport", "moonCreatedBody", map[string]string{"g": gs, "s": ss, "pos": ps, "diameter": ds})
	for _, uid := range []string{defUserID, attUserID} {
		if uid == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 2, $3, $4)
		`, ids.New(), uid, subj, body); err != nil {
			return fmt.Errorf("moon message: %w", err)
		}
	}
	return nil
}
