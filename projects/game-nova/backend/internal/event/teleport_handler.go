package event

// План 65 Ф.6 (D-032+U-009): event-handler KindTeleportPlanet.
//
// HTTP-handler в internal/planet/teleport_handler.go вставляет event с
// payload {target_galaxy, target_system, target_position, cost_oxsars,
// idempotency_key} и fire_at = now + duration. Этот handler срабатывает
// в worker'е по fire_at и применяет смещение координат.
//
// Семантика:
//
//  1. SELECT planet (FOR UPDATE) — текущие галактика/система/позиция,
//     is_moon, ownership. Если planet удалена (destroyed_at) — no-op +
//     refund (event пришёл с пустой целью).
//  2. SELECT target slot. Если занят (другой игрок успел встать на
//     координаты между HTTP-handler'ом и срабатыванием event'а) —
//     refund + audit + return nil (no-op, транзакция должна commit'нуться,
//     чтобы не зациклить retry).
//  3. UPDATE planets SET galaxy=?, system=?, position=?.
//  4. UPDATE users SET last_planet_teleport_at = now (cooldown
//     активируется в момент срабатывания, не в момент HTTP-call'а —
//     иначе при отказе телепорта игрок «теряет» cooldown без оплаты).
//  5. Audit slog (R3).
//
// Идемпотентность: повторный запуск handler'а с тем же event_id
// (что возможно при крэше воркера между Update'ом и Commit'ом) должен
// быть безопасен. Для этого сравниваем текущие координаты планеты с
// payload'ом: если они уже совпадают → no-op.
//
// Refund: вызывается через TeleportRefunder, передаваемый при сборке
// handler'а в worker. Это позволяет тестам подменять billing-client
// на мок, а package event — не зависеть от internal/billing/client.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// TeleportPlanetPayload — payload события KindTeleportPlanet (план 65 Ф.6).
//
// Поля:
//   - TargetGalaxy/TargetSystem/TargetPosition — куда телепортировать
//     планету. Допустимые диапазоны валидируются HTTP-handler'ом
//     (planet/teleport_handler.go).
//   - CostOxsars — сколько было списано через billing.Spend на этапе
//     планирования. Хранится в payload, чтобы Refund мог использовать
//     ту же сумму при отказе.
//   - IdempotencyKey — Idempotency-Key из HTTP-запроса (передан в
//     billing.Spend). Используется для построения refund-ключа
//     (IdempotencyKey + ":refund").
type TeleportPlanetPayload struct {
	TargetGalaxy   int    `json:"target_galaxy"`
	TargetSystem   int    `json:"target_system"`
	TargetPosition int    `json:"target_position"`
	CostOxsars     int64  `json:"cost_oxsars"`
	IdempotencyKey string `json:"idempotency_key"`
}

// TeleportRefunder — callback для возврата оксаров при отказе телепорта
// (target slot занят на момент срабатывания event'а, planet удалена и т.п.).
//
// Реализация — в worker.go (internal/billing/client). Если refunder=nil
// (тесты или dev без billing), refund молча пропускается, а в slog
// пишется warning — это безопасное поведение, потому что billing
// сам по себе reconcile-friendly.
type TeleportRefunder func(ctx context.Context, userID, planetID string, payload TeleportPlanetPayload) error

