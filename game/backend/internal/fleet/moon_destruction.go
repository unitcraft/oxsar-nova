package fleet

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/pkg/ids"
	"github.com/oxsar/nova/backend/pkg/rng"
)

// fleetInfo и survivorFleet — package-level типы, вынесены из
// acs_attack.go, чтобы tryDestroyMoonACS мог принимать их в параметрах.
type fleetInfo struct {
	id          string
	ownerUserID string
	cm, cs, ch  int64
	isMoon      bool
	g, sys, pos int
}

type survivorFleet struct {
	info      fleetInfo
	survivors []unitStack
}

// План 20 Ф.6: Moon Destruction.
//
// Атаки kind=25 (single) и kind=27 (alliance) — после обычного боя
// делается дополнительный roll на разрушение луны и уничтожение DS:
//
//   P_destroy_moon  = clamp((100 - sqrt(diameter)) * sqrt(rip_count), 0, 100)
//   P_destroy_fleet = clamp((100 - sqrt(rip_count)) * sqrt(diameter) / 200, 0, 100)
//
// rip_count — выжившие Death Star у атакующих. Если DS не выжило —
// roll не делается.
//
// Если луна уничтожена — planets.destroyed_at=now(). Сообщения обоим
// игрокам (folder=2 — отчёты).
//
// Если флот уничтожен — все DS у атакующих удаляются из fleet_ships.

