package fleet

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/galaxy"
)

// План 20 Ф.5: Stargate Jump — мгновенный прыжок флота между лунами.
//
// POST /api/stargate {src_planet_id, dst_planet_id, ships{unit_id:count}}
//
// Условия:
//   * src и dst — луны игрока с jump_gate >= 1
//   * cooldown на src: 3600 * 0.7^(jump_gate_level - 1) секунд
//   * запрещённые юниты: щиты (49,50,354,355), ракеты (51,52)
//   * ограничения по позиции: 1-2 нельзя, 3 только на луны (у нас все
//     stargate moon_only — это ограничение фактически означает «нельзя
//     с планеты позиции <=2»). Pos>=3 → можно.
//
// Поведение: ships перемещаются напрямую src.ships → dst.ships
// (без полёта). last_jump_at записывается в stargate_cooldowns.

const unitJumpGate = 56 // star_gate в buildings.yml

// Запрещённые юниты для прыжка.
var stargateBannedUnits = map[int]bool{
	49: true,  // small_shield
	50: true,  // large_shield
	51:  true, // interceptor_rocket
	52:  true, // interplanetary_rocket
	354: true, // small_planet_shield
	355: true, // large_planet_shield
}

var (
	ErrStargateNotMoon       = errors.New("stargate: source and destination must be moons")
	ErrStargateNotInstalled  = errors.New("stargate: jump_gate not installed on one of the moons")
	ErrStargateCooldown      = errors.New("stargate: cooldown active, try later")
	ErrStargateBannedUnit    = errors.New("stargate: shields and rockets cannot jump")
	ErrStargatePositionLimit = errors.New("stargate: source position too low (need >= 3)")
	ErrStargateNotEnoughShip = errors.New("stargate: not enough ships on source")
	ErrStargateNotOwner      = errors.New("stargate: source not owned by you")
)

// StargateJumpInput — параметры прыжка.
type StargateJumpInput struct {
	UserID      string
	SrcPlanetID string
	DstPlanetID string
	Ships       map[int]int64 // unit_id -> count
}

// StargateJumpResult — результат прыжка.
type StargateJumpResult struct {
	JumpedAt    time.Time `json:"jumped_at"`
	NextJumpAt  time.Time `json:"next_jump_at"`
	CooldownSec int       `json:"cooldown_sec"`
	Ships       map[int]int64 `json:"ships"`
}

