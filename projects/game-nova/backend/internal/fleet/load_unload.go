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
	// План 72.1.48: legacy `getRemainFleetControls` + back_consumption.
	ErrControlsExhausted     = errors.New("fleet: max load/unload operations exceeded")
	ErrInsufficientReturnFuel = errors.New("fleet: cannot unload H below back_consumption reserve")
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
		// 1. Lock fleet + control_times + back_consumption.
		var (
			ownerID, srcPlanet           string
			state                        string
			dstGalaxy, dstSystem, dstPos int
			dstIsMoon                    bool
			cm, csil, ch                 int64
			controlTimes, maxControls    int
		)
		err := tx.QueryRow(ctx, `
			SELECT owner_user_id, src_planet_id, state, dst_galaxy, dst_system, dst_position, dst_is_moon,
			       carried_metal, carried_silicon, carried_hydrogen,
			       control_times, max_control_times
			FROM fleets WHERE id=$1 FOR UPDATE
		`, in.FleetID).Scan(&ownerID, &srcPlanet, &state, &dstGalaxy, &dstSystem, &dstPos, &dstIsMoon,
			&cm, &csil, &ch, &controlTimes, &maxControls)
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
		// План 72.1.48: rate-limit control_times.
		if maxControls > 0 && controlTimes >= maxControls {
			return ErrControlsExhausted
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
		// 3. Комиссия (legacy `getControlComis`): planet_owner == fleet_owner
		// → exchange_rate брокера (default 1.2 = 20%); заглушка: 5% если своё.
		// Берём комиссию по флоту-владельцу (in.UserID == ownerID — гарантировано выше).
		comis, err := s.getControlCommission(ctx, tx, in.UserID, curUser == ownerID)
		if err != nil {
			return err
		}
		// 4. Capacity check.
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
		// 5. Доступность на планете c учётом комиссии (clamp).
		// Расход с планеты: requested + ceil(requested × comis / 100).
		// `getControlResource(stock, comis)` = floor(stock × 100 / (comis + 100)).
		maxFromStock := func(stock int64) int64 {
			if comis <= 0 {
				return stock
			}
			return stock * 100 / (int64(comis) + 100)
		}
		if in.Metal > maxFromStock(int64(pMetal)) {
			in.Metal = maxFromStock(int64(pMetal))
		}
		if in.Silicon > maxFromStock(int64(pSilicon)) {
			in.Silicon = maxFromStock(int64(pSilicon))
		}
		if in.Hydrogen > maxFromStock(int64(pHydro)) {
			in.Hydrogen = maxFromStock(int64(pHydro))
		}
		if in.Metal == 0 && in.Silicon == 0 && in.Hydrogen == 0 {
			return nil
		}
		// Реальное списание = requested + комиссия (на планете).
		debitMetal := in.Metal + ceilFee(in.Metal, comis)
		debitSilicon := in.Silicon + ceilFee(in.Silicon, comis)
		debitHydro := in.Hydrogen + ceilFee(in.Hydrogen, comis)
		// 6. Atomic swap.
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal=metal-$1, silicon=silicon-$2, hydrogen=hydrogen-$3 WHERE id=$4
		`, debitMetal, debitSilicon, debitHydro, in.CurrentPlanetID); err != nil {
			return fmt.Errorf("load: charge planet: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE fleets SET carried_metal=carried_metal+$1, carried_silicon=carried_silicon+$2,
			                  carried_hydrogen=carried_hydrogen+$3,
			                  control_times=control_times+1
			WHERE id=$4
		`, in.Metal, in.Silicon, in.Hydrogen, in.FleetID); err != nil {
			return fmt.Errorf("load: credit fleet: %w", err)
		}
		// 7. res_log.
		_, _ = tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'fleet_load', $3, $4, $5)
		`, in.UserID, in.CurrentPlanetID, -debitMetal, -debitSilicon, -debitHydro)
		return nil
	})
}

// ceilFee = ceil(x × comis / 100). Для целочисленного percent.
func ceilFee(x int64, comis int) int64 {
	if comis <= 0 || x <= 0 {
		return 0
	}
	return (x*int64(comis) + 99) / 100
}

// getControlCommission — план 72.1.48: legacy `getControlComis`.
// Закомментировано в legacy: 5% если свой holding. Активная ветка:
// (exchange_rate - 1) × 100 (default 1.2 → 20%).
func (s *TransportService) getControlCommission(ctx context.Context, tx pgx.Tx, userID string, _ bool) (int, error) {
	var rate float64
	if err := tx.QueryRow(ctx,
		`SELECT COALESCE(exchange_rate, 1.2) FROM users WHERE id=$1`, userID,
	).Scan(&rate); err != nil {
		return 0, fmt.Errorf("control comis: %w", err)
	}
	if rate < 1.0 {
		rate = 1.0
	}
	c := int((rate - 1.0) * 100.0)
	if c < 0 {
		c = 0
	}
	return c, nil
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
			controlTimes, maxControls    int
			backConsumption              int64
		)
		err := tx.QueryRow(ctx, `
			SELECT owner_user_id, state, dst_galaxy, dst_system, dst_position, dst_is_moon,
			       carried_metal, carried_silicon, carried_hydrogen,
			       control_times, max_control_times, back_consumption
			FROM fleets WHERE id=$1 FOR UPDATE
		`, in.FleetID).Scan(&ownerID, &state, &dstGalaxy, &dstSystem, &dstPos, &dstIsMoon,
			&cm, &csil, &ch, &controlTimes, &maxControls, &backConsumption)
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
		// План 72.1.48: rate-limit.
		if maxControls > 0 && controlTimes >= maxControls {
			return ErrControlsExhausted
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
		// План 72.1.48: back_consumption check — нельзя выгрузить H ниже
		// резерва на возврат (legacy `data.hydrogen >= back_consumption`).
		remainingH := ch - in.Hydrogen
		if remainingH < backConsumption {
			return fmt.Errorf("%w: back_consumption=%d, would remain %d",
				ErrInsufficientReturnFuel, backConsumption, remainingH)
		}
		if in.Metal == 0 && in.Silicon == 0 && in.Hydrogen == 0 {
			return nil
		}
		// Комиссия для unload — со списания флота. На планету попадает
		// requested - комиссия.
		comis, err := s.getControlCommission(ctx, tx, in.UserID, curUser == ownerID)
		if err != nil {
			return err
		}
		creditMetal := in.Metal - ceilFee(in.Metal, comis)
		creditSilicon := in.Silicon - ceilFee(in.Silicon, comis)
		creditHydro := in.Hydrogen - ceilFee(in.Hydrogen, comis)
		if creditMetal < 0 {
			creditMetal = 0
		}
		if creditSilicon < 0 {
			creditSilicon = 0
		}
		if creditHydro < 0 {
			creditHydro = 0
		}
		if _, err := tx.Exec(ctx, `
			UPDATE fleets SET carried_metal=carried_metal-$1, carried_silicon=carried_silicon-$2,
			                  carried_hydrogen=carried_hydrogen-$3,
			                  control_times=control_times+1
			WHERE id=$4
		`, in.Metal, in.Silicon, in.Hydrogen, in.FleetID); err != nil {
			return fmt.Errorf("unload: charge fleet: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal=metal+$1, silicon=silicon+$2, hydrogen=hydrogen+$3 WHERE id=$4
		`, creditMetal, creditSilicon, creditHydro, in.CurrentPlanetID); err != nil {
			return fmt.Errorf("unload: credit planet: %w", err)
		}
		_, _ = tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'fleet_unload', $3, $4, $5)
		`, in.UserID, in.CurrentPlanetID, creditMetal, creditSilicon, creditHydro)
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
