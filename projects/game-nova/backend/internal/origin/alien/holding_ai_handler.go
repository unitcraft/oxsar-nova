package alien

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/pkg/rng"
)

// HoldingSubphase — типизированное имя одной из 8 подфаз HOLDING_AI
// (R13). 2 активные + 6 заглушек, как в origin AlienAI.class.php:940-947.
type HoldingSubphase string

const (
	SubphaseExtractAlienShips     HoldingSubphase = "extract_alien_ships"
	SubphaseUnloadAlienResources  HoldingSubphase = "unload_alien_resources"
	SubphaseRepairUserUnits       HoldingSubphase = "repair_user_units"
	SubphaseAddUserUnits          HoldingSubphase = "add_user_units"
	SubphaseAddCredits            HoldingSubphase = "add_credits"
	SubphaseAddArtefact           HoldingSubphase = "add_artefact"
	SubphaseGenerateAsteroid      HoldingSubphase = "generate_asteroid"
	SubphaseFindPlanetAfterBattle HoldingSubphase = "find_planet_after_battle"
)

// holdingSubphasesOrder — порядок выбора, как в PHP-варианте
// (определяет детерминированный rand-обход). Веса все по 10 = равные
// 1/8 (origin AlienAI:940-947).
var holdingSubphasesOrder = []HoldingSubphase{
	SubphaseExtractAlienShips,
	SubphaseUnloadAlienResources,
	SubphaseRepairUserUnits,
	SubphaseAddUserUnits,
	SubphaseAddCredits,
	SubphaseAddArtefact,
	SubphaseGenerateAsteroid,
	SubphaseFindPlanetAfterBattle,
}

// pickHoldingSubphase — равновесный выбор 1 из 8 (origin
// AlienAI.class.php:949-966). Pure-функция для property-тестов.
func pickHoldingSubphase(r *rng.R) HoldingSubphase {
	idx := r.IntN(len(holdingSubphasesOrder))
	return holdingSubphasesOrder[idx]
}

