package fleet

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
)

// План 20 Ф.4: Сенсорная Фаланга (Star Surveillance, id=55, moon_only).
//
// Legacy (`MonitorPlanet.class.php`, `NS.class.php:1109`):
//   range = round((level^2 - 1) * (1 + hyperspace_tech / 10))
//
//   * level=1, hyperspace=0 → 0 (только своя система)
//   * level=3, hyperspace=5 → round(8 * 1.5) = 12 систем
//
// Стоимость скана: 5000H списывается с планеты-источника.
// Скан возвращает все fleet-события в целевой системе.
//
// Ограничения:
//   * источник — луна текущего игрока с star_surveillance >= 1
//   * та же галактика, |src_system - target_system| <= range
//   * хватает водорода на планете-источнике

// unitStarSurveillance — id здания (UNIT_STAR_SURVEILLANCE = 55).
const unitStarSurveillance = 55

// unitHyperspaceTech — id исследования.
const unitHyperspaceTech = 19

// starSurveillanceCost — водорода за 1 скан (STAR_SURVEILLANCE_CONSUMPTION).
const starSurveillanceCost int64 = 5000

// PhalanxScan — один флот, обнаруженный фалангой.
type PhalanxScan struct {
	FleetID     string    `json:"fleet_id"`
	OwnerID     string    `json:"owner_id"`
	Ownername   string    `json:"owner_name,omitempty"`
	Mission     int       `json:"mission"`
	State       string    `json:"state"` // outbound|returning
	SrcGalaxy   int       `json:"src_galaxy"`
	SrcSystem   int       `json:"src_system"`
	SrcPosition int       `json:"src_position"`
	DstGalaxy   int       `json:"dst_galaxy"`
	DstSystem   int       `json:"dst_system"`
	DstPosition int       `json:"dst_position"`
	DstIsMoon   bool      `json:"dst_is_moon"`
	DepartAt    time.Time `json:"depart_at"`
	ArriveAt    time.Time `json:"arrive_at"`
	ReturnAt    time.Time `json:"return_at"`
}

var (
	ErrPhalanxNotAMoon       = errors.New("phalanx: source must be a moon owned by you")
	ErrPhalanxNotInstalled   = errors.New("phalanx: source moon has no star_surveillance building")
	ErrPhalanxOutOfRange     = errors.New("phalanx: target out of scan range")
	ErrPhalanxNoHydrogen     = errors.New("phalanx: not enough hydrogen (5000 required)")
	ErrPhalanxDifferentGalax = errors.New("phalanx: target in different galaxy")
)

// Phalanx выполняет скан указанной системы. Возвращает список флотов,
// связанных с системой (src или dst = target). Самообслуживание
// (ваши собственные флоты) не отфильтровывается — показываем всё.
func (s *TransportService) Phalanx(ctx context.Context, userID, sourcePlanetID string,
	targetGalaxy, targetSystem int) ([]PhalanxScan, error) {
	var scans []PhalanxScan
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Читаем источник.
		var (
			srcG, srcS int
			isMoon     bool
			ownerID    string
			hydro      int64
			srcID      string
		)
		err := tx.QueryRow(ctx, `
			SELECT id, user_id, galaxy, system, is_moon, hydrogen
			FROM planets WHERE id=$1 AND destroyed_at IS NULL
			FOR UPDATE
		`, sourcePlanetID).Scan(&srcID, &ownerID, &srcG, &srcS, &isMoon, &hydro)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPhalanxNotAMoon
			}
			return fmt.Errorf("phalanx: read src: %w", err)
		}
		if ownerID != userID || !isMoon {
			return ErrPhalanxNotAMoon
		}
		if srcG != targetGalaxy {
			return ErrPhalanxDifferentGalax
		}

		// 2. Проверяем наличие здания.
		var buildingLevel int
		_ = tx.QueryRow(ctx, `
			SELECT COALESCE(level, 0) FROM buildings
			WHERE planet_id=$1 AND unit_id=$2
		`, srcID, unitStarSurveillance).Scan(&buildingLevel)
		if buildingLevel < 1 {
			return ErrPhalanxNotInstalled
		}

		// 3. Уровень hyperspace_tech.
		var hyperLvl int
		_ = tx.QueryRow(ctx, `
			SELECT COALESCE(level, 0) FROM research WHERE user_id=$1 AND unit_id=$2
		`, userID, unitHyperspaceTech).Scan(&hyperLvl)

		// 4. Radius = round((level^2 - 1) * (1 + hyper/10)).
		radius := int(math.Round(
			(math.Pow(float64(buildingLevel), 2) - 1) *
				(1 + float64(hyperLvl)/10.0),
		))
		if radius < 0 {
			radius = 0
		}
		delta := targetSystem - srcS
		if delta < 0 {
			delta = -delta
		}
		if delta > radius {
			return fmt.Errorf("%w (distance=%d, radius=%d)", ErrPhalanxOutOfRange, delta, radius)
		}

		// 5. Хватает водорода.
		if hydro < starSurveillanceCost {
			return ErrPhalanxNoHydrogen
		}
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET hydrogen = hydrogen - $1 WHERE id=$2`,
			starSurveillanceCost, srcID); err != nil {
			return fmt.Errorf("phalanx: charge hydrogen: %w", err)
		}

		// 6. Выбираем флоты, связанные с targetSystem.
		rows, err := tx.Query(ctx, `
			SELECT f.id, f.owner_user_id, COALESCE(u.username, ''),
			       f.mission, f.state,
			       ps.galaxy, ps.system, ps.position,
			       f.dst_galaxy, f.dst_system, f.dst_position, f.dst_is_moon,
			       f.depart_at, f.arrive_at, f.return_at
			FROM fleets f
			JOIN users u ON u.id = f.owner_user_id
			JOIN planets ps ON ps.id = f.src_planet_id
			WHERE f.state IN ('outbound', 'returning')
			  AND (
			    (f.dst_galaxy = $1 AND f.dst_system = $2)
			    OR (ps.galaxy = $1 AND ps.system = $2)
			  )
		`, targetGalaxy, targetSystem)
		if err != nil {
			return fmt.Errorf("phalanx: select fleets: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var sc PhalanxScan
			if err := rows.Scan(
				&sc.FleetID, &sc.OwnerID, &sc.Ownername,
				&sc.Mission, &sc.State,
				&sc.SrcGalaxy, &sc.SrcSystem, &sc.SrcPosition,
				&sc.DstGalaxy, &sc.DstSystem, &sc.DstPosition, &sc.DstIsMoon,
				&sc.DepartAt, &sc.ArriveAt, &sc.ReturnAt,
			); err != nil {
				return err
			}
			scans = append(scans, sc)
		}
		return rows.Err()
	})
	return scans, err
}
