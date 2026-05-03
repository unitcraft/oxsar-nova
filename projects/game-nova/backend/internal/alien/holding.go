package alien

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/pkg/ids"
)

// HoldingDefender — alien-флот, стоящий на планете в HOLDING, готовый
// участвовать в обороне против стороннего атакующего. Возвращается
// вызывающему слою (fleet/attack), который встраивает его в Defenders.
type HoldingDefender struct {
	EventID string      // id HOLDING-события (для закрытия если разбит)
	Side    battle.Side // готовая сторона с IsAliens=true и alien-юнитами
}

// LoadHoldingDefender читает активное HOLDING-событие для планеты (если
// есть) и собирает battle.Side с флотом пришельцев. Возвращает nil, если
// планета не в HOLDING.
//
// Используется из fleet/attack при атаке стороннего игрока — alien-флот
// подтягивается на сторону защитника (см. legacy Assault::loadDefenders,
// строки 207–219).
func LoadHoldingDefender(ctx context.Context, tx pgx.Tx, planetID string, cat *config.Catalog) (*HoldingDefender, error) {
	var eventID string
	var payload []byte
	err := tx.QueryRow(ctx, `
		SELECT id, payload FROM events
		WHERE planet_id = $1 AND kind = $2 AND state = 'wait'
		ORDER BY fire_at ASC LIMIT 1
	`, planetID, int(event.KindAlienHolding)).Scan(&eventID, &payload)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("alien holding defender: query: %w", err)
	}
	var hp holdingPayload
	if err := json.Unmarshal(payload, &hp); err != nil {
		return nil, fmt.Errorf("alien holding defender: parse payload: %w", err)
	}
	units := stacksToAlienUnits(hp.AlienFleet, cat)
	if len(units) == 0 {
		return nil, nil
	}
	return &HoldingDefender{
		EventID: eventID,
		Side: battle.Side{
			UserID:   "aliens",
			Username: "aliens:holding",
			IsAliens: true,
			Units:    units,
		},
	}, nil
}

