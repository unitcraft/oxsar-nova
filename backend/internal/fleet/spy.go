// SPY (mission=11) — разведка чужой планеты.
//
// Поток прибытия:
//  1. Читаем fleet, подсчитываем количество espionage_sensor (id=38).
//  2. Находим планету-цель. Если нет — state='returning' без отчёта.
//  3. ratio = probes + spy_self - spy_target (spyware-tech, id=13).
//  4. Собираем видимые секции:
//     * всегда: metal/silicon/hydrogen цели.
//     * ratio >= 2: ships.
//     * ratio >= 4: defense.
//     * ratio >= 6: buildings.
//  5. INSERT espionage_reports + message(folder=4) шпиону.
//     Цели — уведомление «вас шпионили ratio=N» (без деталей).
//  6. fleet.state='returning'.
//
// Counter-espionage: если цель имеет defense, часть зондов может быть
// уничтожена. Если все зонды сбиты — шпион получает уведомление
// «перехвачен» и флот уничтожается (без возврата).
// Research (ratio>=8) показывает уровни исследований цели.
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// unitEspionageSensor — id шпионских зондов (legacy UNIT_ESPIONAGE_SENSOR).
const unitEspionageSensor = 38

// unitSpywareTech — id spy-tech в research (legacy UNIT_SPYWARE).
const unitSpywareTech = 13

