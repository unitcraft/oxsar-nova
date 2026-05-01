// Package exchange — BrokerStats: статистика лотов брокера за период
// (закрывает legacy `Exchange.class.php::showStatistics`).
//
// Plan: 72.1 §20.12 P2P-биржа task. Origin не имел экрана статистики
// «мои проданные лоты» — добавляем его как `/p2p-exchange`. Backend
// агрегирует exchange_lots WHERE seller_user_id=$user AND
// sold_at∈[date_min, date_max] AND status IN (sold/cancelled/expired);
// возвращает list лотов с пагинацией (USER_PER_PAGE=25 как legacy)
// + summary (total, sold, turnover, profit).
//
// Profit считается так же как legacy: price * fee% / 100. Для MVP
// используем фиксированный fee=5% (legacy брокер настраивает в
// exchange-таблице, но в nova таблицы exchange нет — это упрощение
// будет закрываться в будущем плане «брокер-настройки»).

package exchange

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// BrokerStatsRow — строка таблицы статистики.
type BrokerStatsRow struct {
	LotID         string    `json:"lot_id"`
	UnitID        int       `json:"unit_id"`
	Quantity      int       `json:"quantity"`
	Price         int64     `json:"price"`
	Status        string    `json:"status"`        // sold|cancelled|expired
	SoldAt        time.Time `json:"sold_at"`
	Profit        float64   `json:"profit"`        // price × fee / 100 если sold, иначе 0
}

// BrokerStatsSummary — агрегаты за период.
type BrokerStatsSummary struct {
	Total    int     `json:"total"`     // всего записей
	Sold     int     `json:"sold"`      // только sold
	Turnover int64   `json:"turnover"`  // Σ price для sold
	Profit   float64 `json:"profit"`    // Σ price × fee / 100 для sold
}

// BrokerStatsFilters — параметры запроса.
type BrokerStatsFilters struct {
	UserID    string
	DateMin   time.Time
	DateMax   time.Time
	SortField string // date | lot | lot_price | lot_amount | lot_profit
	SortOrder string // asc | desc
	Page      int    // 1-indexed
	PerPage   int    // default 25 (USER_PER_PAGE)
}

// DefaultBrokerFee — дефолтный fee % для брокеров без custom-настроек.
// Legacy `exchange.comission` создаётся со значением 5.0 при первом
// обращении (Exchange.class.php).
const DefaultBrokerFee = 5.0

// BrokerSettings — план 72.1.45 §8: настройки брокера (legacy `exchange`
// таблица). user_id = брокер, fee_percent = его комиссия.
type BrokerSettings struct {
	UserID     string  `json:"user_id"`
	Title      string  `json:"title"`
	FeePercent float64 `json:"fee_percent"`
}

// GetBrokerSettings возвращает настройки брокера. Если строки нет —
// создаёт со значениями по умолчанию (legacy-поведение).
func (s *Service) GetBrokerSettings(ctx context.Context, userID string) (BrokerSettings, error) {
	var bs BrokerSettings
	bs.UserID = userID
	err := s.db.Pool().QueryRow(ctx, `
		SELECT title, fee_percent
		  FROM exchange_settings
		 WHERE user_id = $1
	`, userID).Scan(&bs.Title, &bs.FeePercent)
	if err == nil {
		return bs, nil
	}
	// Не нашли — вставляем дефолт и возвращаем.
	bs.Title = "My exchange"
	bs.FeePercent = DefaultBrokerFee
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO exchange_settings (user_id, title, fee_percent)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO NOTHING
	`, userID, bs.Title, bs.FeePercent)
	if err != nil {
		return BrokerSettings{}, fmt.Errorf("init broker settings: %w", err)
	}
	return bs, nil
}

// getBrokerSettingsTx — план 72.1.46 P1#1: вариант для использования
// внутри открытой транзакции BuyLot. Не создаёт записи (read-only) —
// если нет строки, возвращает дефолт без INSERT.
//
// В unit-тестах с fakeRepo `tx` может быть nil — в этом случае
// возвращаем дефолт (тесты не покрывают per-broker fee, для них fee=0
// эквивалентно отсутствию изменения seller_profit).
func (s *Service) getBrokerSettingsTx(ctx context.Context, tx pgx.Tx, userID string) (BrokerSettings, error) {
	bs := BrokerSettings{UserID: userID, Title: "My exchange", FeePercent: DefaultBrokerFee}
	if tx == nil {
		return bs, nil
	}
	err := tx.QueryRow(ctx, `
		SELECT title, fee_percent
		  FROM exchange_settings
		 WHERE user_id = $1
	`, userID).Scan(&bs.Title, &bs.FeePercent)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return BrokerSettings{}, fmt.Errorf("read broker settings: %w", err)
	}
	return bs, nil
}

// UpdateBrokerSettings — план 72.1.45 §9: ExchangeOpts admin страница.
// Брокер обновляет свой title и fee_percent (clamp 0..50).
func (s *Service) UpdateBrokerSettings(ctx context.Context, userID, title string, feePercent float64) (BrokerSettings, error) {
	if feePercent < 0 {
		feePercent = 0
	}
	if feePercent > 50 {
		feePercent = 50
	}
	if title == "" {
		title = "My exchange"
	}
	if len(title) > 100 {
		title = title[:100]
	}
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO exchange_settings (user_id, title, fee_percent, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id) DO UPDATE SET
		    title = EXCLUDED.title,
		    fee_percent = EXCLUDED.fee_percent,
		    updated_at = now()
	`, userID, title, feePercent)
	if err != nil {
		return BrokerSettings{}, fmt.Errorf("update broker settings: %w", err)
	}
	return BrokerSettings{UserID: userID, Title: title, FeePercent: feePercent}, nil
}

