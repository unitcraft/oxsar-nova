package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// BuildingPayload — payload события завершения стройки здания или
// исследования. Структура одинаковая, различается только Kind события
// и таблица, куда применяется уровень.
//
// Для KindDemolishConstruction TargetLevel — желаемый уровень ПОСЛЕ
// сноса (обычно curLevel-1, может быть 0). См. HandleDemolishConstruction.
type BuildingPayload struct {
	QueueID     string `json:"queue_id"`
	UnitID      int    `json:"unit_id"`
	TargetLevel int    `json:"target_level"`
}

// ShipyardPayload — payload события окончания постройки кораблей/обороны.
// Здесь важнее count, а не target_level.
type ShipyardPayload struct {
	QueueID string `json:"queue_id"`
	UnitID  int    `json:"unit_id"`
	Count   int64  `json:"count"`
	IsDefense bool `json:"is_defense"`
}

// HandleBuildConstruction повышает уровень здания на планете,
// закрывает запись в construction_queue. Идемпотентен: если уровень уже
// >= target, ничего не делает.
func HandleBuildConstruction(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl BuildingPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.PlanetID == nil {
		return fmt.Errorf("building event without planet_id")
	}

	var cur int
	err := tx.QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		*e.PlanetID, pl.UnitID,
	).Scan(&cur)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("select level: %w", err)
	}
	if cur >= pl.TargetLevel {
		// уже применено ранее — идемпотентность
		_, _ = tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID)
		return nil
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO buildings (planet_id, unit_id, level)
		VALUES ($1, $2, $3)
		ON CONFLICT (planet_id, unit_id) DO UPDATE SET level = EXCLUDED.level
	`, *e.PlanetID, pl.UnitID, pl.TargetLevel); err != nil {
		return fmt.Errorf("upsert building: %w", err)
	}

	// План 23: инкрементируем used_fields только при первой постройке
	// здания (cur==0, target==1). Апгрейд того же здания поля не занимает.
	// Solar satellite и ракеты не считаются «зданиями» на полях в legacy
	// (см. Planet.class.php:717 — getFields), но поскольку их нельзя
	// построить через construction_queue как buildings (они через
	// shipyard), здесь не фильтруем.
	if cur == 0 && pl.TargetLevel == 1 {
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET used_fields = used_fields + 1 WHERE id = $1`,
			*e.PlanetID); err != nil {
			return fmt.Errorf("inc used_fields: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID); err != nil {
		return fmt.Errorf("close queue: %w", err)
	}
	return nil
}

// HandleResearch повышает уровень research у игрока. Идемпотентен.
func HandleResearch(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl BuildingPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.UserID == nil {
		return fmt.Errorf("research event without user_id")
	}
	var cur int
	err := tx.QueryRow(ctx,
		`SELECT level FROM research WHERE user_id=$1 AND unit_id=$2`,
		*e.UserID, pl.UnitID,
	).Scan(&cur)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("select research: %w", err)
	}
	if cur >= pl.TargetLevel {
		_, _ = tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID)
		return nil
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO research (user_id, unit_id, level)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, unit_id) DO UPDATE SET level = EXCLUDED.level
	`, *e.UserID, pl.UnitID, pl.TargetLevel); err != nil {
		return fmt.Errorf("upsert research: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID); err != nil {
		return fmt.Errorf("close queue: %w", err)
	}
	return nil
}

// HandleBuildFleet применяет постройку корабля.
//
// В отличие от постройки здания/исследования, тут порция кораблей
// (Count) добавляется к существующему запасу. Идемпотентность
// обеспечивается через проверку статуса очереди: если status=done,
// событие уже было обработано ранее, ничего не делаем.
func HandleBuildFleet(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl ShipyardPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.PlanetID == nil {
		return fmt.Errorf("fleet event without planet_id")
	}

	var status string
	err := tx.QueryRow(ctx,
		`SELECT status FROM shipyard_queue WHERE id=$1`, pl.QueueID,
	).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil // уже удалено — идемпотентность
		}
		return fmt.Errorf("select queue: %w", err)
	}
	if status == "done" {
		return nil
	}

	targetTable := "ships"
	if pl.IsDefense {
		targetTable = "defense"
	}

	if _, err := tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (planet_id, unit_id, count)
		VALUES ($1, $2, $3)
		ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = %s.count + EXCLUDED.count
	`, targetTable, targetTable),
		*e.PlanetID, pl.UnitID, pl.Count,
	); err != nil {
		return fmt.Errorf("upsert %s: %w", targetTable, err)
	}
	if _, err := tx.Exec(ctx, `UPDATE shipyard_queue SET status='done' WHERE id=$1`, pl.QueueID); err != nil {
		return fmt.Errorf("close shipyard queue: %w", err)
	}
	return nil
}

