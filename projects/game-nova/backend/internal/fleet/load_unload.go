// Package fleet — load/unload resources для HOLDING-флотов.
//
// План 72.1.47: legacy `Mission.class.php::loadResourcesToFleet` и
// `unloadResourcesFromFleet`. Доступно только для флота в state='hold'
// на dst-планете, где сейчас находится игрок.
//
// Семантика:
//   - load (planet → fleet): должен иметь:
//     - fleet.state = 'hold'
//     - fleet.dst = current planet (galaxy/system/position/is_moon)
//     - fleet.src != current planet (нельзя грузить на свою же исходную)
//     - cargo capacity >= already_carried + load_amount
//   - unload (fleet → planet): должен иметь:
//     - fleet.state = 'hold'
//     - fleet.dst = current planet
//     - carried >= unload_amount
//
// Идемпотентность: каждый запрос — отдельная транзакция; при повторе
// клиент должен использовать новый Idempotency-Key, иначе get-кешируется.

package fleet

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrFleetNotHolding   = errors.New("fleet: not in holding state")
	ErrPlanetNotDst      = errors.New("fleet: current planet is not destination")
	ErrLoadFromOwnSrc    = errors.New("fleet: cannot load from own source planet")
	ErrLoadCapacity      = errors.New("fleet: insufficient cargo capacity")
	ErrUnloadInsufficient = errors.New("fleet: not enough resources on fleet to unload")
)

// LoadUnloadInput — параметры для load/unload.
type LoadUnloadInput struct {
	UserID         string
	FleetID        string
	CurrentPlanetID string
	Metal          int64
	Silicon        int64
	Hydrogen       int64
}