// SpyHandler — event.Handler для KindSpy=11.
func (s *TransportService) SpyHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl transportPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("spy: parse payload: %w", err)
		}
		var (
			state          string
			spyUserID      string
			g, sys, pos    int
			isMoon         bool
		)
		err := tx.QueryRow(ctx, `
			SELECT state, owner_user_id,
			       dst_galaxy, dst_system, dst_position, dst_is_moon
			FROM fleets WHERE id = $1 FOR UPDATE
		`, pl.FleetID).Scan(&state, &spyUserID, &g, &sys, &pos, &isMoon)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("spy: read fleet: %w", err)
		}
		if state != "outbound" {
			return nil
		}

		// Считаем probes в этом флоте. SPY-миссия нелегальна без
		// хотя бы одного probe'а, но если игрок прислал что-то ещё —
		// игнорируем (probe — единственный источник ratio).
		var probes int
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(count, 0) FROM fleet_ships WHERE fleet_id=$1 AND unit_id=$2`,
			pl.FleetID, unitEspionageSensor).Scan(&probes); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("spy: read probes: %w", err)
		}
		if probes <= 0 {
			// Нечем шпионить — просто возвращаемся.
			_, uerr := tx.Exec(ctx,
				`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID)
			return uerr
		}

		// Цель.
		var (
			planetID                   string
			targetUserID               string
			metal, silicon, hydrogen   float64
		)
		err = tx.QueryRow(ctx, `
			SELECT id, user_id, metal, silicon, hydrogen
			FROM planets
			WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
			  AND destroyed_at IS NULL
		`, g, sys, pos, isMoon).Scan(&planetID, &targetUserID, &metal, &silicon, &hydrogen)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_, uerr := tx.Exec(ctx,
					`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID)
				return uerr
			}
			return fmt.Errorf("spy: find target: %w", err)
		}

		// Counter-espionage: defense on target planet may intercept probes.
		// maxInterceptable = min(total_defense_count / 10, probes), rounded down.
		var defTotal int64
		_ = tx.QueryRow(ctx,
			`SELECT COALESCE(SUM(count), 0) FROM defense WHERE planet_id=$1`, planetID,
		).Scan(&defTotal)
		maxInterceptable := int(defTotal) / 10
		if maxInterceptable > probes {
			maxInterceptable = probes
		}
		if maxInterceptable > 0 {
			intercepted := rand.IntN(maxInterceptable + 1)
			probes -= intercepted
			if probes <= 0 {
				// All probes destroyed — fleet annihilated.
				notifSubj := fmt.Sprintf("Зонды перехвачены %d:%d:%d", g, sys, pos)
				notifBody := fmt.Sprintf("Все %d зонд(ов) уничтожены обороной цели. Флот потерян.", intercepted)
				_, _ = tx.Exec(ctx, `
					INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
					VALUES ($1, $2, $3, 4, $4, $5)
				`, ids.New(), spyUserID, targetUserID, notifSubj, notifBody)
				_, uerr := tx.Exec(ctx,
					`DELETE FROM fleets WHERE id=$1`, pl.FleetID)
				return uerr
			}
		}

		// Spy-tech обеих сторон.
		spySelf := readSpyLevel(ctx, tx, spyUserID)
		spyTarget := readSpyLevel(ctx, tx, targetUserID)
		ratio := probes + spySelf - spyTarget

		report := buildEspionageReport(ctx, tx, planetID, ratio, int64(metal), int64(silicon), int64(hydrogen), s.catalog)
		reportJSON, _ := json.Marshal(report)

		// INSERT report.
		reportID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO espionage_reports (id, spy_user_id, target_user_id, planet_id,
			                               ratio, probes, report)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, reportID, spyUserID, targetUserID, planetID, ratio, probes, reportJSON); err != nil {
			return fmt.Errorf("spy: insert report: %w", err)
		}

		// Сообщение шпиону с деталями.
		spySubj := fmt.Sprintf("Разведка %d:%d:%d (ratio=%d)", g, sys, pos, ratio)
		spyBody := fmt.Sprintf("Металл: %d, Кремний: %d, Водород: %d",
			int64(metal), int64(silicon), int64(hydrogen))
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, espionage_report_id)
			VALUES ($1, $2, $3, 4, $4, $5, $6)
		`, ids.New(), spyUserID, targetUserID, spySubj, spyBody, reportID); err != nil {
			return fmt.Errorf("spy: spy message: %w", err)
		}
		// Уведомление цели — без деталей, только факт попытки.
		if targetUserID != "" && targetUserID != spyUserID {
			tgtSubj := fmt.Sprintf("Вас шпионили %d:%d:%d", g, sys, pos)
			tgtBody := fmt.Sprintf("Противник послал %d зонд(ов). Соотношение %d.",
				probes, ratio)
			if _, err := tx.Exec(ctx, `
				INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
				VALUES ($1, $2, $3, 4, $4, $5)
			`, ids.New(), targetUserID, spyUserID, tgtSubj, tgtBody); err != nil {
				return fmt.Errorf("spy: target message: %w", err)
			}
		}

		if _, err := tx.Exec(ctx,
			`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID); err != nil {
			return fmt.Errorf("spy: update state: %w", err)
		}
		return nil
	}
}

// readSpyLevel — уровень spyware-технологии пользователя. 0 если нет.
func readSpyLevel(ctx context.Context, tx pgx.Tx, userID string) int {
	if userID == "" {
		return 0
	}
	var lvl int
	err := tx.QueryRow(ctx,
		`SELECT level FROM research WHERE user_id=$1 AND unit_id=$2`,
		userID, unitSpywareTech).Scan(&lvl)
	if err != nil {
		return 0
	}
	return lvl
}

type espionageReport struct {
	Ratio     int           `json:"ratio"`
	Probes    int           `json:"probes"`
	Metal     int64         `json:"metal"`
	Silicon   int64         `json:"silicon"`
	Hydrogen  int64         `json:"hydrogen"`
	Ships     map[int]int64 `json:"ships,omitempty"`
	Defense   map[int]int64 `json:"defense,omitempty"`
	Buildings map[int]int   `json:"buildings,omitempty"`
	Research  map[int]int   `json:"research,omitempty"`
}

// buildEspionageReport собирает report по ratio: ресурсы всегда,
// ships от 2, defense от 4, buildings от 6, research от 8.
func buildEspionageReport(ctx context.Context, tx pgx.Tx, planetID string,
	ratio int, m, s, h int64, _cat any) espionageReport {
	rep := espionageReport{
		Ratio:    ratio,
		Metal:    m,
		Silicon:  s,
		Hydrogen: h,
	}
	if ratio >= 2 {
		rep.Ships = readPlanetCounts(ctx, tx, "ships", planetID)
	}
	if ratio >= 4 {
		rep.Defense = readPlanetCounts(ctx, tx, "defense", planetID)
	}
	if ratio >= 6 {
		rep.Buildings = readPlanetLevels(ctx, tx, planetID)
	}
	if ratio >= 8 {
		rep.Research = readOwnerResearch(ctx, tx, planetID)
	}
	return rep
}

// readPlanetCounts — unit_id → count для таблиц ships/defense.
func readPlanetCounts(ctx context.Context, tx pgx.Tx, table, planetID string) map[int]int64 {
	out := map[int]int64{}
	rows, err := tx.Query(ctx,
		`SELECT unit_id, count FROM `+table+` WHERE planet_id=$1 AND count > 0`, planetID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var c int64
		if err := rows.Scan(&id, &c); err == nil {
			out[id] = c
		}
	}
	return out
}

func readPlanetLevels(ctx context.Context, tx pgx.Tx, planetID string) map[int]int {
	out := map[int]int{}
	rows, err := tx.Query(ctx,
		`SELECT unit_id, level FROM buildings WHERE planet_id=$1 AND level > 0`, planetID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, lvl int
		if err := rows.Scan(&id, &lvl); err == nil {
			out[id] = lvl
		}
	}
	return out
}

// readOwnerResearch — research levels for the owner of planetID.
func readOwnerResearch(ctx context.Context, tx pgx.Tx, planetID string) map[int]int {
	out := map[int]int{}
	rows, err := tx.Query(ctx, `
		SELECT r.unit_id, r.level
		FROM research r
		JOIN planets p ON p.user_id = r.user_id
		WHERE p.id = $1 AND r.level > 0
	`, planetID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, lvl int
		if err := rows.Scan(&id, &lvl); err == nil {
			out[id] = lvl
		}
	}
	return out
}