// DeliveryArtefactsPayload — payload события KindDeliveryArtefacts (план 65 Ф.2,
// D-035). Доставка артефактов флотом-курьером (источник — биржа артефактов
// плана 68 либо premium-механика подарков).
//
// Поля:
//   - FleetID — UUID записи fleets (флот, везущий груз; по прибытии переходит
//     в state='returning' для возврата домой).
//   - ArtefactIDs — список UUID artefacts_user.id, чьи владельцы должны быть
//     переписаны на e.UserID/e.PlanetID (получатель + планета назначения
//     достаются из самого Event'а, как у demolish — см. R10).
type DeliveryArtefactsPayload struct {
	FleetID     string   `json:"fleet_id"`
	ArtefactIDs []string `json:"artefact_ids"`
}

// HandleDeliveryArtefacts применяет доставку артефактов адресату.
//
// Семантика origin (EventHandler::transport ветка EVENT_DELIVERY_ARTEFACTS,
// EventHandler.class.php:2718-2754 + Artefact::onOwnerChange,
// Artefact.class.php:379):
//   - для каждого артефакта в payload: UPDATE user_id, planet_id на
//     получателя (destuser/destination в legacy → e.UserID/e.PlanetID
//     в nova);
//   - активный артефакт деактивируется (active=0 + revert эффекта),
//     запланированные delay/expire события снимаются;
//   - флот переходит в state='returning' и улетает обратно (sendBack);
//   - ресурсы НЕ передаются (отличие от EVENT_TRANSPORT/DELIVERY_RESOURSES,
//     см. EventHandler.class.php:2688 — ветка `if mode != DELIVERY_ARTEFACTS`
//     пропускает updateUserRes).
//
// Решения для nova:
//
//   - **Состояние артефакта после доставки** — переводим в `held`. Это
//     отличие от origin, где `active=0` без явного состояния «в инвентаре»
//     (в legacy active=0 ↔ держится в инвентаре). В nova у нас artefact_state
//     enum (held/delayed/active/expired/consumed); `held` — точный аналог
//     «в инвентаре, не активирован». Если артефакт прилетел `active` (что
//     возможно, если биржевой код не сбросил состояние перед выставлением
//     лота), просто откидываем active → held: revert-эффекта пройдёт лениво
//     при следующей пересборке через ScoreRecalc/ActiveBattleModifiers
//     (effect-стек в nova вычисляется по списку активных артефактов
//     каждый раз, см. service.go:349).
//
//     **Сознательное упрощение** (фиксируем в simplifications.md): не
//     зовём `applyChange(revert)` синхронно для `active → held` — в origin
//     это нужно, потому что эффекты артефактов зашиты как инкременты
//     полей `users.*`/`planets.*`. В nova зашиты так же (см. effects.go),
//     но дельта суммируется при чтении (`ActiveBattleModifiers`,
//     `applyChange` зовётся только при Activate/Deactivate), поэтому
//     откат значений колонок не требуется. Для `factor_user` /
//     `factor_planet` (которые применяются через `applyChange` и
//     остаются в колонках до явного Deactivate) handler полагается на
//     то, что биржевая операция ставит артефакт в `held` ДО полёта.
//     Если в проде поймаем `active`-артефакт в delivery — добавим
//     явный revert-вызов (отдельный план).
//
//   - **Идемпотентность** — артефакт уже принадлежит e.UserID и e.PlanetID:
//     skip (no-op); state уже не active: skip revert. Флот в state ≠
//     'outbound' (returning/done): skip всё (как и ArriveHandler в
//     fleet/events.go:52).
//
//   - **Per-universe (R10)** — соблюдается через FK artefacts_user.user_id
//     → users.id (universe-bound) и e.UserID/e.PlanetID, которые
//     гарантированно из той же вселенной (event создан в её контексте).
//     Дополнительно проверяем, что user_id артефакта-источника из той же
//     вселенной что и e.UserID — защита от рассинхрона на стыке с биржей
//     (план 68): если поломается — лучше падать с ошибкой, чем перенести
//     груз через границу вселенной.
//
//   - **Audit** — структурированный slog (R3) с полями event_id, fleet_id,
//     planet_id_to, user_id_to, artefact_count, transferred (после
//     handler'а — фактически переписанные ID).
//
//   - **Метрики (R8)** — автоматически на уровне worker'а
//     (`oxsar_events_processed{kind="23"}`).
//
// Граничные случаи:
//
//   - payload.ArtefactIDs пустой → ошибка: пустая доставка не имеет смысла.
//   - artefact не найден в БД (удалён до прибытия): пропускаем с warning,
//     не падаем — биржевой код мог уже расформировать лот.
//   - artefact принадлежит уже e.UserID + e.PlanetID: идемпотентный skip.
//   - флот не найден: возвращаем nil (как ArriveHandler).
func HandleDeliveryArtefacts(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl DeliveryArtefactsPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.UserID == nil {
		return fmt.Errorf("delivery_artefacts event without user_id (recipient)")
	}
	if e.PlanetID == nil {
		return fmt.Errorf("delivery_artefacts event without planet_id (destination)")
	}
	if pl.FleetID == "" {
		return fmt.Errorf("delivery_artefacts payload missing fleet_id")
	}
	if len(pl.ArtefactIDs) == 0 {
		return fmt.Errorf("delivery_artefacts payload has empty artefact_ids")
	}

	// Шаг 1: проверка состояния флота. Зеркалит ArriveHandler
	// (fleet/events.go:41-54): только outbound доставляется, остальное —
	// идемпотентный no-op.
	var fleetState string
	err := tx.QueryRow(ctx,
		`SELECT state FROM fleets WHERE id=$1 FOR UPDATE`,
		pl.FleetID,
	).Scan(&fleetState)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Флот удалён — груз уже доставлен ранее или отозван. Не падаем.
			slog.InfoContext(ctx, "event_delivery_artefacts_skip_no_fleet",
				slog.String("event_id", e.ID),
				slog.String("fleet_id", pl.FleetID))
			return nil
		}
		return fmt.Errorf("select fleet: %w", err)
	}
	if fleetState != "outbound" {
		slog.InfoContext(ctx, "event_delivery_artefacts_skip_fleet_state",
			slog.String("event_id", e.ID),
			slog.String("fleet_id", pl.FleetID),
			slog.String("state", fleetState))
		return nil
	}

	// Шаг 2: для каждого артефакта — переписать владельца. Идемпотентность
	// через сравнение текущих user_id/planet_id с целевыми (зеркалит
	// idempotency-паттерн demolish handler'а).
	transferred := make([]string, 0, len(pl.ArtefactIDs))
	for _, artID := range pl.ArtefactIDs {
		var (
			curUser   string
			curPlanet *string
			curState  string
		)
		err := tx.QueryRow(ctx, `
			SELECT user_id, planet_id, state
			FROM artefacts_user WHERE id=$1 FOR UPDATE
		`, artID).Scan(&curUser, &curPlanet, &curState)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Артефакт удалён до прибытия — не падаем (биржа могла
				// расформировать лот, см. exchange ttl план 68).
				slog.WarnContext(ctx, "event_delivery_artefacts_skip_no_artefact",
					slog.String("event_id", e.ID),
					slog.String("artefact_id", artID))
				continue
			}
			return fmt.Errorf("select artefact %s: %w", artID, err)
		}

		// Per-universe (R10): артефакт-источник и получатель должны быть
		// в одной вселенной. Защита от багов биржи плана 68. Сравнение
		// через JOIN users.universe_id.
		var sameUniverse bool
		if err := tx.QueryRow(ctx, `
			SELECT (
				SELECT universe_id FROM users WHERE id=$1
			) = (
				SELECT universe_id FROM users WHERE id=$2
			)
		`, curUser, *e.UserID).Scan(&sameUniverse); err != nil {
			return fmt.Errorf("check universe parity for artefact %s: %w", artID, err)
		}
		if !sameUniverse {
			return fmt.Errorf("delivery_artefacts: artefact %s sender %s and recipient %s in different universes",
				artID, curUser, *e.UserID)
		}

		// Идемпотентный skip: уже у получателя.
		var planetMatches bool
		if curPlanet != nil && *curPlanet == *e.PlanetID {
			planetMatches = true
		}
		if curUser == *e.UserID && planetMatches && curState != StateActiveArtefact {
			slog.InfoContext(ctx, "event_delivery_artefacts_skip_idempotent",
				slog.String("event_id", e.ID),
				slog.String("artefact_id", artID))
			continue
		}

		// Переписываем владельца + сбрасываем active → held (см. doc-комментарий
		// о упрощении: revert эффектов лежит на пересборке effect-стека).
		if _, err := tx.Exec(ctx, `
			UPDATE artefacts_user
			SET user_id   = $1,
			    planet_id = $2,
			    state     = CASE WHEN state = 'active' THEN 'held'::artefact_state ELSE state END,
			    activated_at = CASE WHEN state = 'active' THEN NULL ELSE activated_at END,
			    expire_at    = CASE WHEN state = 'active' THEN NULL ELSE expire_at END
			WHERE id = $3
		`, *e.UserID, *e.PlanetID, artID); err != nil {
			return fmt.Errorf("transfer artefact %s: %w", artID, err)
		}
		transferred = append(transferred, artID)
	}

	// Шаг 3: флот → returning (зеркалит ArriveHandler).
	if _, err := tx.Exec(ctx,
		`UPDATE fleets SET state='returning' WHERE id=$1`, pl.FleetID); err != nil {
		return fmt.Errorf("update fleet returning: %w", err)
	}

	slog.InfoContext(ctx, "event_delivery_artefacts_applied",
		slog.String("event_id", e.ID),
		slog.String("fleet_id", pl.FleetID),
		slog.String("planet_id_to", *e.PlanetID),
		slog.String("user_id_to", *e.UserID),
		slog.Int("artefact_count", len(pl.ArtefactIDs)),
		slog.Int("transferred_count", len(transferred)))
	return nil
}