// LoadResources — план 72.1.47: загрузить ресурсы с планеты во флот
// (legacy `loadResourcesToFleet`). Списывает ресурсы с планеты, увеличивает
// fleets.carried_*. capacity-check учитывает текущий carried и cargo всех
// кораблей флота.
func (s *TransportService) LoadResources(ctx context.Context, in LoadUnloadInput) error {
	if in.Metal < 0 || in.Silicon < 0 || in.Hydrogen < 0 {
		return fmt.Errorf("%w: negative amount", ErrInvalidDispatch)
	}
	if in.Metal == 0 && in.Silicon == 0 && in.Hydrogen == 0 {
		return nil
	}
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Lock fleet и проверка state/dst/src.
		var (
			ownerID, srcPlanet           string
			state                        string
			dstGalaxy, dstSystem, dstPos int
			dstIsMoon                    bool
			cm, csil, ch                 int64
		)
		err := tx.QueryRow(ctx, `
			SELECT owner_user_id, src_planet_id, state, dst_galaxy, dst_system, dst_position, dst_is_moon,
			       carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id=$1 FOR UPDATE
		`, in.FleetID).Scan(&ownerID, &srcPlanet, &state, &dstGalaxy, &dstSystem, &dstPos, &dstIsMoon, &cm, &csil, &ch)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrFleetNotFound
			}
			return fmt.Errorf("load: select fleet: %w", err)
		}
		if ownerID != in.UserID {
			return ErrPlanetOwnership
		}
		if state != "hold" {
			return ErrFleetNotHolding
		}
		// 2. CurrentPlanet должен быть dst.
		var (
			curUser    string
			cpG, cpS, cpP int
			cpIsMoon   bool
			pMetal, pSilicon, pHydro float64
		)
		err = tx.QueryRow(ctx, `
			SELECT user_id, galaxy, system, position, is_moon,
			       COALESCE(metal,0), COALESCE(silicon,0), COALESCE(hydrogen,0)
			FROM planets WHERE id=$1 FOR UPDATE
		`, in.CurrentPlanetID).Scan(&curUser, &cpG, &cpS, &cpP, &cpIsMoon, &pMetal, &pSilicon, &pHydro)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTargetNotFound
			}
			return fmt.Errorf("load: select planet: %w", err)
		}
		if curUser != in.UserID {
			return ErrPlanetOwnership
		}
		if cpG != dstGalaxy || cpS != dstSystem || cpP != dstPos || cpIsMoon != dstIsMoon {
			return ErrPlanetNotDst
		}
		if in.CurrentPlanetID == srcPlanet {
			return ErrLoadFromOwnSrc
		}
		// 3. Capacity check.
		totalCap, err := s.fleetCargoCapacity(ctx, tx, in.FleetID)
		if err != nil {
			return err
		}
		used := cm + csil + ch
		free := totalCap - used
		if free < 0 {
			free = 0
		}
		want := in.Metal + in.Silicon + in.Hydrogen
		if want > free {
			return fmt.Errorf("%w: free=%d want=%d", ErrLoadCapacity, free, want)
		}
		// 4. Доступность на планете (clamp).
		if in.Metal > int64(pMetal) {
			in.Metal = int64(pMetal)
		}
		if in.Silicon > int64(pSilicon) {
			in.Silicon = int64(pSilicon)
		}
		if in.Hydrogen > int64(pHydro) {
			in.Hydrogen = int64(pHydro)
		}
		if in.Metal == 0 && in.Silicon == 0 && in.Hydrogen == 0 {
			return nil
		}
		// 5. Atomic swap: списываем с планеты, зачисляем во fleet.carry.
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal=metal-$1, silicon=silicon-$2, hydrogen=hydrogen-$3 WHERE id=$4
		`, in.Metal, in.Silicon, in.Hydrogen, in.CurrentPlanetID); err != nil {
			return fmt.Errorf("load: charge planet: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE fleets SET carried_metal=carried_metal+$1, carried_silicon=carried_silicon+$2,
			                  carried_hydrogen=carried_hydrogen+$3
			WHERE id=$4
		`, in.Metal, in.Silicon, in.Hydrogen, in.FleetID); err != nil {
			return fmt.Errorf("load: credit fleet: %w", err)
		}
		// 6. res_log.
		_, _ = tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'fleet_load', $3, $4, $5)
		`, in.UserID, in.CurrentPlanetID, -in.Metal, -in.Silicon, -in.Hydrogen)
		return nil
	})
}

// UnloadResources — план 72.1.47: разгрузить ресурсы с флота на планету
// (legacy `unloadResourcesFromFleet`). Списывает с fleets.carried_*,
// зачисляет на planets.
func (s *TransportService) UnloadResources(ctx context.Context, in LoadUnloadInput) error {
	if in.Metal < 0 || in.Silicon < 0 || in.Hydrogen < 0 {
		return fmt.Errorf("%w: negative amount", ErrInvalidDispatch)
	}
	if in.Metal == 0 && in.Silicon == 0 && in.Hydrogen == 0 {
		return nil
	}
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			ownerID                      string
			state                        string
			dstGalaxy, dstSystem, dstPos int
			dstIsMoon                    bool
			cm, csil, ch                 int64
		)
		err := tx.QueryRow(ctx, `
			SELECT owner_user_id, state, dst_galaxy, dst_system, dst_position, dst_is_moon,
			       carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id=$1 FOR UPDATE
		`, in.FleetID).Scan(&ownerID, &state, &dstGalaxy, &dstSystem, &dstPos, &dstIsMoon, &cm, &csil, &ch)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrFleetNotFound
			}
			return fmt.Errorf("unload: select fleet: %w", err)
		}
		if ownerID != in.UserID {
			return ErrPlanetOwnership
		}
		if state != "hold" {
			return ErrFleetNotHolding
		}
		var (
			curUser                  string
			cpG, cpS, cpP            int
			cpIsMoon                 bool
		)
		err = tx.QueryRow(ctx, `
			SELECT user_id, galaxy, system, position, is_moon
			FROM planets WHERE id=$1 FOR UPDATE
		`, in.CurrentPlanetID).Scan(&curUser, &cpG, &cpS, &cpP, &cpIsMoon)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTargetNotFound
			}
			return fmt.Errorf("unload: select planet: %w", err)
		}
		if curUser != in.UserID {
			return ErrPlanetOwnership
		}
		if cpG != dstGalaxy || cpS != dstSystem || cpP != dstPos || cpIsMoon != dstIsMoon {
			return ErrPlanetNotDst
		}
		// Clamp по carried.
		if in.Metal > cm {
			in.Metal = cm
		}
		if in.Silicon > csil {
			in.Silicon = csil
		}
		if in.Hydrogen > ch {
			in.Hydrogen = ch
		}
		if in.Metal == 0 && in.Silicon == 0 && in.Hydrogen == 0 {
			return nil
		}
		if _, err := tx.Exec(ctx, `
			UPDATE fleets SET carried_metal=carried_metal-$1, carried_silicon=carried_silicon-$2,
			                  carried_hydrogen=carried_hydrogen-$3
			WHERE id=$4
		`, in.Metal, in.Silicon, in.Hydrogen, in.FleetID); err != nil {
			return fmt.Errorf("unload: charge fleet: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal=metal+$1, silicon=silicon+$2, hydrogen=hydrogen+$3 WHERE id=$4
		`, in.Metal, in.Silicon, in.Hydrogen, in.CurrentPlanetID); err != nil {
			return fmt.Errorf("unload: credit planet: %w", err)
		}
		_, _ = tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'fleet_unload', $3, $4, $5)
		`, in.UserID, in.CurrentPlanetID, in.Metal, in.Silicon, in.Hydrogen)
		return nil
	})
}

// fleetCargoCapacity суммирует cargo всех кораблей флота из catalog.
func (s *TransportService) fleetCargoCapacity(ctx context.Context, tx pgx.Tx, fleetID string) (int64, error) {
	rows, err := tx.Query(ctx, `SELECT unit_id, count FROM fleet_ships WHERE fleet_id=$1`, fleetID)
	if err != nil {
		return 0, fmt.Errorf("read fleet_ships: %w", err)
	}
	defer rows.Close()
	var total int64
	for rows.Next() {
		var unitID int
		var cnt int64
		if err := rows.Scan(&unitID, &cnt); err != nil {
			return 0, err
		}
		for _, spec := range s.catalog.Ships.Ships {
			if spec.ID == unitID {
				total += spec.Cargo * cnt
				break
			}
		}
	}
	return total, rows.Err()
}
