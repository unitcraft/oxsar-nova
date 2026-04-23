package alien

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/economy"
	"github.com/oxsar/nova/backend/pkg/rng"
)

type unitStack struct {
	UnitID       int
	Count        int64
	Damaged      int64
	ShellPercent float64
}

func readPlanetShips(ctx context.Context, tx pgx.Tx, planetID string) ([]unitStack, error) {
	rows, err := tx.Query(ctx, `
		SELECT unit_id, count, damaged, shell_percent
		FROM ships WHERE planet_id = $1 AND count > 0
	`, planetID)
	if err != nil {
		return nil, fmt.Errorf("alien: read ships: %w", err)
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
	rows, err := tx.Query(ctx, `
		SELECT unit_id, count, 0::bigint, 0::float8
		FROM defense WHERE planet_id = $1 AND count > 0
	`, planetID)
	if err != nil {
		return nil, fmt.Errorf("alien: read defense: %w", err)
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

func readUserTech(ctx context.Context, tx pgx.Tx, userID string, cat *config.Catalog) (battle.Tech, error) {
	rows, err := tx.Query(ctx, `SELECT unit_id, level FROM research WHERE user_id = $1`, userID)
	if err != nil {
		return battle.Tech{}, fmt.Errorf("alien: read tech: %w", err)
	}
	defer rows.Close()
	levels := map[int]int{}
	for rows.Next() {
		var uid, lvl int
		if err := rows.Scan(&uid, &lvl); err != nil {
			return battle.Tech{}, err
		}
		levels[uid] = lvl
	}
	if err := rows.Err(); err != nil {
		return battle.Tech{}, err
	}

	// Применить бонусы профессии.
	var prof string
	_ = tx.QueryRow(ctx, `SELECT profession FROM users WHERE id=$1`, userID).Scan(&prof)
	if prof != "" && prof != "none" {
		if spec, ok := cat.Professions.Professions[prof]; ok {
			for k, v := range spec.Bonus {
				if id, ok2 := economy.ProfessionKeyToID[k]; ok2 {
					levels[id] += v
				}
			}
			for k, v := range spec.Malus {
				if id, ok2 := economy.ProfessionKeyToID[k]; ok2 {
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

func stacksToBattleUnits(stacks []unitStack, cat *config.Catalog, isDefense bool) []battle.Unit {
	var out []battle.Unit
	for _, s := range stacks {
		if s.Count <= 0 {
			continue
		}
		var attack, shield, shell int
		var cost config.ResCost
		var found bool
		if isDefense {
			for _, spec := range cat.Defense.Defense {
				if spec.ID == s.UnitID {
					attack, shield, shell, cost, found = spec.Attack, spec.Shield, spec.Shell, spec.Cost, true
					break
				}
			}
		} else {
			for _, spec := range cat.Ships.Ships {
				if spec.ID == s.UnitID {
					attack, shield, shell, cost, found = spec.Attack, spec.Shield, spec.Shell, spec.Cost, true
					break
				}
			}
		}
		if !found {
			continue
		}
		out = append(out, battle.Unit{
			UnitID:       s.UnitID,
			Quantity:     s.Count,
			Damaged:      s.Damaged,
			ShellPercent: s.ShellPercent,
			Attack:       [3]float64{float64(attack), 0, 0},
			Shield:       [3]float64{float64(shield), 0, 0},
			Shell:        float64(shell),
			Cost:         battle.UnitCost{Metal: cost.Metal, Silicon: cost.Silicon, Hydrogen: cost.Hydrogen},
		})
	}
	return out
}

// calcDefPower — суммарная боевая мощь обороняющихся (attack × quantity).
// Используется для масштабирования флота пришельцев.
func calcDefPower(units []battle.Unit) float64 {
	var total float64
	for _, u := range units {
		maxAtk := u.Attack[0]
		for _, a := range u.Attack[1:] {
			if a > maxAtk {
				maxAtk = a
			}
		}
		total += maxAtk * float64(u.Quantity)
	}
	return total
}

// alienShipOrder — порядок добавления кораблей пришельцев от слабых к сильным.
// ID совпадают с configs/ships.yml (unit_a_corvette=200 .. unit_a_torpedocarier=204).
var alienShipOrder = []struct {
	unitID int
	name   string
}{
	{200, "Alien Corvette"},
	{201, "Alien Screen"},
	{202, "Alien Paladin"},
	{203, "Alien Frigate"},
	{204, "Alien Torpedocarrier"},
}

// scaledAlienFleet создаёт флот пришельцев с суммарной мощью 90–110% от defPower.
// Использует корабли UNIT_A_* из каталога. Если defPower = 0 (планета пуста)
// или каталог не содержит alien-кораблей — возвращает fallback из 5 alien corvette.
func scaledAlienFleet(defPower float64, r *rng.R, cat *config.Catalog) []battle.Unit {
	// Целевая мощь: defPower × random(0.9, 1.1).
	scale := 0.9 + float64(r.IntN(21))/100.0
	targetPower := defPower * scale
	if targetPower < 100 {
		targetPower = 100 // минимальная сила атаки даже для пустой планеты
	}

	// Найти характеристики кораблей пришельцев в каталоге.
	type alienUnit struct {
		unitID  int
		name    string
		attack  float64
		shell   float64
		shield  float64
		front   int
	}
	var shipDefs []alienUnit
	for _, entry := range alienShipOrder {
		for _, spec := range cat.Ships.Ships {
			if spec.ID == entry.unitID {
				shipDefs = append(shipDefs, alienUnit{
					unitID: entry.unitID,
					name:   entry.name,
					attack: float64(spec.Attack),
					shell:  float64(spec.Shell),
					shield: float64(spec.Shield),
					front:  0, // Ships каталог не хранит Front для пришельцев
				})
				break
			}
		}
	}
	if len(shipDefs) == 0 {
		// Каталог не загружен или alien ships отсутствуют — fallback.
		return []battle.Unit{{
			UnitID: 200, Quantity: 5, Name: "Alien Corvette",
			Attack: [3]float64{150, 0, 0}, Shell: 2000,
		}}
	}

	// Итеративно добавляем корабли от слабых к сильным, пока не достигнем targetPower.
	var result []battle.Unit
	var currentPower float64
	for currentPower < targetPower {
		remaining := targetPower - currentPower
		// Выбираем самый сильный корабль, который не превышает remaining × 1.5.
		chosen := shipDefs[0]
		for _, sd := range shipDefs {
			if sd.attack <= remaining*1.5 && sd.attack > chosen.attack {
				chosen = sd
			}
		}
		// Сколько таких кораблей добавить (от 1 до 20).
		maxAdd := int(remaining/chosen.attack) + 1
		if maxAdd > 20 {
			maxAdd = 20
		}
		if maxAdd < 1 {
			maxAdd = 1
		}

		// Найти или создать запись для этого unit_id.
		found := false
		for i := range result {
			if result[i].UnitID == chosen.unitID {
				result[i].Quantity += int64(maxAdd)
				found = true
				break
			}
		}
		if !found {
			result = append(result, battle.Unit{
				UnitID:   chosen.unitID,
				Quantity: int64(maxAdd),
				Name:     chosen.name,
				Attack:   [3]float64{chosen.attack, 0, 0},
				Shell:    chosen.shell,
				Shield:   [3]float64{chosen.shield, 0, 0},
			})
		}
		currentPower += chosen.attack * float64(maxAdd)

		// Защита от бесконечного цикла при очень малом attack.
		if chosen.attack <= 0 {
			break
		}
	}
	return result
}

func fnvHash(s string) uint64 {
	var h uint64 = 14695981039346656037
	for i := range len(s) {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

func rapidfireToMap(cat *config.Catalog) map[int]map[int]int {
	if cat == nil {
		return nil
	}
	return cat.Rapidfire.Rapidfire
}

func applyDefenderLosses(ctx context.Context, tx pgx.Tx, planetID string,
	startShips, startDefense []unitStack, end []battle.UnitResult) error {
	endByID := map[int]battle.UnitResult{}
	for _, r := range end {
		endByID[r.UnitID] = r
	}
	for _, s := range startShips {
		r, ok := endByID[s.UnitID]
		if !ok {
			continue
		}
		if r.QuantityEnd == 0 {
			if _, err := tx.Exec(ctx,
				`UPDATE ships SET count=0, damaged_count=0, shell_percent=0 WHERE planet_id=$1 AND unit_id=$2`,
				planetID, s.UnitID); err != nil {
				return err
			}
			continue
		}
		if _, err := tx.Exec(ctx,
			`UPDATE ships SET count=$1, damaged_count=$2, shell_percent=$3 WHERE planet_id=$4 AND unit_id=$5`,
			r.QuantityEnd, r.DamagedEnd, r.ShellPercentEnd, planetID, s.UnitID); err != nil {
			return err
		}
	}
	for _, s := range startDefense {
		r, ok := endByID[s.UnitID]
		if !ok {
			continue
		}
		cnt := r.QuantityEnd
		if _, err := tx.Exec(ctx,
			`UPDATE defense SET count=$1 WHERE planet_id=$2 AND unit_id=$3`,
			cnt, planetID, s.UnitID); err != nil {
			return err
		}
	}
	return nil
}