// StateActiveArtefact — копия `artefact.StateActive`. Дублируем константу,
// чтобы пакет event не зависел от пакета artefact (избегаем циклов:
// artefact уже импортирует event для KindArtefactExpire/Delay).
const StateActiveArtefact = "active"

// HandleAllianceAttackAdditional — handler для KindAllianceAttackAdditional=30
// (план 65 Ф.5). В legacy origin (EventHandler.class.php:707-708) этот
// тип события — служебный referrer для основного EVENT_ATTACK_ALLIANCE
// (Kind=12): он маркирует «дополнительный флот, примыкающий к ACS-атаке»,
// и сам по себе ничего не делает (`case EVENT_ALLIANCE_ATTACK_ADDITIONAL: break`).
//
// **В nova ACS архитектурно иной**: все флоты группы получают одно и то
// же событие KindAttackAlliance с общим acs_group_id, и leader (первый
// по created_at) выполняет всю работу за группу — см. ACSAttackHandler
// в [internal/fleet/acs_attack.go]. Поэтому KindAllianceAttackAdditional
// в nova концептуально излишен — но мы регистрируем его как явный no-op
// для совместимости с возможной репликацией events из game-legacy-php
// (если когда-нибудь сделаем общую events-таблицу для legacy/nova) и
// чтобы события этого Kind'а не шли в StateError при импорте архива
// origin.
//
// Идемпотентность: тривиальная — handler ничего не меняет, повтор
// безопасен.
//
// **R15 уточнение**: это НЕ trade-off в simplifications.md — no-op
// handler в nova адекватно отражает no-op-семантику legacy. R8/R9/R12
// неприменимы (нет мутации, нет user-facing вывода). R3 audit — пишем
// info-slog для отладки (если событие появится — увидим в логах).
func HandleAllianceAttackAdditional(ctx context.Context, tx pgx.Tx, e Event) error {
	slog.InfoContext(ctx, "event_alliance_attack_additional_noop",
		slog.String("event_id", e.ID),
		slog.Int("kind", int(e.Kind)))
	return nil
}

