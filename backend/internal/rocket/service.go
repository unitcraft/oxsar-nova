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

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/galaxy"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// unitInterplanetary — id ракеты в ships (legacy UNIT_INTERPLANETARY_ROCKET).
const unitInterplanetary = 52

// kindRocketAttack — event.KindRocketAttack=16 (значение задаётся в
// event/kinds.go). Держим копию, чтобы не импортировать event
// из пакета rocket и избежать циклов, если они появятся.
const kindRocketAttack = 16

// missileDamage — урон одной ракеты. По legacy — 12000 (ogame classic).
const missileDamage = 12000

type Service struct {
	db      repo.Exec
	catalog *config.Catalog
	speed   float64 // GAMESPEED
}

func NewService(db repo.Exec, cat *config.Catalog, gameSpeed float64) *Service {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	return &Service{db: db, catalog: cat, speed: gameSpeed}
}

var (
	ErrInvalidInput      = errors.New("rocket: invalid input")
	ErrPlanetOwnership   = errors.New("rocket: source planet not owned by user")
	ErrNoRockets         = errors.New("rocket: not enough rockets on source planet")
	ErrTargetNotFound    = errors.New("rocket: target coords are empty")
)

// Launch — пуск `count` ракет с srcPlanetID на dst. Создаёт событие
// kind=16, которое при срабатывании наносит урон defense цели.
type LaunchResult struct {
	ImpactID string    `json:"impact_id"`
	Count    int64     `json:"count"`
	LaunchAt time.Time `json:"launch_at"`
	ImpactAt time.Time `json:"impact_at"`
}

func (s *Service) Launch(ctx context.Context, userID, srcPlanetID string,
	dst galaxy.Coords, count int64) (LaunchResult, error) {
	if count <= 0 {
		return LaunchResult{}, ErrInvalidInput
	}
	if err := dst.Validate(); err != nil {
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

		// Проверка цели (должна быть планета/луна).
		var exists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM planets
				WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
				  AND destroyed_at IS NULL
			)
		`, dst.Galaxy, dst.System, dst.Position, dst.IsMoon).Scan(&exists)
		if err != nil {
			return fmt.Errorf("rocket: check target: %w", err)
		}
		if !exists {
			return ErrTargetNotFound
		}

		// Списание ракет.
		if _, err := tx.Exec(ctx,
			`UPDATE ships SET count = count - $1 WHERE planet_id=$2 AND unit_id=$3`,
			count, srcPlanetID, unitInterplanetary); err != nil {
			return fmt.Errorf("rocket: charge: %w", err)
		}

		// Время полёта (ogame-like, упрощённо: 30 + 60*sqrt(dist/100)).
		dist := float64(galaxy.Distance(
			galaxy.Coords{Galaxy: srcG, System: srcS, Position: srcP}, dst))
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
			"impact_id":   impactID,
			"attacker_id": userID,
			"src_planet":  srcPlanetID,
			"dst": map[string]any{
				"galaxy":   dst.Galaxy,
				"system":   dst.System,
				"position": dst.Position,
				"is_moon":  dst.IsMoon,
			},
			"count": count,
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