// HoldingAIHandler возвращает event.Handler для KindAlienHoldingAI
// (Kind=80).
//
// Семантика origin (AlienAI.class.php:924-1014):
//
//  1. Загрузить parent KindAlienHolding (если ушёл/done — silent skip).
//  2. Прочитать paid_credit из payload, обнулить его в payload (это
//     consume-once значение; платёж агрегируется в paid_sum_credit /
//     paid_times). Вызвать sub-phase: 1 из 8 равновесно.
//  3. control_times++.
//  4. Спланировать следующий тик: fire_at = now + clamp(min(12h,
//     30s*times) ... max(24h, 60s*times)), ограниченный parent.fire_at.
//     Если parent почти закончится — этот тик финальный, не планируем.
//  5. Если sub-phase или paid_credit > 0 → продлеваем parent fire_at:
//     parent.fire_at += 2h * paid_credit / 50, capped на
//     start_time + HaltingMaxRealTime (15 дней).
//  6. С вероятностью 1% — checkAlientNeeds (в nova спавнится через
//     scheduler alien_spawn — этот тик игнорируем эту ветку: 1%
//     случайный спавн миссии излишен при наличии глобального
//     scheduler'а; задокументировано как сознательное расхождение).
//
// Активные sub-phases: ExtractAlienShips (убывание alien-флота),
// UnloadAlienResources (часть alien-флота уходит + дарит игроку
// ресурсы из parent-snapshot 0..10%×times², capped 70%).
//
// Заглушки (как в origin PHP:1086-1124 — пустые тела):
// RepairUserUnits, AddUserUnits, AddCredits, AddArtefact,
// GenerateAsteroid, FindPlanetAfterBattle.
//
// Идемпотентность: один тик HOLDING_AI = одно мутирующее действие в
// одной БД-транзакции; worker гарантирует exactly-once через FOR
// UPDATE SKIP LOCKED. control_times++ только при успехе всей tx.
//
// R3 audit: structured slog с event_id / planet_id / subphase /
// control_times. R8 metrics — автоматически worker'ом.
// R10 — universe изоляция уже на уровне events.user_id FK.
// R12 — i18n сообщений через s.tr() (переиспользуем
// `holding.giftSubject/Body` и `holding.scatteredSubject/Body`,
// созданные в плане 15).
func (s *Service) HoldingAIHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl HoldingAIPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("alien holding_ai: parse payload: %w", err)
		}
		if pl.PlanetID == "" || pl.UserID == "" || pl.HoldingEventID == "" {
			return fmt.Errorf("alien holding_ai: empty planet_id/user_id/holding_event_id")
		}

		// 1) Проверяем parent — если HOLDING закрыт/удалён, тихо завершаемся.
		var parentState string
		var parentFireAt time.Time
		var parentRaw []byte
		err := tx.QueryRow(ctx, `
			SELECT state, fire_at, payload FROM events WHERE id = $1::uuid
		`, pl.HoldingEventID).Scan(&parentState, &parentFireAt, &parentRaw)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.InfoContext(ctx, "event_alien_holding_ai_skip_parent_gone",
					slog.String("event_id", e.ID),
					slog.String("holding_event_id", pl.HoldingEventID))
				return nil
			}
			return fmt.Errorf("alien holding_ai: load parent: %w", err)
		}
		if parentState != "wait" {
			slog.InfoContext(ctx, "event_alien_holding_ai_skip_parent_state",
				slog.String("event_id", e.ID),
				slog.String("parent_state", parentState))
			return nil
		}

		// 2) Consume paid_credit (обнуляем в payload — но используем
		// это значение для возможного продления parent ниже).
		paidThisTick := pl.PaidCredit
		pl.PaidCredit = 0
		if paidThisTick > 0 {
			pl.PaidSumCredit += paidThisTick
			pl.PaidTimes++
		}

		// Origin: $times = max(1, control_times) — на первом тике
		// times=1 (control_times=0).
		times := pl.ControlTimes
		if times < 1 {
			times = 1
		}

		// 3) Выбираем подфазу. Seed детерминированный по event.ID, чтобы
		// retry того же тика выбирал ту же ветку (сама ветка идемпотентна
		// внутри tx).
		r := rng.New(fnvHashString(e.ID))
		sub := pickHoldingSubphase(r)

		// Загружаем parent-snapshot (для UnloadAlienResources). В origin
		// это `parent_event["data"]["metal"]/...` — захваченные ресурсы.
		// В nova HALT/HOLDING может не сохранять snapshot (план 15
		// упрощение — поля могут быть 0); тогда unload даёт 0 ресурсов,
		// как и в origin при пустом snapshot.
		var parentSnap HoldingParentSnapshot
		_ = json.Unmarshal(parentRaw, &parentSnap)

		// 4) Выполняем подфазу. parentChanged = true ⇒ ниже надо обновить
		// parent.payload (и применить продление, как в origin:991-1003).
		var parentChanged bool
		switch sub {
		case SubphaseExtractAlienShips:
			closed, changed, err := s.subphaseExtractAlienShips(ctx, tx, e, &pl, times, r, false)
			if err != nil {
				return fmt.Errorf("alien holding_ai: extract: %w", err)
			}
			if closed {
				return nil
			}
			parentChanged = changed
		case SubphaseUnloadAlienResources:
			closed, changed, err := s.subphaseUnloadAlienResources(ctx, tx, e, &pl, &parentSnap, times, r)
			if err != nil {
				return fmt.Errorf("alien holding_ai: unload: %w", err)
			}
			if closed {
				return nil
			}
			parentChanged = changed
		case SubphaseRepairUserUnits:
			s.subphaseStub(ctx, e, sub)
		case SubphaseAddUserUnits:
			s.subphaseStub(ctx, e, sub)
		case SubphaseAddCredits:
			s.subphaseStub(ctx, e, sub)
		case SubphaseAddArtefact:
			s.subphaseStub(ctx, e, sub)
		case SubphaseGenerateAsteroid:
			s.subphaseStub(ctx, e, sub)
		case SubphaseFindPlanetAfterBattle:
			s.subphaseStub(ctx, e, sub)
		}

		// 5) control_times++ (origin AlienAI:978).
		pl.ControlTimes++

		// 6) Если parent_changed (alien-флот изменился) или paid_credit > 0,
		// продлеваем parent.fire_at (origin:991-1003):
		//   end_time = parent.fire_at + 2h * paid_credit / 50
		//   capped at start + HaltingMaxRealTime
		//   apply only if parent_changed || paid > 0.
		// Также сохраняем обновлённый parent.payload (alien_fleet,
		// snapshot ресурсов, paid_*).
		if parentChanged || paidThisTick > 0 {
			newParentFireAt := parentFireAt
			if paidThisTick > 0 {
				addSec := s.cfg.HoldingPaySecondsPerCredit * float64(paidThisTick)
				newParentFireAt = parentFireAt.Add(time.Duration(addSec) * time.Second)
				cap := pl.StartTime.Add(s.cfg.HaltingMaxRealTime)
				if newParentFireAt.After(cap) {
					newParentFireAt = cap
				}
			}
			// Парсим parent payload в HoldingAIPayload-shape (совместимо).
			var parentPL HoldingAIPayload
			if err := json.Unmarshal(parentRaw, &parentPL); err != nil {
				return fmt.Errorf("alien holding_ai: parse parent payload: %w", err)
			}
			// Проводим обновлённые поля флота / snapshot / paid_* из
			// локального pl в parent (single source of truth =
			// parent KindAlienHolding event).
			parentPL.AlienFleet = pl.AlienFleet
			parentPL.PaidCredit = 0
			parentPL.PaidSumCredit = pl.PaidSumCredit
			parentPL.PaidTimes = pl.PaidTimes
			parentPL.Metal = parentSnap.Metal
			parentPL.Silicon = parentSnap.Silicon
			parentPL.Hydrogen = parentSnap.Hydrogen
			newParentRaw, err := json.Marshal(parentPL)
			if err != nil {
				return fmt.Errorf("alien holding_ai: marshal parent: %w", err)
			}
			if _, err := tx.Exec(ctx, `
				UPDATE events SET payload = $1, fire_at = $2
				WHERE id = $3::uuid AND state = 'wait'
			`, newParentRaw, newParentFireAt, pl.HoldingEventID); err != nil {
				return fmt.Errorf("alien holding_ai: update parent: %w", err)
			}
			parentFireAt = newParentFireAt
		}

		// 7) Планируем следующий тик HOLDING_AI (origin:973-988).
		nextDur := HoldingAISubphaseDuration(s.cfg, pl.ControlTimes, r)
		now := nowFn()
		nextAt := now.Add(nextDur)
		// Cap на parent.fire_at-2s — следующий тик не должен пересечь
		// окончание HOLDING (origin:975 "$parent_time = $parent_event[time]-2").
		parentDeadline := parentFireAt.Add(-2 * time.Second)
		if nextAt.After(parentDeadline) {
			// Если parent почти закончится (parent - now < 30 минут),
			// origin вообще не планирует следующий тик (PHP:976
			// `$parent_time - $start_time > 60*30`).
			if parentDeadline.Sub(now) <= 30*time.Minute {
				slog.InfoContext(ctx, "event_alien_holding_ai_final_tick",
					slog.String("event_id", e.ID),
					slog.String("subphase", string(sub)),
					slog.Int("control_times", pl.ControlTimes))
				return nil
			}
			nextAt = parentDeadline
		}

		if _, err := event.Insert(ctx, tx, event.InsertOpts{
			UserID:   &pl.UserID,
			PlanetID: &pl.PlanetID,
			Kind:     event.KindAlienHoldingAI,
			FireAt:   nextAt,
			Payload:  pl,
		}); err != nil {
			return fmt.Errorf("alien holding_ai: insert next: %w", err)
		}
		slog.InfoContext(ctx, "event_alien_holding_ai_tick",
			slog.String("event_id", e.ID),
			slog.String("subphase", string(sub)),
			slog.Int("control_times", pl.ControlTimes),
			slog.Int64("paid_this_tick", paidThisTick),
			slog.Int64("paid_sum", pl.PaidSumCredit),
			slog.Bool("parent_changed", parentChanged),
			slog.Time("next_at", nextAt))
		return nil
	}
}