// HandleDemolishConstruction понижает уровень здания на планете до
// TargetLevel (обычно curLevel-1, допускается 0 = полное удаление).
// Зеркалит HandleBuildConstruction.
//
// Семантика origin (EventHandler::demolish, EventHandler.class.php:2257):
//   - level > 0 → UPDATE building level = TargetLevel.
//   - level == 0 → DELETE строки building (легаси).
//
// В nova таблица buildings без UNIQUE-NOT-NULL на level, поэтому DELETE
// эквивалентен UPDATE level=0 (см. SELECT ниже — отсутствие строки
// читается как 0). Используем UPDATE: данные о факте «когда-то было
// построено» можно восстановить через events_dead. Альтернатива (DELETE)
// прокатилась бы, но усложняет audit-замеры «сколько построек игрок
// демонтировал за период».
//
// Идемпотентность: если cur <= TargetLevel — событие уже применено
// (или применено раньше воркером, или заявка перезатёрта новой). В этом
// случае только закрываем очередь и возвращаемся без ошибки.
//
// Поля поля планеты: при demolish здания **до 0** возвращаем 1 used_field
// (зеркалит HandleBuildConstruction:69). При понижении уровня (>0) — нет.
//
// Очки: пересчитываются батчем (KindScoreRecalcAll, Kind=70) — здесь не
// трогаем. Это отличается от legacy oxsar2 (инкремент UPDATE user.points
// в той же транзакции), но в nova очки derived state, восстанавливаемые
// из buildings/research/ships.
//
// Audit: пишем структурированный slog (R3) с полями event_id, planet_id,
// unit_id, level_from, level_to. Отдельной audit-таблицы для player-action
// в nova нет — slog уезжает в централизованный лог-агрегатор и достаточен
// для построения временного ряда «снос построек». Если понадобится
// SQL-доступ к истории — события остаются в events / events_dead с
// исходным payload.
func HandleDemolishConstruction(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl BuildingPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.PlanetID == nil {
		return fmt.Errorf("demolish event without planet_id")
	}
	if pl.TargetLevel < 0 {
		return fmt.Errorf("demolish target_level must be >=0, got %d", pl.TargetLevel)
	}

	var cur int
	err := tx.QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		*e.PlanetID, pl.UnitID,
	).Scan(&cur)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("select level: %w", err)
	}
	// Идемпотентность: уже применён demolish (или ниже).
	if cur <= pl.TargetLevel {
		_, _ = tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID)
		slog.InfoContext(ctx, "event_demolish_skip_idempotent",
			slog.String("event_id", e.ID),
			slog.String("planet_id", *e.PlanetID),
			slog.Int("unit_id", pl.UnitID),
			slog.Int("level_current", cur),
			slog.Int("level_target", pl.TargetLevel))
		return nil
	}

	if _, err := tx.Exec(ctx, `
		UPDATE buildings SET level=$3
		WHERE planet_id=$1 AND unit_id=$2
	`, *e.PlanetID, pl.UnitID, pl.TargetLevel); err != nil {
		return fmt.Errorf("downgrade building: %w", err)
	}

	// План 23 (зеркало HandleBuildConstruction): полностью снесённое
	// здание (target=0) освобождает поле планеты. Понижение уровня (target>0)
	// поле не освобождает — само здание остаётся.
	if pl.TargetLevel == 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET used_fields = GREATEST(used_fields - 1, 0) WHERE id = $1`,
			*e.PlanetID); err != nil {
			return fmt.Errorf("dec used_fields: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID); err != nil {
		return fmt.Errorf("close queue: %w", err)
	}

	slog.InfoContext(ctx, "event_demolish_applied",
		slog.String("event_id", e.ID),
		slog.String("planet_id", *e.PlanetID),
		slog.Int("unit_id", pl.UnitID),
		slog.Int("level_from", cur),
		slog.Int("level_to", pl.TargetLevel))
	return nil
}