// HandleTeleportPlanet возвращает Handler для регистрации в worker'е.
//
// Параметр refunder может быть nil — тогда refund-вызовы молча
// пропускаются (см. doc TeleportRefunder).
func HandleTeleportPlanet(refunder TeleportRefunder) Handler {
	return func(ctx context.Context, tx pgx.Tx, e Event) error {
		if e.UserID == nil {
			return fmt.Errorf("teleport event without user_id")
		}
		if e.PlanetID == nil {
			return fmt.Errorf("teleport event without planet_id")
		}

		var pl TeleportPlanetPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("parse teleport payload: %w", err)
		}

		userID, planetID := *e.UserID, *e.PlanetID

		var (
			ownerID          string
			curG, curS, curP int
			isMoon           bool
		)
		err := tx.QueryRow(ctx, `
			SELECT user_id, galaxy, system, position, is_moon
			FROM planets
			WHERE id = $1 AND destroyed_at IS NULL
			FOR UPDATE
		`, planetID).Scan(&ownerID, &curG, &curS, &curP, &isMoon)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Планета удалена — refund + audit, событие закрываем (return nil).
				slog.WarnContext(ctx, "teleport_handler_planet_missing_refunding",
					slog.String("event_id", e.ID),
					slog.String("user_id", userID),
					slog.String("planet_id", planetID))
				safeRefund(ctx, refunder, userID, planetID, pl, "planet_missing")
				return nil
			}
			return fmt.Errorf("select planet: %w", err)
		}
		if ownerID != userID {
			// Это не должно происходить (ownership проверяется при
			// создании event'а), но если случилось — refund и no-op.
			slog.ErrorContext(ctx, "teleport_handler_owner_mismatch",
				slog.String("event_id", e.ID),
				slog.String("event_user_id", userID),
				slog.String("planet_owner", ownerID),
				slog.String("planet_id", planetID))
			safeRefund(ctx, refunder, userID, planetID, pl, "owner_mismatch")
			return nil
		}

		// Идемпотентность: координаты уже совпадают с целевыми → handler
		// уже отработал ранее (крэш между UPDATE и commit'ом).
		if curG == pl.TargetGalaxy && curS == pl.TargetSystem && curP == pl.TargetPosition {
			slog.InfoContext(ctx, "teleport_handler_idempotent_skip",
				slog.String("event_id", e.ID),
				slog.String("planet_id", planetID))
			return nil
		}

		// Target slot occupied? Уникальный constraint в planets
		// (galaxy, system, position, is_moon) гарантирует, что INSERT
		// дважды на тот же slot невозможен; здесь явная проверка позволяет
		// избежать SQL-violation и сделать чистый refund.
		var existsID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM planets
			WHERE galaxy = $1 AND system = $2 AND position = $3 AND is_moon = $4
			  AND destroyed_at IS NULL AND id <> $5
			LIMIT 1
		`, pl.TargetGalaxy, pl.TargetSystem, pl.TargetPosition, isMoon, planetID).Scan(&existsID)
		if err == nil {
			slog.WarnContext(ctx, "teleport_handler_target_occupied_refunding",
				slog.String("event_id", e.ID),
				slog.String("user_id", userID),
				slog.String("planet_id", planetID),
				slog.String("occupied_by", existsID),
				slog.Int("target_galaxy", pl.TargetGalaxy),
				slog.Int("target_system", pl.TargetSystem),
				slog.Int("target_position", pl.TargetPosition))
			safeRefund(ctx, refunder, userID, planetID, pl, "target_occupied")
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("check target slot: %w", err)
		}

		// Применяем телепорт.
		if _, err := tx.Exec(ctx, `
			UPDATE planets
			SET galaxy = $1, system = $2, position = $3
			WHERE id = $4
		`, pl.TargetGalaxy, pl.TargetSystem, pl.TargetPosition, planetID); err != nil {
			return fmt.Errorf("update planet coords: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE users SET last_planet_teleport_at = now() WHERE id = $1
		`, userID); err != nil {
			return fmt.Errorf("update user cooldown: %w", err)
		}

		slog.InfoContext(ctx, "event_planet_teleported",
			slog.String("event_id", e.ID),
			slog.String("user_id", userID),
			slog.String("planet_id", planetID),
			slog.Int("from_galaxy", curG), slog.Int("from_system", curS), slog.Int("from_position", curP),
			slog.Int("to_galaxy", pl.TargetGalaxy), slog.Int("to_system", pl.TargetSystem), slog.Int("to_position", pl.TargetPosition),
			slog.Int64("cost_oxsars", pl.CostOxsars))
		return nil
	}
}

// safeRefund — nil-safe refund-вызов. Логирует warning при ошибке/nil-refunder,
// но не возвращает её (event-handler уже принял решение «отменить телепорт»,
// блокировать его на refund-сбое нельзя, иначе зависнем в retry-loop'е).
func safeRefund(ctx context.Context, refunder TeleportRefunder, userID, planetID string, pl TeleportPlanetPayload, reason string) {
	if refunder == nil {
		slog.WarnContext(ctx, "teleport_refund_skipped_no_refunder",
			slog.String("user_id", userID),
			slog.String("planet_id", planetID),
			slog.String("reason", reason),
			slog.Int64("cost_oxsars", pl.CostOxsars))
		return
	}
	if err := refunder(ctx, userID, planetID, pl); err != nil {
		slog.ErrorContext(ctx, "teleport_refund_failed",
			slog.String("user_id", userID),
			slog.String("planet_id", planetID),
			slog.String("reason", reason),
			slog.String("err", err.Error()))
	}
}