// subphaseExtractAlienShips — порт PHP:1025-1079 (без unload).
//
// Выбирает случайный alien-стек с quantity >= 2, вычисляет
//
//	extract = ceil(q × 0.01 × times²)
//
// (capped на q-1 чтобы стек не уходил полностью), и обнуляет
// эту часть. Если все стеки == 1 — стеки уходят целиком и HOLDING
// закрывается (closed=true).
//
// Возвращает (closed, changed, err): closed=true ⇒ HOLDING закрыт
// и handler НЕ должен планировать next_tick. changed=true ⇒
// alien_fleet в payload изменился, parent.payload обновить.
func (s *Service) subphaseExtractAlienShips(
	ctx context.Context, tx pgx.Tx, e event.Event,
	pl *HoldingAIPayload, times int, r *rng.R, _ bool,
) (closed, changed bool, err error) {

	if len(pl.AlienFleet) == 0 {
		// Защита: если флот уже пуст (HOLDING разбит сторонним атакующим
		// через CloseHoldingIfWiped), то закрываем HOLDING.
		if err := s.closeHoldingScattered(ctx, tx, pl); err != nil {
			return false, false, err
		}
		return true, false, nil
	}

	// Выбираем random стек с quantity >= 2 (один всегда остаётся в
	// origin — array_rand + проверка quantity).
	extractable := make([]int, 0, len(pl.AlienFleet))
	for i, st := range pl.AlienFleet {
		if st.Quantity >= 2 {
			extractable = append(extractable, i)
		}
	}
	if len(extractable) == 0 {
		// Все стеки == 1 → весь флот уходит, HOLDING закрывается.
		if err := s.closeHoldingScattered(ctx, tx, pl); err != nil {
			return false, false, err
		}
		return true, true, nil
	}
	idx := extractable[r.IntN(len(extractable))]
	stack := &pl.AlienFleet[idx]
	// extract = ceil(q × 0.01 × times²), capped q-1.
	frac := float64(stack.Quantity) * 0.01 * float64(times) * float64(times)
	extract := int64(math.Ceil(frac))
	if extract < 1 {
		extract = 1
	}
	if extract > stack.Quantity-1 {
		extract = stack.Quantity - 1
	}
	stack.Quantity -= extract

	slog.InfoContext(ctx, "event_alien_holding_ai_extract",
		slog.String("event_id", e.ID),
		slog.Int("unit_id", stack.UnitID),
		slog.Int64("extracted", extract),
		slog.Int64("remaining", stack.Quantity),
		slog.Int("times", times))
	return false, true, nil
}

