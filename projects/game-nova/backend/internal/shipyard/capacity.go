// План 72.1.41: legacy `Shipyard.class.php` строки 41-51, 470-490 —
// capacity checks для shield/rocket-юнитов. Игрок не может построить
// больше юнитов чем влезет в:
//   freeShieldFields = shield_tech × 10 - Σ(qty × slot_cost) - pending,
//   freeRocketFields = rocket_station × 15 - Σ(qty × slot_cost) - pending.
//
// Slot costs (legacy строки 49-50, 471-485):
//   small_shield (49)        = 1
//   large_shield (50)        = 5  (legacy *5 при подсчёте занятого, /2 при размещении? — нет, /5 эквивалент cap×2)
//   small_planet_shield (354)= 10
//   large_planet_shield (355)= 40
// Wait — legacy строки 471-477 используют /2, /4, /8 при размещении,
// но строки 49-50 — *5, *10, *40 при «занятых». Отличие в том, что
// legacy свёл /2 → 5/10 cost (large×5×slot, large_planet×10×4 etc).
// Не критично — ниже используем единый паттерн «slot per unit × qty».

package shipyard

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Legacy unit IDs (consts.php).
const (
	unitSmallShield        = 49
	unitLargeShield        = 50
	unitInterceptorRocket  = 51
	unitInterplanetaryRkt  = 52
	unitSmallPlanetShield  = 354
	unitLargePlanetShield  = 355

	unitShieldTech         = 16
	unitRocketStation      = 53
)

// Slot-cost: сколько «полей щита/ракет» занимает один юнит.
// Legacy строки 49-50 (для freeShieldFields подсчёта occupied) —
// 1, 5, 10, 40 для shield-юнитов; 1, 2 для rocket-юнитов.
var shieldSlotCost = map[int]int{
	unitSmallShield:       1,
	unitLargeShield:       5,
	unitSmallPlanetShield: 10,
	unitLargePlanetShield: 40,
}
var rocketSlotCost = map[int]int{
	unitInterceptorRocket: 1,
	unitInterplanetaryRkt: 2,
}

// ErrCapacityExceeded — попытка построить больше слотов чем доступно.
var ErrCapacityExceeded = errors.New("shipyard: shield/rocket capacity exceeded")

// CapacityInfo — UI-индикатор для DefenseScreen (legacy
// `freeShieldFields`, `freeRocketFields`).
type CapacityInfo struct {
	FreeShieldFields int64 `json:"free_shield_fields"`
	MaxShieldFields  int64 `json:"max_shield_fields"`
	FreeRocketFields int64 `json:"free_rocket_fields"`
	MaxRocketFields  int64 `json:"max_rocket_fields"`
}

