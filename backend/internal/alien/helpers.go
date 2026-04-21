package alien

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/config"
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

func readUserTech(ctx context.Context, tx pgx.Tx, userID string) (battle.Tech, error) {
	rows, err := tx.Query(ctx, `SELECT unit_id, level FROM research WHERE user_id = $1`, userID)
	if err != nil {
		return battle.Tech{}, fmt.Errorf("alien: read tech: %w", err)
	}
	defer rows.Close()
	var tech battle.Tech
	for rows.Next() {
		var uid, lvl int
		if err := rows.Scan(&uid, &lvl); err != nil {
			return tech, err
		}
		switch uid {
		case 109:
			tech.Gun = lvl
		case 110:
			tech.Shield = lvl
		case 111:
			tech.Shell = lvl
		case 113:
			tech.Ballistics = lvl
		case 104:
			tech.Masking = lvl
		}
	}
	return tech, rows.Err()
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
