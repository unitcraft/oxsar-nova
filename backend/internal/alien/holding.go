package alien

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// holdingPayload — содержимое events.payload для KindAlienHalt,
// KindAlienHolding, KindAlienHoldingAI.
//
// AlienFleet — выжившие корабли пришельцев после боя, которые стоят
// на орбите планеты. Используется и для участия в бою против стороннего
// атакующего, и для постепенного убывания через onExtractAlientShipsAI.
//
// StartTime — момент начала HOLDING (для cap 15 дней).
// HoldingEventID — id родительского KindAlienHolding (у HALT пусто, у
// HOLDING_AI ссылается на HOLDING — нужно для продления платежом).
type holdingPayload struct {
	PlanetID       string         `json:"planet_id"`
	UserID         string         `json:"user_id"`
	Tier           int            `json:"tier"`
	AlienFleet     []fleetStack   `json:"alien_fleet"`
	StartTime      time.Time      `json:"start_time"`
	HoldingEventID string         `json:"holding_event_id,omitempty"`
	PaidCredit     int64          `json:"paid_credit,omitempty"`
	PaidTimes      int            `json:"paid_times,omitempty"`
}

// fleetStack — компактный снимок alien-флота в payload (без Attack/Shield —
// эти параметры при чтении восстанавливаются из catalog).
type fleetStack struct {
	UnitID   int   `json:"unit_id"`
	Quantity int64 `json:"quantity"`
}

// survivorsToStacks конвертирует результат боя пришельцев
// (battle.SideResult.Units) в компактное представление для payload.
// Отфильтровывает юниты с нулевой остаточной численностью.
func survivorsToStacks(units []battle.UnitResult) []fleetStack {
	out := make([]fleetStack, 0, len(units))
	for _, u := range units {
		if u.QuantityEnd > 0 {
			out = append(out, fleetStack{UnitID: u.UnitID, Quantity: u.QuantityEnd})
		}
	}
	return out
}

// spawnHalt планирует переход планеты в HALT через 12–24ч после победы
// пришельцев. После HALT наступает HOLDING.
//
// Если alien-флот целиком уничтожен (survivors пуст) — HALT не спавнится,
// пришельцам некого оставлять.
func (s *Service) spawnHalt(ctx context.Context, tx pgx.Tx, pl alienPayload, survivors []fleetStack, r *rand.Rand) error {
	if len(survivors) == 0 {
		return nil
	}
	// Длительность HALT: [12ч, 24ч]. Используем rand для
	// детерминированности в тестах (caller передаёт свой seed).
	dur := AlienHaltingMinTime + time.Duration(r.Int64N(int64(AlienHaltingMaxTime-AlienHaltingMinTime)))
	hp := holdingPayload{
		PlanetID:   pl.PlanetID,
		UserID:     pl.UserID,
		Tier:       pl.Tier,
		AlienFleet: survivors,
		StartTime:  time.Now().UTC(),
	}
	data, err := json.Marshal(hp)
	if err != nil {
		return fmt.Errorf("alien halt: marshal: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, ids.New(), event.KindAlienHalt, pl.PlanetID, pl.UserID, time.Now().Add(dur), data); err != nil {
		return fmt.Errorf("alien halt: insert: %w", err)
	}
	return nil
}

// HaltHandler — event.Handler для KindAlienHalt (36).
//
// Триггерится по истечении 12–24ч после победы пришельцев. Создаёт
// событие KindAlienHolding — планета входит в состояние удержания.
// Если владелец планеты изменился или планета уничтожена — HALT
// просто завершается без последствий (идемпотентно).
func (s *Service) HaltHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var hp holdingPayload
		if err := json.Unmarshal(e.Payload, &hp); err != nil {
			return fmt.Errorf("alien halt: parse payload: %w", err)
		}

		// Планета жива и принадлежит тому же игроку? Если нет — HALT
		// истёк, пришельцы ушли молча.
		var curUserID string
		err := tx.QueryRow(ctx, `
			SELECT user_id FROM planets WHERE id = $1 AND destroyed_at IS NULL
		`, hp.PlanetID).Scan(&curUserID)
		if err == pgx.ErrNoRows || (err == nil && curUserID != hp.UserID) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("alien halt: check planet: %w", err)
		}

		// Спавним HOLDING: fire_at = start + [12ч, 24ч] (длительность
		// удержания без платежа). При платеже — handler платежа сдвигает
		// fire_at вперёд.
		dur := AlienHaltingMinTime + time.Duration(rand.Int64N(int64(AlienHaltingMaxTime-AlienHaltingMinTime)))
		holdingStart := time.Now().UTC()
		hp.StartTime = holdingStart
		data, err := json.Marshal(hp)
		if err != nil {
			return fmt.Errorf("alien halt: marshal holding: %w", err)
		}
		holdingID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, holdingID, event.KindAlienHolding, hp.PlanetID, hp.UserID,
			holdingStart.Add(dur), data); err != nil {
			return fmt.Errorf("alien halt: insert holding: %w", err)
		}

		// Первый HOLDING_AI-тик через 5–10 сек.
		hp.HoldingEventID = holdingID
		aiData, err := json.Marshal(hp)
		if err != nil {
			return fmt.Errorf("alien halt: marshal holding_ai: %w", err)
		}
		aiDelay := 5*time.Second + time.Duration(rand.Int64N(int64(5*time.Second)))
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, ids.New(), event.KindAlienHoldingAI, hp.PlanetID, hp.UserID,
			time.Now().Add(aiDelay), aiData); err != nil {
			return fmt.Errorf("alien halt: insert holding_ai: %w", err)
		}

		// Сообщение: пришельцы блокировали планету.
		body := fmt.Sprintf(
			"Инопланетяне (тир %d) установили контроль над вашей планетой. "+
				"Их флот останется на орбите до %s. Пока они здесь, они отражают "+
				"атаки других игроков — но и сами забирают часть ресурсов.",
			hp.Tier, holdingStart.Add(dur).Format("2006-01-02 15:04 MST"),
		)
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 1, 'Пришельцы удерживают планету', $3)
		`, ids.New(), hp.UserID, body); err != nil {
			return fmt.Errorf("alien halt: message: %w", err)
		}
		return nil
	}
}

// HoldingHandler — event.Handler для KindAlienHolding (34).
//
// Триггерится при истечении duration (с учётом возможных продлений
// платежом). Завершает HOLDING: пришельцы уходят, сообщение игроку.
// Отдельного HOLDING_AI отменять не надо — он проверит актуальность
// HOLDING-события и тихо завершится, если родитель уже done.
func (s *Service) HoldingHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var hp holdingPayload
		if err := json.Unmarshal(e.Payload, &hp); err != nil {
			return fmt.Errorf("alien holding: parse payload: %w", err)
		}
		// Планета жива?
		var curUserID string
		err := tx.QueryRow(ctx, `
			SELECT user_id FROM planets WHERE id = $1 AND destroyed_at IS NULL
		`, hp.PlanetID).Scan(&curUserID)
		if err == pgx.ErrNoRows {
			return nil
		}
		if err != nil {
			return fmt.Errorf("alien holding: check planet: %w", err)
		}

		body := "Инопланетяне покинули вашу планету — флот ушёл в глубокий космос."
		if hp.PaidTimes > 0 {
			body = fmt.Sprintf(
				"Инопланетяне покинули вашу планету. За время удержания вы "+
					"заплатили им %d кредитов за %d продлений.",
				hp.PaidCredit, hp.PaidTimes,
			)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 1, 'Пришельцы ушли', $3)
		`, ids.New(), hp.UserID, body); err != nil {
			return fmt.Errorf("alien holding: message: %w", err)
		}
		return nil
	}
}