// CloseHoldingIfWiped закрывает HOLDING-событие, если alien-флот
// уничтожен в бою. Принимает id HOLDING-события и результат его стороны
// из битвы. Вызывается из fleet/attack после применения потерь.
func CloseHoldingIfWiped(ctx context.Context, tx pgx.Tx, holdingEventID string, units []battle.UnitResult) error {
	var alive int64
	for _, u := range units {
		alive += u.QuantityEnd
	}
	if alive > 0 {
		// Обновим payload — оставшийся флот меньше, чтобы следующее
		// применение defender'а использовало актуальную численность.
		var payload []byte
		if err := tx.QueryRow(ctx,
			`SELECT payload FROM events WHERE id = $1`, holdingEventID).Scan(&payload); err != nil {
			return fmt.Errorf("alien holding update: read: %w", err)
		}
		var hp holdingPayload
		if err := json.Unmarshal(payload, &hp); err != nil {
			return fmt.Errorf("alien holding update: parse: %w", err)
		}
		hp.AlienFleet = survivorsToStacks(units)
		newPayload, err := json.Marshal(hp)
		if err != nil {
			return fmt.Errorf("alien holding update: marshal: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE events SET payload = $1 WHERE id = $2`,
			newPayload, holdingEventID); err != nil {
			return fmt.Errorf("alien holding update: save: %w", err)
		}
		return nil
	}
	// Флот полностью разбит — закрываем HOLDING. Дочерний HOLDING_AI
	// проверит родителя и завершится сам на следующем тике.
	if _, err := tx.Exec(ctx, `
		UPDATE events SET state = 'ok', processed_at = now() WHERE id = $1
	`, holdingEventID); err != nil {
		return fmt.Errorf("alien holding close: %w", err)
	}
	return nil
}

// stacksToAlienUnits разворачивает компактные fleetStack в battle.Unit
// с характеристиками из configs/ships.yml (поля 200–204). Имена берутся
// из alienShipOrder (helpers.go) — для сообщений и отладочных логов.
func stacksToAlienUnits(stacks []fleetStack, cat *config.Catalog) []battle.Unit {
	if len(stacks) == 0 {
		return nil
	}
	specByID := map[int]config.ShipSpec{}
	for _, spec := range cat.Ships.Ships {
		specByID[spec.ID] = spec
	}
	nameByID := map[int]string{}
	for _, entry := range alienShipOrder {
		nameByID[entry.unitID] = entry.name
	}
	out := make([]battle.Unit, 0, len(stacks))
	for _, s := range stacks {
		spec, ok := specByID[s.UnitID]
		if !ok {
			continue
		}
		out = append(out, battle.Unit{
			UnitID:   s.UnitID,
			Quantity: s.Quantity,
			Attack:   float64(spec.Attack),
			Shield:   float64(spec.Shield),
			Shell:    float64(spec.Shell),
			Name:     nameByID[s.UnitID],
		})
	}
	return out
}

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
//
// CapturedMetal/Silicon/Hydrogen — РАНЕЕ захваченные пришельцами
// ресурсы при первоначальной атаке (план 72.1.56 B8, legacy
// `AlienAI.class.php:1053-1061`). Используются в onUnloadAlienResources
// для возврата процента от _захваченного_, а не от текущих ресурсов
// планеты. Декрементируются по мере выгрузки.
type holdingPayload struct {
	PlanetID       string         `json:"planet_id"`
	UserID         string         `json:"user_id"`
	Tier           int            `json:"tier"`
	AlienFleet     []fleetStack   `json:"alien_fleet"`
	StartTime      time.Time      `json:"start_time"`
	HoldingEventID string         `json:"holding_event_id,omitempty"`
	PaidCredit     int64          `json:"paid_credit,omitempty"`
	PaidTimes      int            `json:"paid_times,omitempty"`
	CapturedMetal    int64 `json:"captured_metal,omitempty"`
	CapturedSilicon  int64 `json:"captured_silicon,omitempty"`
	CapturedHydrogen int64 `json:"captured_hydrogen,omitempty"`
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
//
// План 72.1.56 B8: lootM/lootS/lootH — ресурсы, захваченные при
// первоначальной атаке. Сохраняются в payload как Captured* для
// последующего unloadAlienResources (legacy
// `AlienAI.class.php:1053-1061` использует `parent_event["data"][$res]`).
func (s *Service) spawnHalt(ctx context.Context, tx pgx.Tx, pl alienPayload,
	survivors []fleetStack, lootM, lootS, lootH int64, r *rand.Rand) error {
	if len(survivors) == 0 {
		return nil
	}
	// Длительность HALT: [12ч, 24ч]. Используем rand для
	// детерминированности в тестах (caller передаёт свой seed).
	dur := AlienHaltingMinTime + time.Duration(r.Int64N(int64(AlienHaltingMaxTime-AlienHaltingMinTime)))
	hp := holdingPayload{
		PlanetID:         pl.PlanetID,
		UserID:           pl.UserID,
		Tier:             pl.Tier,
		AlienFleet:       survivors,
		StartTime:        time.Now().UTC(),
		CapturedMetal:    lootM,
		CapturedSilicon:  lootS,
		CapturedHydrogen: lootH,
	}
	if _, err := event.Insert(ctx, tx, event.InsertOpts{
		UserID:   &pl.UserID,
		PlanetID: &pl.PlanetID,
		Kind:     event.KindAlienHalt,
		FireAt:   time.Now().Add(dur),
		Payload:  hp,
	}); err != nil {
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
		holdingID, err := event.Insert(ctx, tx, event.InsertOpts{
			UserID:   &hp.UserID,
			PlanetID: &hp.PlanetID,
			Kind:     event.KindAlienHolding,
			FireAt:   holdingStart.Add(dur),
			Payload:  hp,
		})
		if err != nil {
			return fmt.Errorf("alien halt: insert holding: %w", err)
		}

		// Первый HOLDING_AI-тик через 5–10 сек.
		hp.HoldingEventID = holdingID
		aiDelay := 5*time.Second + time.Duration(rand.Int64N(int64(5*time.Second)))
		if _, err := event.Insert(ctx, tx, event.InsertOpts{
			UserID:   &hp.UserID,
			PlanetID: &hp.PlanetID,
			Kind:     event.KindAlienHoldingAI,
			FireAt:   time.Now().Add(aiDelay),
			Payload:  hp,
		}); err != nil {
			return fmt.Errorf("alien halt: insert holding_ai: %w", err)
		}

		// Сообщение: пришельцы блокировали планету.
		holdVars := map[string]string{
			"tier":  fmt.Sprintf("%d", hp.Tier),
			"until": holdingStart.Add(dur).Format("2006-01-02 15:04 MST"),
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 1, $3, $4)
		`, ids.New(), hp.UserID,
			s.tr("alien", "holding.startSubject", nil),
			s.tr("alien", "holding.startBody", holdVars)); err != nil {
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

		body := s.tr("alien", "holding.leftBody", nil)
		if hp.PaidTimes > 0 {
			body = s.tr("alien", "holding.leftPaidBody", map[string]string{
				"credits": fmt.Sprintf("%d", hp.PaidCredit),
				"times":   fmt.Sprintf("%d", hp.PaidTimes),
			})
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 1, $3, $4)
		`, ids.New(), hp.UserID, s.tr("alien", "holding.leftSubject", nil), body); err != nil {
			return fmt.Errorf("alien holding: message: %w", err)
		}
		return nil
	}
}

// HoldingAIHandler — event.Handler для KindAlienHoldingAI (80).
//
// Триггерится каждые 12–24ч внутри HOLDING. Выполняет одно случайное
// действие из двух реально реализованных в legacy (остальные 6 —
// пустые тела в AlienAI.class.php:1086–1126, см. docs/simplifications.md):
//
//  1. onUnloadAlienResoursesAI — 7–10% от ресурсов переходит на склад
//     игрока.
//  2. onExtractAlientShipsAI — часть alien-флота отделяется и улетает,
//     ослабляя HOLDING. Когда флот уходит полностью — HOLDING закрывается.
//
// После действия планирует следующий тик, если HOLDING ещё активен.
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

		// Выбор действия 50/50. В legacy — 2 из 8 вариантов реально
		// что-то делают, остальные 6 — заглушки, поэтому вероятность
		// 2×(1/8) ≈ 0.25 на каждое, итого ~50/50 между активными.
		if rand.IntN(2) == 0 {
			// План 72.1.56 B8: передаём указатель — функция декрементирует
			// CapturedX в hp, чтобы next-tick INSERT увидел обновление.
			if err := unloadAlienResources(ctx, tx, &hp, s.tr); err != nil {
				return fmt.Errorf("alien holding_ai: unload: %w", err)
			}
		} else {
			closed, err := extractAlienShips(ctx, tx, hp, s.tr)
			if err != nil {
				return fmt.Errorf("alien holding_ai: extract: %w", err)
			}
			if closed {
				return nil // HOLDING закрыт — не планируем следующий тик
			}
		}

		// Планируем следующий тик через 12–24ч (если HOLDING ещё не
		// истёк — event-worker позже отфильтрует по state).
		nextDelay := AlienHaltingMinTime + time.Duration(rand.Int64N(int64(AlienHaltingMaxTime-AlienHaltingMinTime)))
		if _, err := event.Insert(ctx, tx, event.InsertOpts{
			UserID:   &hp.UserID,
			PlanetID: &hp.PlanetID,
			Kind:     event.KindAlienHoldingAI,
			FireAt:   time.Now().Add(nextDelay),
			Payload:  hp,
		}); err != nil {
			return fmt.Errorf("alien holding_ai: insert next: %w", err)
		}
		return nil
	}
}

