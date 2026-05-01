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
	"strconv"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/automsg"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/repo"
)

type Service struct {
	db      repo.Exec
	automsg *automsg.Service
	bundle  *i18n.Bundle
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

// WithAutoMsg подключает automsg для уведомления MSG_CREDIT_MARKET
// при покупке ресурса за кредиты (план 72.1.21).
func (s *Service) WithAutoMsg(am *automsg.Service) *Service {
	s.automsg = am
	return s
}

// WithBundle — i18n для текста AutoMsg.
func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
}

// userLang читает язык получателя AutoMsg.
func (s *Service) userLang(ctx context.Context, tx pgx.Tx, userID string) i18n.Lang {
	var lang string
	if tx != nil {
		_ = tx.QueryRow(ctx, `SELECT language FROM users WHERE id=$1`, userID).Scan(&lang)
	} else {
		_ = s.db.Pool().QueryRow(ctx, `SELECT language FROM users WHERE id=$1`, userID).Scan(&lang)
	}
	if lang == "" {
		return i18n.LangRu
	}
	return i18n.Lang(lang)
}

var (
	ErrInvalidResource  = errors.New("market: invalid resource")
	ErrSameResource     = errors.New("market: from == to")
	ErrInvalidAmount    = errors.New("market: amount must be > 0")
	ErrNotEnough        = errors.New("market: not enough resource on planet")
	ErrPlanetOwnership  = errors.New("market: planet not owned by user")
	ErrPlanetNotFound   = errors.New("market: planet not found")
	ErrLotNotFound      = errors.New("market: lot not found")
	ErrLotNotOpen       = errors.New("market: lot is not open")
	ErrOwnLot           = errors.New("market: cannot accept own lot")
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

// CreditRates — стоимость 1 кредита в ресурсах (condition unit).
// 1 credit = 100 condition units (т.е. 100 metal / 50 silicon / 25 hydrogen).
const CreditRatePerUnit = 100.0

// ExchangeCredit покупает ресурс за кредиты (premium → ресурсы).
// resource: metal|silicon|hydrogen
// amount: количество кредитов, которое тратится.
//
// Обратное направление (продажа ресурсов за кредиты, бывшее direction
// "to_credit") удалено 2026-04-26 как уязвимость: позволяло бесконечно
// фармить premium-валюту через производство ресурсов. См. ADR/dev-log.
//
// Поле Direction в ответе сохранено для совместимости с frontend и теперь
// всегда равно "from_credit".
type CreditExchangeResult struct {
	Direction     string  `json:"direction"`
	Resource      string  `json:"resource"`
	ResourceDelta int64   `json:"resource_delta"`
	CreditDelta   float64 `json:"credit_delta"`
}

func (s *Service) ExchangeCredit(ctx context.Context, userID, planetID, direction, resource string, amount float64) (CreditExchangeResult, error) {
	cost, ok := resourceCost[resource]
	if !ok {
		return CreditExchangeResult{}, ErrInvalidResource
	}
	if amount <= 0 {
		return CreditExchangeResult{}, ErrInvalidAmount
	}
	// Удалено направление "to_credit" (продажа ресурсов за кредиты).
	// Принимаем только "from_credit" или пустую строку (default = from_credit
	// для обратной совместимости со старым клиентом, ещё не обновлённым).
	if direction != "" && direction != "from_credit" {
		return CreditExchangeResult{}, ErrInvalidResource
	}

	var out CreditExchangeResult
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверка владельца.
		var ownerID string
		var balance float64
		err := tx.QueryRow(ctx,
			`SELECT user_id, `+resource+` FROM planets WHERE id = $1 FOR UPDATE`,
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
		_ = balance // нужен для FOR UPDATE-lock на planets row

		var userRate, credit float64
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(exchange_rate, 1.2), credit FROM users WHERE id = $1`, userID).Scan(&userRate, &credit); err != nil {
			return fmt.Errorf("market: read user: %w", err)
		}
		if userRate <= 0 {
			userRate = 1.2
		}

		// from_credit: amount — количество кредитов, покупаем ресурс.
		if credit < amount {
			return ErrNotEnough
		}
		resAmount := int64(math.Floor(amount * CreditRatePerUnit / cost / userRate))
		if resAmount <= 0 {
			return ErrInvalidAmount
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET credit = credit - $1 WHERE id = $2`,
			amount, userID); err != nil {
			return fmt.Errorf("market: debit user: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+resource+` = `+resource+` + $1 WHERE id = $2`,
			resAmount, planetID); err != nil {
			return fmt.Errorf("market: credit planet: %w", err)
		}

		// План 72.1.21: AutoMsg `MSG_CREDIT_MARKET` (legacy
		// `Market.class.php::Credit_ex` строки 250-258) в folder=8
		// (FolderCredit). Best-effort — ошибка глотается, чтобы не
		// откатывать саму покупку.
		if s.automsg != nil && s.bundle != nil && amount > 0 {
			lang := s.userLang(ctx, tx, userID)
			vars := map[string]string{
				"credits":  strconv.FormatFloat(amount, 'f', 0, 64),
				"resource": resource,
				"amount":   strconv.FormatInt(resAmount, 10),
			}
			title := s.bundle.Tr(lang, "autoMessages", "creditMarketPurchase.title", vars)
			body := s.bundle.Tr(lang, "autoMessages", "creditMarketPurchase.body", vars)
			_ = s.automsg.SendDirect(ctx, tx, userID, automsg.FolderCredit, title, body)
		}

		out = CreditExchangeResult{Direction: "from_credit", Resource: resource, ResourceDelta: resAmount, CreditDelta: -amount}
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

// -----------------------------------------------------------------
// Ордерная книга (market_lots)
// -----------------------------------------------------------------

// Lot — запись в market_lots.
type Lot struct {
	ID           string  `json:"id"`
	SellerID     string  `json:"seller_id"`
	SellerName   string  `json:"seller_name"`
	PlanetID     string  `json:"planet_id"`
	SellResource string  `json:"sell_resource"`
	SellAmount   int64   `json:"sell_amount"`
	BuyResource  string  `json:"buy_resource"`
	BuyAmount    int64   `json:"buy_amount"`
	State        string  `json:"state"`
	CreatedAt    string  `json:"created_at"`
}

func validResource(r string) bool {
	return r == "metal" || r == "silicon" || r == "hydrogen"
}

// ListLots возвращает открытые лоты, опционально фильтруя по sell_resource.
// Фильтр kind='resource' добавлен после миграции 0055, чтобы флотовые
// лоты выдавались отдельным API.
func (s *Service) ListLots(ctx context.Context, sellResource string, limit int) ([]Lot, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `
		SELECT ml.id, ml.seller_id, COALESCE(u.username,''),
		       ml.planet_id, ml.sell_resource, ml.sell_amount,
		       ml.buy_resource, ml.buy_amount, ml.state,
		       ml.created_at
		FROM market_lots ml
		LEFT JOIN users u ON u.id = ml.seller_id
		WHERE ml.state = 'open' AND ml.kind = 'resource'`
	args := []any{limit}
	if sellResource != "" && validResource(sellResource) {
		query += ` AND ml.sell_resource = $2`
		args = append([]any{sellResource, limit}, args[1:]...)
		query += ` ORDER BY ml.created_at DESC LIMIT $2`
	} else {
		query += ` ORDER BY ml.created_at DESC LIMIT $1`
	}

	rows, err := s.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("market lots list: %w", err)
	}
	defer rows.Close()
	var out []Lot
	for rows.Next() {
		var l Lot
		var createdAt interface{}
		if err := rows.Scan(&l.ID, &l.SellerID, &l.SellerName,
			&l.PlanetID, &l.SellResource, &l.SellAmount,
			&l.BuyResource, &l.BuyAmount, &l.State, &createdAt); err != nil {
			return nil, err
		}
		l.CreatedAt = fmt.Sprintf("%v", createdAt)
		out = append(out, l)
	}
	return out, rows.Err()
}

// CreateLot создаёт новый лот. Ресурс списывается с планеты продавца сразу (escrow).
func (s *Service) CreateLot(ctx context.Context, userID, planetID,
	sellResource string, sellAmount int64, buyResource string, buyAmount int64) (Lot, error) {
	if !validResource(sellResource) || !validResource(buyResource) {
		return Lot{}, ErrInvalidResource
	}
	if sellResource == buyResource {
		return Lot{}, ErrSameResource
	}
	if sellAmount <= 0 || buyAmount <= 0 {
		return Lot{}, ErrInvalidAmount
	}

	var out Lot
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var ownerID string
		var balance float64
		err := tx.QueryRow(ctx,
			`SELECT user_id, `+sellResource+` FROM planets WHERE id=$1 FOR UPDATE`,
			planetID).Scan(&ownerID, &balance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPlanetNotFound
			}
			return fmt.Errorf("market lot: read planet: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
		}
		if int64(balance) < sellAmount {
			return ErrNotEnough
		}
		// Списываем ресурс с планеты (escrow).
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+sellResource+` = `+sellResource+` - $1 WHERE id = $2`,
			sellAmount, planetID); err != nil {
			return fmt.Errorf("market lot: debit: %w", err)
		}

		var lotID string
		err = tx.QueryRow(ctx, `
			INSERT INTO market_lots (seller_id, planet_id, sell_resource, sell_amount,
			                         buy_resource, buy_amount)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, userID, planetID, sellResource, sellAmount, buyResource, buyAmount).Scan(&lotID)
		if err != nil {
			return fmt.Errorf("market lot: insert: %w", err)
		}

		var sellerName string
		_ = tx.QueryRow(ctx, `SELECT username FROM users WHERE id=$1`, userID).Scan(&sellerName)
		out = Lot{
			ID: lotID, SellerID: userID, SellerName: sellerName,
			PlanetID: planetID, SellResource: sellResource, SellAmount: sellAmount,
			BuyResource: buyResource, BuyAmount: buyAmount, State: "open",
		}
		return nil
	})
	return out, err
}

// CancelLot отменяет лот (только продавец). Возвращает ресурс на планету.
func (s *Service) CancelLot(ctx context.Context, userID, lotID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var sellerID, planetID, sellResource string
		var sellAmount int64
		var state string
		err := tx.QueryRow(ctx, `
			SELECT seller_id, planet_id, sell_resource, sell_amount, state
			FROM market_lots WHERE id=$1 FOR UPDATE
		`, lotID).Scan(&sellerID, &planetID, &sellResource, &sellAmount, &state)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrLotNotFound
			}
			return fmt.Errorf("market cancel: read lot: %w", err)
		}
		if sellerID != userID {
			return ErrPlanetOwnership
		}
		if state != "open" {
			return ErrLotNotOpen
		}
		if _, err := tx.Exec(ctx,
			`UPDATE market_lots SET state='cancelled', updated_at=now() WHERE id=$1`, lotID); err != nil {
			return fmt.Errorf("market cancel: update: %w", err)
		}
		// Возвращаем escrow на планету.
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+sellResource+` = `+sellResource+` + $1 WHERE id = $2`,
			sellAmount, planetID); err != nil {
			return fmt.Errorf("market cancel: refund: %w", err)
		}
		return nil
	})
}

// AcceptLot — покупатель принимает лот. Деньги списываются с его планеты,
// ресурс переходит от продавца к покупателю.
func (s *Service) AcceptLot(ctx context.Context, buyerID, buyerPlanetID, lotID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var sellerID, sellerPlanetID, sellResource, buyResource string
		var sellAmount, buyAmount int64
		var state string
		err := tx.QueryRow(ctx, `
			SELECT seller_id, planet_id, sell_resource, sell_amount,
			       buy_resource, buy_amount, state
			FROM market_lots WHERE id=$1 FOR UPDATE
		`, lotID).Scan(&sellerID, &sellerPlanetID, &sellResource, &sellAmount,
			&buyResource, &buyAmount, &state)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrLotNotFound
			}
			return fmt.Errorf("market accept: read lot: %w", err)
		}
		if sellerID == buyerID {
			return ErrOwnLot
		}
		if state != "open" {
			return ErrLotNotOpen
		}

		// Проверяем баланс покупателя.
		var buyerOwner string
		var buyerBalance float64
		err = tx.QueryRow(ctx,
			`SELECT user_id, `+buyResource+` FROM planets WHERE id=$1 FOR UPDATE`,
			buyerPlanetID).Scan(&buyerOwner, &buyerBalance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPlanetNotFound
			}
			return fmt.Errorf("market accept: read buyer planet: %w", err)
		}
		if buyerOwner != buyerID {
			return ErrPlanetOwnership
		}
		if int64(buyerBalance) < buyAmount {
			return ErrNotEnough
		}

		// Списываем buyResource с покупателя → на планету продавца.
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+buyResource+` = `+buyResource+` - $1 WHERE id=$2`,
			buyAmount, buyerPlanetID); err != nil {
			return fmt.Errorf("market accept: debit buyer: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+buyResource+` = `+buyResource+` + $1 WHERE id=$2`,
			buyAmount, sellerPlanetID); err != nil {
			return fmt.Errorf("market accept: credit seller buy_resource: %w", err)
		}
		// sellResource уже в escrow (списан при создании) → зачисляем покупателю.
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+sellResource+` = `+sellResource+` + $1 WHERE id=$2`,
			sellAmount, buyerPlanetID); err != nil {
			return fmt.Errorf("market accept: credit buyer sell_resource: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE market_lots SET state='accepted', buyer_id=$1, updated_at=now()
			WHERE id=$2
		`, buyerID, lotID); err != nil {
			return fmt.Errorf("market accept: update lot: %w", err)
		}
		return nil
	})
}
