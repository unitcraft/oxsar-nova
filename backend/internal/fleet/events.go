package fleet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/event"
)

// TransportArrivePayload — payload KindTransport=7. Совпадает с тем,
// что пишет transport.Send.
type transportPayload struct {
	FleetID string           `json:"fleet_id"`
	Carried map[string]int64 `json:"carried"`
}

// ArriveHandler — event.Handler для KindTransport. В точке прибытия:
//   - списываем ресурсы из fleets.carried в planets цели (если цель
//     существует; если уничтожена — обнуляем carry и ждём RETURN).
//   - state='returning'.
// Идемпотентно: повторный запуск видит уже state='returning' и ничего
// не делает.
func (s *TransportService) ArriveHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl transportPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("transport arrive: parse payload: %w", err)
		}
		var state string
		var g, sys, pos int
		var isMoon bool
		var cm, csil, ch int64
		err := tx.QueryRow(ctx, `
			SELECT state, dst_galaxy, dst_system, dst_position, dst_is_moon,
			       carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id = $1 FOR UPDATE
		`, pl.FleetID).Scan(&state, &g, &sys, &pos, &isMoon, &cm, &csil, &ch)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil // удалён — идемпотентность
			}
			return fmt.Errorf("select fleet: %w", err)
		}
		if state != "outbound" {
			return nil
		}

		// Цель по координатам.
		var planetID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM planets
			WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
			  AND destroyed_at IS NULL
		`, g, sys, pos, isMoon).Scan(&planetID)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("find target planet: %w", err)
		}
		if err == nil {
			// Передаём ресурсы.
			if _, err := tx.Exec(ctx, `
				UPDATE planets
				SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
				WHERE id = $4
			`, cm, csil, ch, planetID); err != nil {
				return fmt.Errorf("credit target: %w", err)
			}
			if _, err := tx.Exec(ctx, `
				UPDATE fleets SET carried_metal = 0, carried_silicon = 0, carried_hydrogen = 0,
				                  state = 'returning'
				WHERE id = $1
			`, pl.FleetID); err != nil {
				return fmt.Errorf("update fleet: %w", err)
			}
			return nil
		}
		// Цель исчезла — просто возвращаем груз домой (не обнуляем).
		if _, err := tx.Exec(ctx,
			`UPDATE fleets SET state = 'returning' WHERE id = $1`, pl.FleetID); err != nil {
			return fmt.Errorf("update fleet returning: %w", err)
		}
		return nil
	}
}

// RecyclingHandler — event.Handler для KindRecycling=9. Забирает
// metal+silicon из debris_fields на координатах флота (UPDATE с
// FOR UPDATE), зажимая свободным cargo. Остаток debris остаётся
// на орбите для следующих recycler'ов. state='returning'.
//
// Идемпотентно: повторный запуск видит state='returning' и
// возвращает nil.
func (s *TransportService) RecyclingHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl transportPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("recycling: parse payload: %w", err)
		}
		var (
			state            string
			g, sys, pos      int
			isMoon           bool
			cm, csil, ch     int64
		)
		err := tx.QueryRow(ctx, `
			SELECT state, dst_galaxy, dst_system, dst_position, dst_is_moon,
			       carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id = $1 FOR UPDATE
		`, pl.FleetID).Scan(&state, &g, &sys, &pos, &isMoon, &cm, &csil, &ch)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil
			}
			return fmt.Errorf("recycling: select fleet: %w", err)
		}
		if state != "outbound" {
			return nil
		}
		// Debris не привязан к is_moon (обломки над планетой и
		// луной считаются одним полем). Но для консистентности с
		// ATTACK-handler'ом, который пишет по (g,sys,pos) без
		// is_moon, читаем так же.
		var dMetal, dSilicon int64
		err = tx.QueryRow(ctx, `
			SELECT metal, silicon FROM debris_fields
			WHERE galaxy=$1 AND system=$2 AND position=$3
			FOR UPDATE
		`, g, sys, pos).Scan(&dMetal, &dSilicon)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("recycling: read debris: %w", err)
		}
		// Считаем cargo cap живых ship'ов во флоте.
		rows, err := tx.Query(ctx,
			`SELECT unit_id, count FROM fleet_ships WHERE fleet_id=$1`, pl.FleetID)
		if err != nil {
			return fmt.Errorf("recycling: read fleet_ships: %w", err)
		}
		var totalCap int64
		for rows.Next() {
			var uid int
			var cnt int64
			if err := rows.Scan(&uid, &cnt); err != nil {
				rows.Close()
				return err
			}
			for _, spec := range s.catalog.Ships.Ships {
				if spec.ID == uid {
					totalCap += spec.Cargo * cnt
					break
				}
			}
		}
		rows.Close()
		free := totalCap - (cm + csil + ch)
		if free < 0 {
			free = 0
		}
		// Забираем metal/silicon пропорционально, зажимая free.
		takeM, takeS := dMetal, dSilicon
		total := takeM + takeS
		if total > free && total > 0 {
			k := float64(free) / float64(total)
			takeM = int64(float64(takeM) * k)
			takeS = int64(float64(takeS) * k)
		}
		if takeM > 0 || takeS > 0 {
			if _, err := tx.Exec(ctx, `
				UPDATE debris_fields
				SET metal=metal-$1, silicon=silicon-$2, last_update=now()
				WHERE galaxy=$3 AND system=$4 AND position=$5
			`, takeM, takeS, g, sys, pos); err != nil {
				return fmt.Errorf("recycling: subtract debris: %w", err)
			}
			if _, err := tx.Exec(ctx, `
				UPDATE fleets SET carried_metal=$1, carried_silicon=$2
				WHERE id=$3
			`, cm+takeM, csil+takeS, pl.FleetID); err != nil {
				return fmt.Errorf("recycling: add carry: %w", err)
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID); err != nil {
			return fmt.Errorf("recycling: update state: %w", err)
		}
		return nil
	}
}

// ReturnHandler — event.Handler для KindReturn=20. Возвращает корабли
// в sток источника + карго (если остался) на источник. state='done'.
func (s *TransportService) ReturnHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl transportPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("transport return: parse payload: %w", err)
		}
		var state, srcPlanet string
		var cm, csil, ch int64
		err := tx.QueryRow(ctx, `
			SELECT state, src_planet_id, carried_metal, carried_silicon, carried_hydrogen
			FROM fleets WHERE id = $1 FOR UPDATE
		`, pl.FleetID).Scan(&state, &srcPlanet, &cm, &csil, &ch)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil
			}
			return fmt.Errorf("select fleet: %w", err)
		}
		if state == "done" {
			return nil
		}

		// Вернуть ресурсы (если что-то ещё в carry — это значит, цель
		// исчезла при arrive и груз пришёл домой).
		if cm > 0 || csil > 0 || ch > 0 {
			if _, err := tx.Exec(ctx, `
				UPDATE planets
				SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
				WHERE id = $4
			`, cm, csil, ch, srcPlanet); err != nil {
				return fmt.Errorf("refund carry: %w", err)
			}
		}
		// Вернуть корабли.
		rows, err := tx.Query(ctx,
			`SELECT unit_id, count FROM fleet_ships WHERE fleet_id = $1`, pl.FleetID)
		if err != nil {
			return fmt.Errorf("read fleet_ships: %w", err)
		}
		type ship struct {
			id    int
			count int64
		}
		var ships []ship
		for rows.Next() {
			var s ship
			if err := rows.Scan(&s.id, &s.count); err != nil {
				rows.Close()
				return err
			}
			ships = append(ships, s)
		}
		rows.Close()
		for _, ss := range ships {
			if _, err := tx.Exec(ctx, `
				INSERT INTO ships (planet_id, unit_id, count)
				VALUES ($1, $2, $3)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = ships.count + EXCLUDED.count
			`, srcPlanet, ss.id, ss.count); err != nil {
				return fmt.Errorf("return ship %d: %w", ss.id, err)
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE fleets SET state='done', carried_metal=0, carried_silicon=0, carried_hydrogen=0
			 WHERE id = $1`, pl.FleetID); err != nil {
			return fmt.Errorf("update fleet done: %w", err)
		}
		return nil
	}
}
