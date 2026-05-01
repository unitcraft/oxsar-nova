// План 72.1.26 ч.B/C — расширения /resource: solar satellite,
// halting fleets, virt-fleet/defense/stock_fleet ряды.
//
// Legacy: `Planet.class.php::getProduction()` строки 316-540 + `Resource.class.php::loadBuildingData()` строки 124-196.

package planet

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
)

// Legacy константы из consts.php / Functions.inc.php.
const (
	unitSolarSatellite     = 39
	unitVirtFleet          = 1_000_000
	unitVirtStockFleet     = 1_000_001
	unitVirtDefense        = 1_000_002
	unitVirtHaltingStart   = 1_000_003

	// Legacy MARKET_BASE_CURS_* (consts.php 645-647).
	marketBaseCursMetal    = 600.0
	marketBaseCursSilicon  = 400.0
	marketBaseCursHydrogen = 200.0

	// Legacy unitGroupConsumptionPerHour (Functions.inc.php 1681).
	unitsGroupConsumptionPowerBase   = 1.000003
	maxGroupUnitConsumptionPerHour   = 0.1
	groupConsumptionMinShips         = 1000

	// energy_tech ID (research.yml: energy_tech).
	unitEnergyTech = 113
)

// unitGroupConsumptionPerHour — порт legacy формулы потребления
// водорода большой группой кораблей/обороны (Functions.inc.php:1681).
//
//	if (!in_fling && count < 1000) return 0;
//	scale = unitid == UNIT_VIRT_DEFENSE ? 0.5 : 1.0;
//	cons = min(MAX, pow(BASE, count) * scale / 10 / 24) * count;
//	return cons >= 0.01 ? cons : 0;
//
// in_fling=true используется только для летящего флота — здесь всегда false.
func unitGroupConsumptionPerHour(unitID int, count int64) float64 {
	if count < groupConsumptionMinShips {
		return 0
	}
	scale := 1.0
	if unitID == unitVirtDefense {
		scale = 0.5
	}
	// pow(BASE, count) для больших count может переполниться.
	// legacy полагается на min(MAX, ...) — clamp до MAX.
	powVal := math.Pow(unitsGroupConsumptionPowerBase, float64(count))
	if math.IsInf(powVal, 0) || math.IsNaN(powVal) {
		powVal = math.Inf(1)
	}
	cons := math.Min(maxGroupUnitConsumptionPerHour, powVal*scale/10.0/24.0) * float64(count)
	if cons < 0.01 {
		return 0
	}
	return cons
}

// shipyardCounts читает агрегаты из ships/defense/exchange_lots:
//   - virtFleetCount: SUM(count) FROM ships WHERE unit_id != solar_satellite
//   - virtDefenseCount: SUM(count) FROM defense
//   - virtStockFleet: SUM(ships) из открытых лотов биржи на этой планете
//   - solarSatCount: count FROM ships WHERE unit_id = solar_satellite
func (s *Service) shipyardCounts(ctx context.Context, planetID string) (virtFleet, virtDefense, virtStock, solarSat int64, _ error) {
	pool := s.db.Pool()

	// solar satellite — отдельно (id=39).
	if err := pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(count), 0) FROM ships WHERE planet_id = $1 AND unit_id = $2`,
		planetID, unitSolarSatellite,
	).Scan(&solarSat); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("shipyard solar count: %w", err)
	}

	// все остальные ships (mode=fleet в legacy).
	if err := pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(count), 0) FROM ships WHERE planet_id = $1 AND unit_id != $2`,
		planetID, unitSolarSatellite,
	).Scan(&virtFleet); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("shipyard fleet count: %w", err)
	}

	if err := pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(count), 0) FROM defense WHERE planet_id = $1`,
		planetID,
	).Scan(&virtDefense); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("defense count: %w", err)
	}

	// stock_fleet: лоты биржи кораблей. План 65 F.5 fleet_lots на этой
	// планете в state='open'. Если таблицы нет в окружении — 0.
	if err := pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(qty), 0) FROM (
			SELECT (jsonb_each_text(ships)).value::bigint AS qty
			FROM fleet_lots WHERE src_planet_id = $1 AND state = 'open'
		) t`,
		planetID,
	).Scan(&virtStock); err != nil {
		// fleet_lots может отсутствовать — игнорируем.
		virtStock = 0
	}
	return
}

// HaltingFleetInfo — строка halting fleet, удерживающего планету.
type HaltingFleetInfo struct {
	FleetID      string
	SrcCoord     string         // "1:23:4"
	TotalShips   int64
	ShipsBy      map[int]int64
}

// haltingFleets — выборка флотов в state='hold' с dst-координатами этой планеты.
//
// Legacy: `events WHERE mode=EVENT_HOLDING AND processed=WAIT AND destination=planetid`.
// Origin: state='hold' в fleets-таблице + match dst по координатам planet.
func (s *Service) haltingFleets(ctx context.Context, p *Planet) ([]HaltingFleetInfo, error) {
	pool := s.db.Pool()
	rows, err := pool.Query(ctx, `
		SELECT f.id,
		       sp.galaxy, sp.system, sp.position,
		       fs.unit_id, fs.count
		FROM fleets f
		JOIN fleet_ships fs ON fs.fleet_id = f.id
		LEFT JOIN planets sp ON sp.id = f.src_planet_id
		WHERE f.state = 'hold'
		  AND f.dst_galaxy = $1
		  AND f.dst_system = $2
		  AND f.dst_position = $3
		  AND f.dst_is_moon = $4
		ORDER BY f.depart_at ASC, f.id ASC, fs.unit_id ASC
	`, p.Galaxy, p.System, p.Position, p.IsMoon)
	if err != nil {
		return nil, fmt.Errorf("halting fleets: %w", err)
	}
	defer rows.Close()

	byID := make(map[string]*HaltingFleetInfo)
	order := []string{}
	for rows.Next() {
		var (
			fleetID string
			gal, sys, pos *int
			unitID int
			count int64
		)
		if err := rows.Scan(&fleetID, &gal, &sys, &pos, &unitID, &count); err != nil {
			return nil, fmt.Errorf("halting scan: %w", err)
		}
		info, ok := byID[fleetID]
		if !ok {
			coord := ""
			if gal != nil && sys != nil && pos != nil {
				coord = fmt.Sprintf("%d:%d:%d", *gal, *sys, *pos)
			}
			info = &HaltingFleetInfo{FleetID: fleetID, SrcCoord: coord, ShipsBy: make(map[int]int64)}
			byID[fleetID] = info
			order = append(order, fleetID)
		}
		info.ShipsBy[unitID] += count
		info.TotalShips += count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]HaltingFleetInfo, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out, nil
}

// energyTechLevel читает уровень energy_tech у игрока.
func (s *Service) energyTechLevel(ctx context.Context, userID string) (int, error) {
	var lvl int
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT COALESCE(level, 0) FROM research WHERE user_id = $1 AND unit_id = $2`,
		userID, unitEnergyTech,
	).Scan(&lvl); err != nil {
		// no row — уровень 0.
		if err.Error() == pgx.ErrNoRows.Error() {
			return 0, nil
		}
		return 0, nil
	}
	return lvl, nil
}