// Capacity вычисляет shield/rocket capacity для UI. Использует pgx pool,
// не транзакция (read-only). Если research/building отсутствуют —
// max=0, free=0.
func (s *Service) Capacity(ctx context.Context, userID, planetID string) (CapacityInfo, error) {
	out := CapacityInfo{}
	pool := s.db.Pool()

	var techLvl int
	_ = pool.QueryRow(ctx,
		`SELECT COALESCE(level, 0) FROM research WHERE user_id=$1 AND unit_id=$2`,
		userID, unitShieldTech,
	).Scan(&techLvl)
	out.MaxShieldFields = int64(techLvl) * 10

	var stationLvl int
	_ = pool.QueryRow(ctx,
		`SELECT COALESCE(level, 0) FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		planetID, unitRocketStation,
	).Scan(&stationLvl)
	out.MaxRocketFields = int64(stationLvl) * 15

	// Deployed shield.
	var dep int64
	for uid, sc := range shieldSlotCost {
		var n int64
		_ = pool.QueryRow(ctx,
			`SELECT COALESCE(count, 0) FROM defense WHERE planet_id=$1 AND unit_id=$2`,
			planetID, uid,
		).Scan(&n)
		dep += n * int64(sc)
	}
	out.FreeShieldFields = out.MaxShieldFields - dep

	// Deployed rocket.
	dep = 0
	for uid, sc := range rocketSlotCost {
		var n int64
		_ = pool.QueryRow(ctx,
			`SELECT COALESCE(count, 0) FROM defense WHERE planet_id=$1 AND unit_id=$2`,
			planetID, uid,
		).Scan(&n)
		dep += n * int64(sc)
	}
	out.FreeRocketFields = out.MaxRocketFields - dep

	// Pending в shipyard_queue.
	rows, err := pool.Query(ctx, `
		SELECT unit_id, count FROM shipyard_queue
		WHERE planet_id=$1 AND status IN ('queued','running')
	`, planetID)
	if err == nil {
		for rows.Next() {
			var uid int
			var qty int64
			if err := rows.Scan(&uid, &qty); err == nil {
				if sc, ok := shieldSlotCost[uid]; ok {
					out.FreeShieldFields -= qty * int64(sc)
				}
				if sc, ok := rocketSlotCost[uid]; ok {
					out.FreeRocketFields -= qty * int64(sc)
				}
			}
		}
		rows.Close()
	}
	if out.FreeShieldFields < 0 {
		out.FreeShieldFields = 0
	}
	if out.FreeRocketFields < 0 {
		out.FreeRocketFields = 0
	}
	return out, nil
}

// checkCapacity — проверяет capacity для shield/rocket юнитов.
//
// Возвращает ErrCapacityExceeded если запрошенный count*slot превышает
// `freeFields`. Для не-shield/не-rocket юнитов — ничего не проверяет.
//
// Учитывает:
//   - текущее количество построенных юнитов на планете (defense table);
//   - pending в shipyard_queue (count × slot_cost) — для уже выставленных
//     задач same или discount-юнитов того же типа.
func (s *Service) checkCapacity(ctx context.Context, tx pgx.Tx,
	userID, planetID string, unitID int, count int64) error {
	var (
		isShield = false
		isRocket = false
		slotCost = 0
	)
	if c, ok := shieldSlotCost[unitID]; ok {
		isShield = true
		slotCost = c
	}
	if c, ok := rocketSlotCost[unitID]; ok {
		isRocket = true
		slotCost = c
	}
	if !isShield && !isRocket {
		return nil
	}

	// freeShieldFields = shield_tech × 10 - Σ(deployed × slot) - Σ(pending × slot).
	// freeRocketFields = rocket_station × 15 - Σ(deployed × slot) - Σ(pending × slot).
	var maxCap int
	if isShield {
		// shield_tech уровень — research у user.
		var techLvl int
		_ = tx.QueryRow(ctx,
			`SELECT COALESCE(level, 0) FROM research WHERE user_id=$1 AND unit_id=$2`,
			userID, unitShieldTech,
		).Scan(&techLvl)
		maxCap = techLvl * 10
	} else {
		var stationLvl int
		_ = tx.QueryRow(ctx,
			`SELECT COALESCE(level, 0) FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			planetID, unitRocketStation,
		).Scan(&stationLvl)
		maxCap = stationLvl * 15
	}
	if maxCap == 0 {
		// Без здания/research нельзя строить вообще.
		return ErrCapacityExceeded
	}

	// Σ deployed × slot (defense-таблица).
	deployed := int64(0)
	costMap := shieldSlotCost
	if isRocket {
		costMap = rocketSlotCost
	}
	for uid, sc := range costMap {
		var n int64
		_ = tx.QueryRow(ctx,
			`SELECT COALESCE(count, 0) FROM defense WHERE planet_id=$1 AND unit_id=$2`,
			planetID, uid,
		).Scan(&n)
		deployed += n * int64(sc)
	}

	// Σ pending × slot (shipyard_queue active).
	pending := int64(0)
	rows, err := tx.Query(ctx, `
		SELECT unit_id, count
		FROM shipyard_queue
		WHERE planet_id=$1 AND status IN ('queued','running')
	`, planetID)
	if err != nil {
		return fmt.Errorf("pending: %w", err)
	}
	for rows.Next() {
		var uid int
		var qty int64
		if err := rows.Scan(&uid, &qty); err != nil {
			rows.Close()
			return err
		}
		if sc, ok := costMap[uid]; ok {
			pending += qty * int64(sc)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	free := int64(maxCap) - deployed - pending
	need := count * int64(slotCost)
	if need > free {
		return ErrCapacityExceeded
	}
	return nil
}