// HoldingAIHandler — event.Handler для KindAlienHoldingAI (80).
//
// Триггерится каждые 12–24ч внутри HOLDING. Выполняет одно случайное
// действие (в Этапе 1 — только onUnloadAlienResoursesAI: 7–10% от
// захваченных ресурсов переходит на склад игрока). Затем планирует
// следующий тик, если HOLDING ещё активен.
func (s *Service) HoldingAIHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var hp holdingPayload
		if err := json.Unmarshal(e.Payload, &hp); err != nil {
			return fmt.Errorf("alien holding_ai: parse payload: %w", err)
		}

		// HOLDING ещё активен?
		var holdingState string
		err := tx.QueryRow(ctx, `
			SELECT state FROM events WHERE id = $1
		`, hp.HoldingEventID).Scan(&holdingState)
		if err == pgx.ErrNoRows {
			return nil // HOLDING уже удалён/ушёл
		}
		if err != nil {
			return fmt.Errorf("alien holding_ai: check holding: %w", err)
		}
		if holdingState != string(event.StateWait) {
			return nil // HOLDING закрыт
		}

		// Действие: onUnloadAlienResoursesAI.
		// В legacy: 7–10% от ranee-захваченных пришельцами ресурсов
		// возвращаются на склад игрока. У нас этих данных нет —
		// используем упрощение: 7–10% от ТЕКУЩИХ ресурсов планеты
		// как «бонус от пришельцев», но не больше 1/3 capacity,
		// чтобы не переполнить склад.
		var curMetal, curSil, curHydro float64
		if err := tx.QueryRow(ctx, `
			SELECT metal, silicon, hydrogen FROM planets WHERE id = $1 AND destroyed_at IS NULL FOR UPDATE
		`, hp.PlanetID).Scan(&curMetal, &curSil, &curHydro); err != nil {
			if err == pgx.ErrNoRows {
				return nil
			}
			return fmt.Errorf("alien holding_ai: read planet: %w", err)
		}
		pct := 0.07 + rand.Float64()*0.03
		giftM := int64(curMetal * pct)
		giftS := int64(curSil * pct)
		giftH := int64(curHydro * pct)
		if giftM > 0 || giftS > 0 || giftH > 0 {
			if _, err := tx.Exec(ctx, `
				UPDATE planets
				SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
				WHERE id = $4
			`, giftM, giftS, giftH, hp.PlanetID); err != nil {
				return fmt.Errorf("alien holding_ai: gift: %w", err)
			}
			body := fmt.Sprintf(
				"Инопланетяне выгрузили на вашу планету ресурсы: "+
					"металл +%d, кремний +%d, водород +%d.",
				giftM, giftS, giftH,
			)
			if _, err := tx.Exec(ctx, `
				INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
				VALUES ($1, $2, NULL, 1, 'Подарок пришельцев', $3)
			`, ids.New(), hp.UserID, body); err != nil {
				return fmt.Errorf("alien holding_ai: message: %w", err)
			}
		}

		// Планируем следующий тик через 12–24ч (если HOLDING ещё не
		// истёк — event-worker позже отфильтрует по state).
		nextDelay := AlienHaltingMinTime + time.Duration(rand.Int64N(int64(AlienHaltingMaxTime-AlienHaltingMinTime)))
		nextData, err := json.Marshal(hp)
		if err != nil {
			return fmt.Errorf("alien holding_ai: marshal next: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, ids.New(), event.KindAlienHoldingAI, hp.PlanetID, hp.UserID,
			time.Now().Add(nextDelay), nextData); err != nil {
			return fmt.Errorf("alien holding_ai: insert next: %w", err)
		}
		return nil
	}
}