// subphaseUnloadAlienResources — порт PHP:1081-1084 (= ExtractAlienShips
// с unload=true PHP:1053-1061).
//
// Делает то же убывание alien-флота + дополнительно дарит игроку
// часть ресурсов из parent-snapshot:
//
//	gift = ceil(min(snapshot[res] × 0.7, snapshot[res] × 0.1 × times))
//
// Если snapshot пустой (0/0/0) — действие сводится к чистому
// ExtractAlienShips (но с переменным diminish — origin: при
// good_bonus=true extract уменьшается ×0.3).
func (s *Service) subphaseUnloadAlienResources(
	ctx context.Context, tx pgx.Tx, e event.Event,
	pl *HoldingAIPayload, snap *HoldingParentSnapshot, times int, r *rng.R,
) (closed, changed bool, err error) {

	// Считаем подарок ДО ExtractAlienShips, чтобы good_bonus был известен.
	timesF := float64(times)
	giftM := unloadGift(snap.Metal, timesF)
	giftS := unloadGift(snap.Silicon, timesF)
	giftH := unloadGift(snap.Hydrogen, timesF)
	goodBonus := giftM > 1_000_000 || giftS > 1_000_000 || giftH > 1_000_000

	// Применяем подарок.
	if giftM > 0 || giftS > 0 || giftH > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal = metal + $1,
			                   silicon = silicon + $2,
			                   hydrogen = hydrogen + $3
			WHERE id = $4::uuid AND destroyed_at IS NULL
		`, giftM, giftS, giftH, pl.PlanetID); err != nil {
			return false, false, fmt.Errorf("unload gift resources: %w", err)
		}
		// Уменьшаем snapshot — origin: $parent_event["data"][$res] -= $data[$res].
		snap.Metal = maxInt64(0, snap.Metal-giftM)
		snap.Silicon = maxInt64(0, snap.Silicon-giftS)
		snap.Hydrogen = maxInt64(0, snap.Hydrogen-giftH)

		if err := s.sendMessage(ctx, tx, pl.UserID,
			"alien", "holding.giftSubject", "holding.giftBody",
			map[string]string{
				"metal":    fmt.Sprintf("%d", giftM),
				"silicon":  fmt.Sprintf("%d", giftS),
				"hydrogen": fmt.Sprintf("%d", giftH),
			}); err != nil {
			return false, false, err
		}
		slog.InfoContext(ctx, "event_alien_holding_ai_gift",
			slog.String("event_id", e.ID),
			slog.String("user_id", pl.UserID),
			slog.Int64("metal", giftM),
			slog.Int64("silicon", giftS),
			slog.Int64("hydrogen", giftH),
			slog.Bool("good_bonus", goodBonus))
	}

	// Теперь стандартный ExtractAlienShips (с поправкой good_bonus —
	// origin:1063 умножает extract на 0.3 при good_bonus).
	if len(pl.AlienFleet) == 0 {
		if err := s.closeHoldingScattered(ctx, tx, pl); err != nil {
			return false, false, err
		}
		return true, true, nil
	}
	extractable := make([]int, 0, len(pl.AlienFleet))
	for i, st := range pl.AlienFleet {
		if st.Quantity >= 2 {
			extractable = append(extractable, i)
		}
	}
	if len(extractable) == 0 {
		if err := s.closeHoldingScattered(ctx, tx, pl); err != nil {
			return false, false, err
		}
		return true, true, nil
	}
	idx := extractable[r.IntN(len(extractable))]
	stack := &pl.AlienFleet[idx]
	frac := float64(stack.Quantity) * 0.01 * float64(times) * float64(times)
	if goodBonus {
		frac *= 0.3
	}
	extract := int64(math.Ceil(frac))
	if extract < 1 {
		extract = 1
	}
	if extract > stack.Quantity-1 {
		extract = stack.Quantity - 1
	}
	stack.Quantity -= extract

	slog.InfoContext(ctx, "event_alien_holding_ai_unload_extract",
		slog.String("event_id", e.ID),
		slog.Int("unit_id", stack.UnitID),
		slog.Int64("extracted", extract),
		slog.Int64("remaining", stack.Quantity),
		slog.Int("times", times),
		slog.Bool("good_bonus", goodBonus))
	return false, true, nil
}

// unloadGift — формула подарка из снимка ресурсов
// (origin AlienAI.class.php:1058):
//
//	gift = ceil(min(snapshot * 0.7, snapshot * 0.1 * times))
//
// Pure-функция для тестов.
func unloadGift(snapshot int64, times float64) int64 {
	if snapshot <= 0 {
		return 0
	}
	cap := math.Floor(float64(snapshot) * 0.7)
	val := math.Ceil(float64(snapshot) * 0.1 * times)
	if val > cap {
		val = cap
	}
	if val <= 0 {
		return 0
	}
	return int64(val)
}

// subphaseStub — заглушка для 6 неактивных подфаз. В origin
// AlienAI.class.php:1086-1124 эти 6 — пустые тела (`function() {}`).
// Семантика «делает ничего, но засчитывается тиком (control_times++)»
// сохранена в HoldingAIHandler выше — он инкрементит control_times
// независимо от подфазы.
//
// Логируем как audit-event, чтобы на проде было видно распределение
// ~1/8 на каждую заглушку (R3). Метрика автоматически на уровне
// worker (R8).
func (s *Service) subphaseStub(ctx context.Context, e event.Event, sub HoldingSubphase) {
	slog.InfoContext(ctx, "event_alien_holding_ai_stub_subphase",
		slog.String("event_id", e.ID),
		slog.String("subphase", string(sub)))
	// TODO(plan-66 / A14): origin AlienAI:1086-1124 эти 6 методов имеют
	// пустые тела. Если когда-нибудь решим расширять (artefact / asteroid /
	// repair / units / credits) — это новая фича, не часть paritет-плана.
}

// closeHoldingScattered — alien-флот ушёл целиком, HOLDING закрывается.
// Помечает parent KindAlienHolding как 'ok', шлёт сообщение
// "пришельцы рассеялись".
func (s *Service) closeHoldingScattered(ctx context.Context, tx pgx.Tx, pl *HoldingAIPayload) error {
	if _, err := tx.Exec(ctx, `
		UPDATE events SET state = 'ok', processed_at = now()
		WHERE id = $1::uuid AND state = 'wait'
	`, pl.HoldingEventID); err != nil {
		return fmt.Errorf("close holding: %w", err)
	}
	if err := s.sendMessage(ctx, tx, pl.UserID,
		"alien", "holding.scatteredSubject", "holding.scatteredBody", nil); err != nil {
		return err
	}
	return nil
}
