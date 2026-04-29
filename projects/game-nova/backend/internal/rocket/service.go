// Package rocket — межпланетарные ракеты (legacy EVENT_ROCKET_ATTACK=16).
//
// В отличие от флотов, ракеты летят напрямую без возврата. Списываются
// из ships (unit_id=52) при запуске, при прибытии уничтожают defense
// на цели пропорционально урону.
//
// Упрощения M5:
//   - без anti_ballistic (перехвата на подлёте);
//   - без per-slot таргета — урон распределяется по всей defense-
//     таблице цели пропорционально текущему count × shell;
//   - детерминированно (без RNG-roll).
package rocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/galaxy"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// unitInterplanetary — id ракеты в ships (legacy UNIT_INTERPLANETARY_ROCKET).
const unitInterplanetary = 52

// buildingMissileSilo — id здания «Ракетная шахта» (legacy UNIT_ROCKET_STATION=53).
const buildingMissileSilo = 53

// kindRocketAttack — event.KindRocketAttack=16 (значение задаётся в
// event/kinds.go). Держим копию, чтобы не импортировать event
// из пакета rocket и избежать циклов, если они появятся.
const kindRocketAttack = 16

// missileDamage — урон одной ракеты. По legacy — 12000 (ogame classic).
const missileDamage = 12000

type Service struct {
	db          repo.Exec
	catalog     *config.Catalog
	speed       float64 // GAMESPEED
	numGalaxies int     // план 72.1 ч.12 — лимит из universes.yaml
	numSystems  int     // план 72.1 ч.12 — кольцевая топология систем
	bundle      *i18n.Bundle
}

func NewService(db repo.Exec, cat *config.Catalog, gameSpeed float64, numGalaxies, numSystems int) *Service {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	return &Service{db: db, catalog: cat, speed: gameSpeed, numGalaxies: numGalaxies, numSystems: numSystems}
}

func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
}

func (s *Service) tr(group, key string, vars map[string]string) string {
	if s.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return s.bundle.Tr(i18n.LangRu, group, key, vars)
}

var (
	ErrInvalidInput    = errors.New("rocket: invalid input")
	ErrPlanetOwnership = errors.New("rocket: source planet not owned by user")
	ErrNoRockets       = errors.New("rocket: not enough rockets on source planet")
	ErrTargetNotFound   = errors.New("rocket: target coords are empty")
	ErrTargetOnVacation = errors.New("rocket: target player is on vacation (protected)")
	ErrSiloLimit       = errors.New("rocket: count exceeds missile silo capacity")
)

// Launch — пуск `count` ракет с srcPlanetID на dst. Создаёт событие
// kind=16, которое при срабатывании наносит урон defense цели.
type LaunchResult struct {
	ImpactID string    `json:"impact_id"`
	Count    int64     `json:"count"`
	LaunchAt time.Time `json:"launch_at"`
	ImpactAt time.Time `json:"impact_at"`
}