// rng-seed для детерминированности воспроизведения отчёта боя.
// Используем seed из report (тот же seed, что и у battle.Calculate).
func (s *TransportService) tryDestroyMoon(ctx context.Context, tx pgx.Tx,
	fleetID, moonPlanetID, attackerUserID, defenderUserID string,
	atkSurvivors []unitStack, battleSeed uint64) error {

	// 1. Считаем DS среди выживших атакующих.
	var ripCount int64
	for _, u := range atkSurvivors {
		if u.UnitID == unitDeathstar {
			ripCount = u.Count
			break
		}
	}
	if ripCount <= 0 {
		// Нет DS — нет roll.
		return nil
	}

	// 2. Diameter луны.
	var diameter int
	err := tx.QueryRow(ctx,
		`SELECT diameter FROM planets WHERE id=$1`, moonPlanetID).Scan(&diameter)
	if err != nil {
		return fmt.Errorf("read moon diameter: %w", err)
	}
	if diameter <= 0 {
		// Странно — диаметр нулевой; пропускаем чтобы не получить деление на ноль.
		return nil
	}

	// 3. Формулы.
	pMoon := (100.0 - math.Sqrt(float64(diameter))) * math.Sqrt(float64(ripCount))
	if pMoon < 0 {
		pMoon = 0
	} else if pMoon > 100 {
		pMoon = 100
	}
	pFleet := (100.0 - math.Sqrt(float64(ripCount))) * math.Sqrt(float64(diameter)) / 200.0
	if pFleet < 0 {
		pFleet = 0
	} else if pFleet > 100 {
		pFleet = 100
	}

	// 4. Roll. Используем seed от боя, чтобы reproducible.
	r := rng.New(battleSeed ^ 0xDEADBEEFCAFEBABE)
	moonRoll := r.Float64() * 100
	fleetRoll := r.Float64() * 100

	moonDestroyed := moonRoll < pMoon
	fleetDestroyed := fleetRoll < pFleet

	// 5. Если луна уничтожена — destroyed_at + сообщение.
	pct := func(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }
	cnt := strconv.FormatInt(int64(ripCount), 10)

	if moonDestroyed {
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET destroyed_at = now() WHERE id=$1 AND destroyed_at IS NULL`,
			moonPlanetID); err != nil {
			return fmt.Errorf("mark moon destroyed: %w", err)
		}
		vars := map[string]string{"count": cnt, "percent": pct(pMoon), "roll": pct(moonRoll)}
		if err := sendMoonMessage(ctx, tx, defenderUserID,
			s.tr("assaultReport", "moonDestroyedSubject", nil),
			s.tr("assaultReport", "moonDestroyedBody", vars)); err != nil {
			return err
		}
		if err := sendMoonMessage(ctx, tx, attackerUserID,
			s.tr("assaultReport", "enemyMoonDestroyedSubject", nil),
			s.tr("assaultReport", "enemyMoonDestroyedBody", vars)); err != nil {
			return err
		}
	}
	// 6. Если флот уничтожен — удаляем DS у атакующих из fleet_ships.
	if fleetDestroyed {
		if _, err := tx.Exec(ctx,
			`DELETE FROM fleet_ships WHERE fleet_id=$1 AND unit_id=$2`,
			fleetID, unitDeathstar); err != nil {
			return fmt.Errorf("wipe DS from fleet: %w", err)
		}
		fleetVars := map[string]string{"percent": pct(pFleet), "roll": pct(fleetRoll)}
		if err := sendMoonMessage(ctx, tx, attackerUserID,
			s.tr("assaultReport", "deathstarLostSubject", nil),
			s.tr("assaultReport", "deathstarLostBody", fleetVars)); err != nil {
			return err
		}
	}
	return nil
}

// tryDestroyMoonACS — ACS-вариант (kind=27). rip_count = сумма выживших
// DS по всем участникам ACS. При успехе fleet_destroy DS удаляются у
// всех участников пропорционально их вкладу.
func (s *TransportService) tryDestroyMoonACS(ctx context.Context, tx pgx.Tx,
	moonPlanetID, leadUserID, defenderUserID string,
	survivingFleets []survivorFleet, ripTotal int64, battleSeed uint64) error {
	if ripTotal <= 0 {
		return nil
	}

	// Diameter луны.
	var diameter int
	if err := tx.QueryRow(ctx,
		`SELECT diameter FROM planets WHERE id=$1`, moonPlanetID).Scan(&diameter); err != nil {
		return fmt.Errorf("acs moon: read diameter: %w", err)
	}
	if diameter <= 0 {
		return nil
	}

	pMoon := (100.0 - math.Sqrt(float64(diameter))) * math.Sqrt(float64(ripTotal))
	if pMoon < 0 {
		pMoon = 0
	} else if pMoon > 100 {
		pMoon = 100
	}
	pFleet := (100.0 - math.Sqrt(float64(ripTotal))) * math.Sqrt(float64(diameter)) / 200.0
	if pFleet < 0 {
		pFleet = 0
	} else if pFleet > 100 {
		pFleet = 100
	}

	r := rng.New(battleSeed ^ 0xDEADBEEFCAFEBABE)
	moonRoll := r.Float64() * 100
	fleetRoll := r.Float64() * 100

	moonDestroyed := moonRoll < pMoon
	fleetDestroyed := fleetRoll < pFleet

	pct := func(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }
	cnt := strconv.FormatInt(ripTotal, 10)

	if moonDestroyed {
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET destroyed_at = now() WHERE id=$1 AND destroyed_at IS NULL`,
			moonPlanetID); err != nil {
			return fmt.Errorf("acs mark moon destroyed: %w", err)
		}
		moonVars := map[string]string{"count": cnt, "percent": pct(pMoon), "roll": pct(moonRoll)}
		_ = sendMoonMessage(ctx, tx, defenderUserID,
			s.tr("assaultReport", "acsMoonDestroyedSubject", nil),
			s.tr("assaultReport", "acsMoonDestroyedBody", moonVars))
		_ = sendMoonMessage(ctx, tx, leadUserID,
			s.tr("assaultReport", "acsEnemyMoonDestroyedSubject", nil),
			s.tr("assaultReport", "acsEnemyMoonDestroyedBody", moonVars))
	}
	if fleetDestroyed {
		fleetVars := map[string]string{"count": cnt, "percent": pct(pFleet), "roll": pct(fleetRoll)}
		// Удаляем DS у каждого участника ACS, у кого они выжили.
		for _, sf := range survivingFleets {
			hasDS := false
			for _, st := range sf.survivors {
				if st.UnitID == unitDeathstar && st.Count > 0 {
					hasDS = true
					break
				}
			}
			if !hasDS {
				continue
			}
			if _, err := tx.Exec(ctx,
				`DELETE FROM fleet_ships WHERE fleet_id=$1 AND unit_id=$2`,
				sf.info.id, unitDeathstar); err != nil {
				return fmt.Errorf("acs wipe DS: %w", err)
			}
			_ = sendMoonMessage(ctx, tx, sf.info.ownerUserID,
				s.tr("assaultReport", "acsDeathstarLostSubject", nil),
				s.tr("assaultReport", "acsDeathstarLostBody", fleetVars))
		}
	}
	return nil
}

func sendMoonMessage(ctx context.Context, tx pgx.Tx,
	toUserID, subject, body string) error {
	if toUserID == "" {
		return nil
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, 2, $3, $4)
	`, ids.New(), toUserID, subject, body)
	if err != nil {
		return fmt.Errorf("send moon msg: %w", err)
	}
	return nil
}