// unloadAlienResources — onUnloadAlienResoursesAI 1:1 с legacy
// `AlienAI.class.php:1053-1061` (план 72.1.56 B8):
//
//	$data[$res] = ceil(min($parent_event["data"][$res] * 0.7,
//	                       $parent_event["data"][$res] * 0.1 * $times));
//	$parent_event["data"][$res] = max(0, $parent_event["data"][$res] - $data[$res]);
//
// Возвращаем игроку процент от РАНЕЕ ЗАХВАЧЕННОГО (CapturedX в payload),
// не от текущих ресурсов планеты. После выгрузки декрементируем
// captured в hp **и** в payload родительского HOLDING-event'а
// (чтобы следующие тики видели обновлённое значение, как
// `parent_event["data"]` в legacy). HoldingAIHandler в caller-е
// перезапишет hp в next-event INSERT.
//
// times = max(1, hp.PaidTimes+1) — у legacy это control_times. В nova
// PaidTimes растёт от add-payment; используем как proxy для ramp-up.
func unloadAlienResources(ctx context.Context, tx pgx.Tx, hp *holdingPayload, tr func(string, string, map[string]string) string) error {
	if hp.CapturedMetal <= 0 && hp.CapturedSilicon <= 0 && hp.CapturedHydrogen <= 0 {
		return nil
	}
	times := int64(hp.PaidTimes) + 1
	giftM := unloadFraction(hp.CapturedMetal, times)
	giftS := unloadFraction(hp.CapturedSilicon, times)
	giftH := unloadFraction(hp.CapturedHydrogen, times)
	if giftM == 0 && giftS == 0 && giftH == 0 {
		return nil
	}
	hp.CapturedMetal -= giftM
	hp.CapturedSilicon -= giftS
	hp.CapturedHydrogen -= giftH
	if hp.CapturedMetal < 0 {
		hp.CapturedMetal = 0
	}
	if hp.CapturedSilicon < 0 {
		hp.CapturedSilicon = 0
	}
	if hp.CapturedHydrogen < 0 {
		hp.CapturedHydrogen = 0
	}
	if _, err := tx.Exec(ctx, `
		UPDATE planets
		SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
		WHERE id = $4 AND destroyed_at IS NULL
	`, giftM, giftS, giftH, hp.PlanetID); err != nil {
		return fmt.Errorf("gift: %w", err)
	}
	// Перезаписываем payload родительского HOLDING-event'а, чтобы
	// следующий тик HOLDING_AI видел уменьшенный captured (зеркалит
	// legacy `$parent_event["data"]`).
	if hp.HoldingEventID != "" {
		newHold, err := json.Marshal(hp)
		if err != nil {
			return fmt.Errorf("marshal holding: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE events SET payload = $1 WHERE id = $2`,
			newHold, hp.HoldingEventID); err != nil {
			return fmt.Errorf("update holding payload: %w", err)
		}
	}
	giftVars := map[string]string{
		"metal":    fmt.Sprintf("%d", giftM),
		"silicon":  fmt.Sprintf("%d", giftS),
		"hydrogen": fmt.Sprintf("%d", giftH),
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, NULL, 1, $3, $4)
	`, ids.New(), hp.UserID,
		tr("alien", "holding.giftSubject", nil),
		tr("alien", "holding.giftBody", giftVars)); err != nil {
		return fmt.Errorf("message: %w", err)
	}
	return nil
}

// unloadFraction — pure helper для legacy `AlienAI.class.php:1056`:
// `ceil(min(captured*0.7, captured*0.1*times))`. Округление вверх
// чтобы любой ненулевой captured хотя бы в 1 единицу выгрузить.
func unloadFraction(captured, times int64) int64 {
	if captured <= 0 || times <= 0 {
		return 0
	}
	cap70 := captured * 7 / 10
	if captured*7%10 != 0 {
		cap70++ // ceil
	}
	capTimes := captured * times / 10
	if captured*times%10 != 0 {
		capTimes++
	}
	if capTimes < cap70 {
		return capTimes
	}
	return cap70
}

// extractAlienShips — onExtractAlientShipsAI (legacy строки 1025–1079):
// от случайного типа alien-корабля «отпочковывается» часть флота и улетает.
// Убывание: quantity = min(q-1, ceil(q × 0.01 × times²)) — где times = 1
// (у нас control_times не продвигается, упрощение). Если после
// уменьшения флот пуст — HOLDING закрывается, возвращаем closed=true.
func extractAlienShips(ctx context.Context, tx pgx.Tx, hp holdingPayload, tr func(string, string, map[string]string) string) (bool, error) {
	// Загружаем актуальный payload HOLDING (может быть обновлён после
	// боя со сторонним атакующим через CloseHoldingIfWiped).
	var holdingPayloadBytes []byte
	if err := tx.QueryRow(ctx,
		`SELECT payload FROM events WHERE id = $1 FOR UPDATE`,
		hp.HoldingEventID).Scan(&holdingPayloadBytes); err != nil {
		if err == pgx.ErrNoRows {
			return true, nil
		}
		return false, fmt.Errorf("load holding: %w", err)
	}
	var hold holdingPayload
	if err := json.Unmarshal(holdingPayloadBytes, &hold); err != nil {
		return false, fmt.Errorf("parse holding: %w", err)
	}
	if len(hold.AlienFleet) == 0 {
		return true, nil
	}

	// Выбираем случайный стек с quantity >= 2 (один всегда остаётся).
	extractable := make([]int, 0, len(hold.AlienFleet))
	for i, s := range hold.AlienFleet {
		if s.Quantity >= 2 {
			extractable = append(extractable, i)
		}
	}
	if len(extractable) == 0 {
		// Все стеки по 1 — уходят целиком.
		for i := range hold.AlienFleet {
			hold.AlienFleet[i].Quantity = 0
		}
	} else {
		idx := extractable[rand.IntN(len(extractable))]
		stack := &hold.AlienFleet[idx]
		// quantity = ceil(q × 0.01) как минимум 1, максимум q-1.
		extract := int64(float64(stack.Quantity)*0.01 + 0.999)
		if extract < 1 {
			extract = 1
		}
		if extract > stack.Quantity-1 {
			extract = stack.Quantity - 1
		}
		stack.Quantity -= extract
	}

	// Отфильтровываем нулевые стеки.
	newFleet := make([]fleetStack, 0, len(hold.AlienFleet))
	for _, s := range hold.AlienFleet {
		if s.Quantity > 0 {
			newFleet = append(newFleet, s)
		}
	}
	hold.AlienFleet = newFleet

	if len(hold.AlienFleet) == 0 {
		// Флот разошёлся целиком — закрываем HOLDING.
		if _, err := tx.Exec(ctx, `
			UPDATE events SET state = 'ok', processed_at = now() WHERE id = $1
		`, hp.HoldingEventID); err != nil {
			return false, fmt.Errorf("close holding: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 1, $3, $4)
		`, ids.New(), hp.UserID,
			tr("alien", "holding.scatteredSubject", nil),
			tr("alien", "holding.scatteredBody", nil)); err != nil {
			return false, fmt.Errorf("message: %w", err)
		}
		return true, nil
	}

	// Сохраняем уменьшенный флот в HOLDING-payload.
	newBytes, err := json.Marshal(hold)
	if err != nil {
		return false, fmt.Errorf("marshal holding: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE events SET payload = $1 WHERE id = $2`,
		newBytes, hp.HoldingEventID); err != nil {
		return false, fmt.Errorf("save holding: %w", err)
	}
	return false, nil
}