// Launch пускает count ракет с srcPlanetID на dst. targetUnitID>0 — приоритетная
// цель (все выжившие ракеты бьют этот defense-стек первым; overflow → остальным).
func (s *Service) Launch(ctx context.Context, userID, srcPlanetID string,
	dst galaxy.Coords, count int64, targetUnitID int) (LaunchResult, error) {
	if count <= 0 {
		return LaunchResult{}, ErrInvalidInput
	}
	if err := dst.Validate(s.numGalaxies, s.numSystems); err != nil {
		return LaunchResult{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	var out LaunchResult
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Источник + координаты для расчёта времени.
		var (
			ownerID         string
			srcG, srcS, srcP int
		)
		err := tx.QueryRow(ctx, `
			SELECT user_id, galaxy, system, position
			FROM planets WHERE id = $1 AND destroyed_at IS NULL
			FOR UPDATE
		`, srcPlanetID).Scan(&ownerID, &srcG, &srcS, &srcP)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPlanetOwnership
			}
			return fmt.Errorf("rocket: read src: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
		}

		// Ракет хватает?
		var stock int64
		err = tx.QueryRow(ctx,
			`SELECT count FROM ships WHERE planet_id=$1 AND unit_id=$2 FOR UPDATE`,
			srcPlanetID, unitInterplanetary).Scan(&stock)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNoRockets
			}
			return fmt.Errorf("rocket: read stock: %w", err)
		}
		if stock < count {
			return ErrNoRockets
		}

		// Silo-limit: max_rockets = silo.level × capacity_per_level.
		var siloCapPerLevel int64 = 10 // legacy default
		for _, spec := range s.catalog.Buildings.Buildings {
			if spec.ID == buildingMissileSilo && spec.RocketCapacityPerLevel != nil {
				siloCapPerLevel = *spec.RocketCapacityPerLevel
				break
			}
		}
		var siloLevel int64
		_ = tx.QueryRow(ctx,
			`SELECT COALESCE(level, 0) FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			srcPlanetID, buildingMissileSilo).Scan(&siloLevel)
		maxRockets := siloLevel * siloCapPerLevel
		if maxRockets > 0 && count > maxRockets {
			return ErrSiloLimit
		}

		// Проверка цели (должна быть планета/луна) + vacation-щит (план 20 Ф.1).
		var exists, targetOnVacation bool
		err = tx.QueryRow(ctx, `
			SELECT
				EXISTS (
					SELECT 1 FROM planets
					WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
					  AND destroyed_at IS NULL
				),
				EXISTS (
					SELECT 1 FROM planets p
					JOIN users u ON u.id = p.user_id
					WHERE p.galaxy=$1 AND p.system=$2 AND p.position=$3 AND p.is_moon=$4
					  AND p.destroyed_at IS NULL AND u.vacation_since IS NOT NULL
				)
		`, dst.Galaxy, dst.System, dst.Position, dst.IsMoon).Scan(&exists, &targetOnVacation)
		if err != nil {
			return fmt.Errorf("rocket: check target: %w", err)
		}
		if !exists {
			return ErrTargetNotFound
		}
		if targetOnVacation {
			return ErrTargetOnVacation
		}

		// Списание ракет.
		if _, err := tx.Exec(ctx,
			`UPDATE ships SET count = count - $1 WHERE planet_id=$2 AND unit_id=$3`,
			count, srcPlanetID, unitInterplanetary); err != nil {
			return fmt.Errorf("rocket: charge: %w", err)
		}

		// Время полёта (ogame-like, упрощённо: 30 + 60*sqrt(dist/100)).
		dist := float64(galaxy.Distance(
			galaxy.Coords{Galaxy: srcG, System: srcS, Position: srcP}, dst, s.numSystems))
		secs := 30.0 + 60.0*math.Sqrt(dist/100.0)
		if s.speed > 0 {
			secs /= s.speed
		}
		if secs < 1 {
			secs = 1
		}
		launchAt := time.Now().UTC()
		impactAt := launchAt.Add(time.Duration(secs * float64(time.Second)))

		// Event payload: всё нужное handler'у, чтобы не читать
		// повторно планеты/флот.
		impactID := ids.New()
		payload, _ := json.Marshal(map[string]any{
			"impact_id":      impactID,
			"attacker_id":    userID,
			"src_planet":     srcPlanetID,
			"dst": map[string]any{
				"galaxy":   dst.Galaxy,
				"system":   dst.System,
				"position": dst.Position,
				"is_moon":  dst.IsMoon,
			},
			"count":          count,
			"target_unit_id": targetUnitID,
		})
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, 'wait', $4, $5)
		`, impactID, userID, kindRocketAttack, impactAt, payload); err != nil {
			return fmt.Errorf("rocket: insert event: %w", err)
		}

		out = LaunchResult{
			ImpactID: impactID,
			Count:    count,
			LaunchAt: launchAt,
			ImpactAt: impactAt,
		}
		return nil
	})
	return out, err
}

// Stock возвращает текущий запас ракет на планете (для UI).
func (s *Service) Stock(ctx context.Context, planetID string) (int64, error) {
	var n int64
	err := s.db.Pool().QueryRow(ctx,
		`SELECT COALESCE(count, 0) FROM ships WHERE planet_id=$1 AND unit_id=$2`,
		planetID, unitInterplanetary).Scan(&n)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("rocket: stock: %w", err)
	}
	return n, nil
}