// BrokerStats возвращает статистику лотов брокера за период.
//
// SortField: date|lot|lot_price|lot_amount|lot_profit (legacy mapping).
// Page начинается с 1.
//
// План 72.1.45 §8: fee теперь per-broker — читаем GetBrokerSettings и
// используем `fee_percent` для Profit вместо хардкода 5%.
func (s *Service) BrokerStats(ctx context.Context, f BrokerStatsFilters) ([]BrokerStatsRow, BrokerStatsSummary, int, error) {
	if f.PerPage <= 0 {
		f.PerPage = 25
	}
	if f.Page <= 0 {
		f.Page = 1
	}

	bs, err := s.GetBrokerSettings(ctx, f.UserID)
	if err != nil {
		return nil, BrokerStatsSummary{}, 0, err
	}
	feePercent := bs.FeePercent

	// Whitelist sort_field → SQL column.
	sortCol := "sold_at"
	switch f.SortField {
	case "lot":
		sortCol = "artifact_unit_id"
	case "lot_price":
		sortCol = "price_oxsarit"
	case "lot_amount":
		sortCol = "quantity"
	case "lot_profit":
		sortCol = "price_oxsarit" // profit = price × fee, сортируем по price
	}
	order := "DESC"
	if f.SortOrder == "asc" {
		order = "ASC"
	}

	// Summary aggregates (без пагинации).
	var summary BrokerStatsSummary
	err = s.db.Pool().QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'sold'),
			COALESCE(SUM(price_oxsarit) FILTER (WHERE status = 'sold'), 0)
		FROM exchange_lots
		WHERE seller_user_id = $1
		  AND sold_at >= $2 AND sold_at <= $3
		  AND status IN ('sold','cancelled','expired')
	`, f.UserID, f.DateMin, f.DateMax).Scan(
		&summary.Total, &summary.Sold, &summary.Turnover,
	)
	if err != nil {
		return nil, BrokerStatsSummary{}, 0, fmt.Errorf("broker stats summary: %w", err)
	}
	summary.Profit = float64(summary.Turnover) * feePercent / 100.0

	// Pages.
	pages := summary.Total / f.PerPage
	if summary.Total%f.PerPage > 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}

	// Rows для текущей страницы.
	offset := (f.Page - 1) * f.PerPage
	query := fmt.Sprintf(`
		SELECT id, artifact_unit_id, quantity, price_oxsarit, status,
		       COALESCE(sold_at, created_at)
		FROM exchange_lots
		WHERE seller_user_id = $1
		  AND sold_at >= $2 AND sold_at <= $3
		  AND status IN ('sold','cancelled','expired')
		ORDER BY %s %s
		LIMIT $4 OFFSET $5
	`, sortCol, order)

	rows, err := s.db.Pool().Query(ctx, query, f.UserID, f.DateMin, f.DateMax, f.PerPage, offset)
	if err != nil {
		return nil, summary, pages, fmt.Errorf("broker stats rows: %w", err)
	}
	defer rows.Close()

	var out []BrokerStatsRow
	for rows.Next() {
		var r BrokerStatsRow
		if err := rows.Scan(&r.LotID, &r.UnitID, &r.Quantity, &r.Price, &r.Status, &r.SoldAt); err != nil {
			return nil, summary, pages, err
		}
		if r.Status == "sold" {
			r.Profit = float64(r.Price) * feePercent / 100.0
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, summary, pages, err
	}
	return out, summary, pages, nil
}
