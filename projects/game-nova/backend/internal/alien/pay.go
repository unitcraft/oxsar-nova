package alien

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/repo"
)

// Ошибки PayHolding — типизированы, чтобы HTTP-слой мог отличить их от
// внутренних сбоев и ответить корректным статусом.
var (
	ErrHoldingNotFound   = errors.New("alien holding not found or not active")
	ErrHoldingNotOwner   = errors.New("holding belongs to another user")
	ErrInsufficientCred  = errors.New("insufficient credits")
	ErrHoldingAtCap      = errors.New("holding already at max real-time cap")
	ErrPayAmountInvalid  = errors.New("pay amount must be positive")
)

// PayResult — успешный результат продления HOLDING платежом.
type PayResult struct {
	HoldingEventID  string    `json:"holding_event_id"`
	PaidThisTime    int64     `json:"paid_this_time"`
	PaidTotal       int64     `json:"paid_total"`
	PaidTimes       int       `json:"paid_times"`
	NewFireAt       time.Time `json:"new_fire_at"`
	CappedAt        time.Time `json:"capped_at"`  // start + 15 дней
	CreditRemaining int64     `json:"credit_remaining"`
}

// PayHolding — продлить HOLDING-событие платежом в `amount` кредитов.
// Формула (legacy AlienAI.class.php:993): fire_at += 2h * amount / 50,
// cap = start + ALIEN_HALTING_MAX_REAL_TIME (15 дней).
//
// Транзакционно: списываем кредиты, обновляем fire_at, инкрементируем
// PaidCredit/PaidTimes в payload. Если fire_at уже на cap — ErrHoldingAtCap
// и списания не происходит. Если у игрока меньше `amount` кредитов —
// ErrInsufficientCred.
func PayHolding(ctx context.Context, db repo.Exec, userID, holdingEventID string, amount int64) (*PayResult, error) {
	if amount <= 0 {
		return nil, ErrPayAmountInvalid
	}
	var result *PayResult
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Загружаем событие FOR UPDATE.
		var kind int
		var state string
		var fireAt time.Time
		var eventUserID string
		var payload []byte
		err := tx.QueryRow(ctx, `
			SELECT kind, state, fire_at, user_id, payload
			FROM events WHERE id = $1 FOR UPDATE
		`, holdingEventID).Scan(&kind, &state, &fireAt, &eventUserID, &payload)
		if err == pgx.ErrNoRows {
			return ErrHoldingNotFound
		}
		if err != nil {
			return fmt.Errorf("pay holding: load: %w", err)
		}
		if kind != int(event.KindAlienHolding) || state != string(event.StateWait) {
			return ErrHoldingNotFound
		}
		if eventUserID != userID {
			return ErrHoldingNotOwner
		}

		var hp holdingPayload
		if err := json.Unmarshal(payload, &hp); err != nil {
			return fmt.Errorf("pay holding: parse payload: %w", err)
		}

		capTime := hp.StartTime.Add(AlienHaltingMaxRealTime)
		if !fireAt.Before(capTime) {
			return ErrHoldingAtCap
		}

		// Баланс игрока.
		var credit int64
		if err := tx.QueryRow(ctx,
			`SELECT credit FROM users WHERE id = $1 FOR UPDATE`, userID,
		).Scan(&credit); err != nil {
			return fmt.Errorf("pay holding: load credit: %w", err)
		}
		if credit < amount {
			return ErrInsufficientCred
		}

		// Продление: Δt = amount * 144 сек.
		extension := time.Duration(float64(amount)*AlienHoldingPaySecondsPerCredit) * time.Second
		newFireAt := fireAt.Add(extension)
		if newFireAt.After(capTime) {
			newFireAt = capTime
		}

		// Списание.
		if _, err := tx.Exec(ctx,
			`UPDATE users SET credit = credit - $1 WHERE id = $2`,
			amount, userID); err != nil {
			return fmt.Errorf("pay holding: debit credit: %w", err)
		}

		// Обновление payload.
		hp.PaidCredit += amount
		hp.PaidTimes++
		newPayload, err := json.Marshal(hp)
		if err != nil {
			return fmt.Errorf("pay holding: marshal payload: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE events SET fire_at = $1, payload = $2 WHERE id = $3
		`, newFireAt, newPayload, holdingEventID); err != nil {
			return fmt.Errorf("pay holding: update event: %w", err)
		}

		result = &PayResult{
			HoldingEventID:  holdingEventID,
			PaidThisTime:    amount,
			PaidTotal:       hp.PaidCredit,
			PaidTimes:       hp.PaidTimes,
			NewFireAt:       newFireAt,
			CappedAt:        capTime,
			CreditRemaining: credit - amount,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