// StargateJump — выполнить прыжок.
func (s *TransportService) StargateJump(ctx context.Context, in StargateJumpInput) (StargateJumpResult, error) {
	if in.SrcPlanetID == "" || in.DstPlanetID == "" {
		return StargateJumpResult{}, fmt.Errorf("%w: src and dst required", ErrInvalidDispatch)
	}
	if in.SrcPlanetID == in.DstPlanetID {
		return StargateJumpResult{}, fmt.Errorf("%w: src == dst", ErrInvalidDispatch)
	}
	if len(in.Ships) == 0 {
		return StargateJumpResult{}, fmt.Errorf("%w: no ships selected", ErrInvalidDispatch)
	}
	for unitID := range in.Ships {
		if stargateBannedUnits[unitID] {
			return StargateJumpResult{}, fmt.Errorf("%w: unit_id=%d", ErrStargateBannedUnit, unitID)
		}
	}

	var out StargateJumpResult
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Источник — наш + луна + position >= 3.
		var (
			srcOwner string
			srcMoon  bool
			srcPos   int
		)
		err := tx.QueryRow(ctx, `
			SELECT user_id, is_moon, position
			FROM planets WHERE id=$1 AND destroyed_at IS NULL
			FOR UPDATE
		`, in.SrcPlanetID).Scan(&srcOwner, &srcMoon, &srcPos)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTargetNotFound
			}
			return fmt.Errorf("stargate: read src: %w", err)
		}
		if srcOwner != in.UserID {
			return ErrStargateNotOwner
		}
		if !srcMoon {
			return ErrStargateNotMoon
		}
		if srcPos < 3 {
			return ErrStargatePositionLimit
		}
		// 2. Цель — наша или союзника, луна. (Используем существующую
		//    проверку POSITION — она именно про это.)
		dst, err := readPlanetInfo(ctx, tx, in.DstPlanetID)
		if err != nil {
			return err
		}
		if !dst.isMoon {
			return ErrStargateNotMoon
		}
		if err := s.checkPositionTarget(ctx, tx, in.UserID, dst.coords); err != nil {
			return err
		}
		// 3. Здание jump_gate >= 1 на src и dst.
		var srcGate, dstGate int
		_ = tx.QueryRow(ctx,
			`SELECT COALESCE(level, 0) FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			in.SrcPlanetID, unitJumpGate).Scan(&srcGate)
		_ = tx.QueryRow(ctx,
			`SELECT COALESCE(level, 0) FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			in.DstPlanetID, unitJumpGate).Scan(&dstGate)
		if srcGate < 1 || dstGate < 1 {
			return ErrStargateNotInstalled
		}
		// 4. Cooldown.
		cooldownSec := int(math.Round(3600 * math.Pow(0.7, float64(srcGate-1))))
		var lastJumpAt *time.Time
		_ = tx.QueryRow(ctx,
			`SELECT last_jump_at FROM stargate_cooldowns WHERE planet_id=$1`,
			in.SrcPlanetID).Scan(&lastJumpAt)
		now := time.Now().UTC()
		if lastJumpAt != nil && now.Sub(*lastJumpAt) < time.Duration(cooldownSec)*time.Second {
			return fmt.Errorf("%w (next jump at %s)",
				ErrStargateCooldown,
				lastJumpAt.Add(time.Duration(cooldownSec)*time.Second).Format(time.RFC3339))
		}
		// 5. Хватает кораблей.
		for unitID, qty := range in.Ships {
			if qty <= 0 {
				continue
			}
			var have int64
			_ = tx.QueryRow(ctx,
				`SELECT count FROM ships WHERE planet_id=$1 AND unit_id=$2 FOR UPDATE`,
				in.SrcPlanetID, unitID).Scan(&have)
			if have < qty {
				return fmt.Errorf("%w: unit_id=%d (%d < %d)",
					ErrStargateNotEnoughShip, unitID, have, qty)
			}
		}
		// 6. Списать с src, добавить на dst.
		for unitID, qty := range in.Ships {
			if qty <= 0 {
				continue
			}
			if _, err := tx.Exec(ctx,
				`UPDATE ships SET count = count - $1 WHERE planet_id=$2 AND unit_id=$3`,
				qty, in.SrcPlanetID, unitID); err != nil {
				return fmt.Errorf("stargate: charge: %w", err)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO ships (planet_id, unit_id, count)
				VALUES ($1, $2, $3)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = ships.count + EXCLUDED.count
			`, in.DstPlanetID, unitID, qty); err != nil {
				return fmt.Errorf("stargate: credit: %w", err)
			}
		}
		// Очищаем строки с count=0, чтобы не накапливались.
		_, _ = tx.Exec(ctx,
			`DELETE FROM ships WHERE planet_id=$1 AND count <= 0`, in.SrcPlanetID)
		// 7. Записать cooldown.
		if _, err := tx.Exec(ctx, `
			INSERT INTO stargate_cooldowns (planet_id, last_jump_at)
			VALUES ($1, $2)
			ON CONFLICT (planet_id) DO UPDATE SET last_jump_at = EXCLUDED.last_jump_at
		`, in.SrcPlanetID, now); err != nil {
			return fmt.Errorf("stargate: cooldown insert: %w", err)
		}
		out = StargateJumpResult{
			JumpedAt:    now,
			NextJumpAt:  now.Add(time.Duration(cooldownSec) * time.Second),
			CooldownSec: cooldownSec,
			Ships:       in.Ships,
		}
		return nil
	})
	return out, err
}

// readPlanetInfo — мини-helper для проверок цели.
type planetInfo struct {
	owner  string
	isMoon bool
	coords galaxy.Coords
}

func readPlanetInfo(ctx context.Context, tx pgx.Tx, planetID string) (*planetInfo, error) {
	var (
		owner       string
		isMoon      bool
		g, sys, pos int
	)
	err := tx.QueryRow(ctx, `
		SELECT user_id, is_moon, galaxy, system, position
		FROM planets WHERE id=$1 AND destroyed_at IS NULL
	`, planetID).Scan(&owner, &isMoon, &g, &sys, &pos)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTargetNotFound
		}
		return nil, fmt.Errorf("read planet: %w", err)
	}
	return &planetInfo{
		owner:  owner,
		isMoon: isMoon,
		coords: galaxy.Coords{Galaxy: g, System: sys, Position: pos, IsMoon: isMoon},
	}, nil
}
