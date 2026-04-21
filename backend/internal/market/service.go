// Package market — обмен ресурсов (metal ↔ silicon ↔ hydrogen).
//
// Базовые «масштабные» курсы (ogame classic): 1 metal ценнее silicon
// в 2 раза, silicon ценнее hydrogen в 2 раза. Т.е. «стоимость» ресурсов:
//     metal=1, silicon=2, hydrogen=4
// Курс X→Y: amountY = amountX × cost(X) / cost(Y) × exchange_rate(user).
//
// exchange_rate у игрока — 1.0 по умолчанию «справедливой» позиции,
// но в oxsar2 default 1.2 (чуть хуже чем 1:1) — без бонуса, а
// артефакт MERCHANTS_MARK понижает до ~1.0. Это означает: игрок
// теряет ~16.7% при каждом обмене (1/1.2 ≈ 0.83), бонус от
// артефакта сокращает потерю.
//
// Отдельной таблицы rate нет — коэффициенты жёсткие. В будущем
// (M6 full-exchange из legacy) заменим на ордерную книгу.
package market

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
)

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

var (
	ErrInvalidResource  = errors.New("market: invalid resource")
	ErrSameResource     = errors.New("market: from == to")
	ErrInvalidAmount    = errors.New("market: amount must be > 0")
	ErrNotEnough        = errors.New("market: not enough resource on planet")
	ErrPlanetOwnership  = errors.New("market: planet not owned by user")
	ErrPlanetNotFound   = errors.New("market: planet not found")
)

// resourceCost — «стоимость» 1 единицы ресурса в условных единицах.
// metal — базис (1), silicon — в 2 раза ценнее по использованию как
// metal, hydrogen — в 4 раза (в OGame для space-science). Логика
// обмена: «из чего» × coefIn = «во что» × coefOut (цены равны).
var resourceCost = map[string]float64{
	"metal":    1.0,
	"silicon":  2.0,
	"hydrogen": 4.0,
}

// ExchangeResult — итог обмена.
type ExchangeResult struct {
	FromResource string  `json:"from"`
	ToResource   string  `json:"to"`
	FromAmount   int64   `json:"from_amount"`
	ToAmount     int64   `json:"to_amount"`
	Rate         float64 `json:"rate"` // applied (включая user.exchange_rate)
}

// Exchange списывает fromAmount ресурса `from` с планеты, зачисляет
// toAmount ресурса `to`. Ошибки на неверных ресурсах, неверном
// amount, недостаточном балансе или чужой планете.
func (s *Service) Exchange(ctx context.Context, userID, planetID string,
	from, to string, fromAmount int64) (ExchangeResult, error) {
	fromCost, ok1 := resourceCost[from]
	toCost, ok2 := resourceCost[to]
	if !ok1 || !ok2 {
		return ExchangeResult{}, ErrInvalidResource
	}
	if from == to {
		return ExchangeResult{}, ErrSameResource
	}
	if fromAmount <= 0 {
		return ExchangeResult{}, ErrInvalidAmount
	}

	var out ExchangeResult
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверка владельца + текущий баланс.
		var ownerID string
		var balance float64
		col := "metal"
		switch from {
		case "metal":
			col = "metal"
		case "silicon":
			col = "silicon"
		case "hydrogen":
			col = "hydrogen"
		}
		err := tx.QueryRow(ctx,
			`SELECT user_id, `+col+` FROM planets WHERE id = $1 FOR UPDATE`,
			planetID).Scan(&ownerID, &balance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPlanetNotFound
			}
			return fmt.Errorf("market: read planet: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
		}
		if int64(balance) < fromAmount {
			return ErrNotEnough
		}

		// user.exchange_rate (default 1.2 — «штраф», <1.0 означает
		// бонус от MERCHANTS_MARK).
		var userRate float64
		if err := tx.QueryRow(ctx,
			`SELECT exchange_rate FROM users WHERE id = $1`, userID).Scan(&userRate); err != nil {
			return fmt.Errorf("market: read user rate: %w", err)
		}
		if userRate <= 0 {
			userRate = 1.2
		}

		// toAmount = fromAmount × (fromCost / toCost) / userRate.
		// При userRate=1.2 игрок получает меньше «выгодного» ресурса,
		// при userRate=1.0 — честный паритет стоимостей.
		toAmount := float64(fromAmount) * fromCost / toCost / userRate
		toAmountInt := int64(math.Floor(toAmount))
		if toAmountInt <= 0 {
			return ErrInvalidAmount
		}

		// Списание/зачисление.
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+from+` = `+from+` - $1 WHERE id = $2`,
			fromAmount, planetID); err != nil {
			return fmt.Errorf("market: debit: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+to+` = `+to+` + $1 WHERE id = $2`,
			toAmountInt, planetID); err != nil {
			return fmt.Errorf("market: credit: %w", err)
		}
		// Явно считаем delta-столбцы для res_log: списание → отрицательное,
		// зачисление → положительное, остальные — 0.
		dm, dsi, dh := int64(0), int64(0), int64(0)
		switch from {
		case "metal":
			dm -= fromAmount
		case "silicon":
			dsi -= fromAmount
		case "hydrogen":
			dh -= fromAmount
		}
		switch to {
		case "metal":
			dm += toAmountInt
		case "silicon":
			dsi += toAmountInt
		case "hydrogen":
			dh += toAmountInt
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason,
			                     delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'market', $3, $4, $5)
		`, userID, planetID, dm, dsi, dh); err != nil {
			return fmt.Errorf("market: res_log: %w", err)
		}

		out = ExchangeResult{
			FromResource: from,
			ToResource:   to,
			FromAmount:   fromAmount,
			ToAmount:     toAmountInt,
			Rate:         fromCost / toCost / userRate,
		}
		return nil
	})
	return out, err
}

// Rates возвращает текущий набор курсов (для UI — «1 M = N Si = M H»).
type Rates struct {
	Metal    float64 `json:"metal"`
	Silicon  float64 `json:"silicon"`
	Hydrogen float64 `json:"hydrogen"`
	UserRate float64 `json:"user_rate"` // exchange_rate текущего юзера
}

func (s *Service) Rates(ctx context.Context, userID string) (Rates, error) {
	var r Rates
	r.Metal = resourceCost["metal"]
	r.Silicon = resourceCost["silicon"]
	r.Hydrogen = resourceCost["hydrogen"]
	err := s.db.Pool().QueryRow(ctx,
		`SELECT exchange_rate FROM users WHERE id = $1`, userID).Scan(&r.UserRate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.UserRate = 1.2
			return r, nil
		}
		return r, fmt.Errorf("market: read user rate: %w", err)
	}
	return r, nil
}
