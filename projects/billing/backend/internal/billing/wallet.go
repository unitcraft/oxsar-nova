// Package billing — кошельки, транзакции, оркестрация billing-операций.
//
// Архитектура: docs/plans/38-billing-service.md.
//
// Принципы:
//   - SELECT FOR UPDATE на wallets.balance — защита от race condition.
//   - transactions immutable (INSERT only). balance — материализованная
//     производная (с reconcile cron-job для сверки).
//   - Все суммы — BIGINT в минимальных единицах (копейки/satoshi).
package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/billing/internal/repo"
)

const (
	DefaultCurrency = "OXC" // Oxsar Credits — внутренняя валюта.
)

var (
	// ErrInsufficient — недостаточно средств. HTTP 402 Payment Required.
	ErrInsufficient = errors.New("billing: insufficient funds")
	// ErrFrozen — кошелёк заморожен (reconcile-расхождение). HTTP 423 Locked.
	ErrFrozen = errors.New("billing: wallet frozen")
	// ErrInvalidAmount — некорректная сумма. HTTP 400.
	ErrInvalidAmount = errors.New("billing: amount must be positive")
)

// Service — основной сервис billing. Оркестрирует wallet/orders/payments.
type Service struct {
	db *repo.PG
}

// New создаёт Service.
func New(db *repo.PG) *Service {
	return &Service{db: db}
}

// Wallet — публичная проекция кошелька.
type Wallet struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	CurrencyCode string    `json:"currency_code"`
	Balance      int64     `json:"balance"` // в минимальных единицах
	Frozen       bool      `json:"frozen"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Transaction — публичная проекция транзакции.
type Transaction struct {
	ID             string    `json:"id"`
	WalletID       string    `json:"wallet_id"`
	Delta          int64     `json:"delta"`
	BalanceAfter   int64     `json:"balance_after"`
	FromAccount    string    `json:"from_account"`
	ToAccount      string    `json:"to_account"`
	Reason         string    `json:"reason"`
	RefID          string    `json:"ref_id,omitempty"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// SpendInput — параметры списания.
type SpendInput struct {
	UserID         string
	Currency       string // если пусто — DefaultCurrency
	Amount         int64  // > 0
	Reason         string // 'feedback_vote' | 'shop_purchase' | ...
	RefID          string // идентификатор причины (feedback_id и т.п.)
	ToAccount      string // куда логически уходит ('vote:feedback:<id>')
	IdempotencyKey string // если пусто — без idempotency
}

// CreditInput — параметры пополнения (внутреннее, например webhook).
type CreditInput struct {
	UserID         string
	Currency       string
	Amount         int64
	Reason         string // 'top_up' | 'admin_grant' | 'refund' | ...
	RefID          string // order_id и т.п.
	FromAccount    string // откуда логически приходит ('payment:robokassa:<order>')
	IdempotencyKey string
}

// Spend атомарно списывает средства с кошелька.
//
// Алгоритм:
//   1. BEGIN.
//   2. SELECT wallet FOR UPDATE (создаём, если нет).
//   3. Проверка balance >= amount, frozen=false.
//   4. INSERT transaction (delta=−amount, balance_after).
//   5. UPDATE wallet.balance.
//   6. COMMIT.
//
// При параллельных запросах второй блокируется на FOR UPDATE до конца первого.
func (s *Service) Spend(ctx context.Context, in SpendInput) (Transaction, error) {
	if in.Amount <= 0 {
		return Transaction{}, ErrInvalidAmount
	}
	currency := in.Currency
	if currency == "" {
		currency = DefaultCurrency
	}
	var tx Transaction
	err := s.db.InTx(ctx, func(ctx context.Context, dbtx pgx.Tx) error {
		w, err := lockOrCreateWallet(ctx, dbtx, in.UserID, currency)
		if err != nil {
			return err
		}
		if w.frozen {
			return ErrFrozen
		}
		if w.balance < in.Amount {
			return ErrInsufficient
		}
		newBalance := w.balance - in.Amount
		fromAcc := fmt.Sprintf("wallet:user_%s:%s", in.UserID, currency)
		row := dbtx.QueryRow(ctx, `
			INSERT INTO transactions
				(wallet_id, delta, balance_after, from_account, to_account, reason, ref_id, idempotency_key)
			VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''))
			RETURNING id, created_at
		`, w.id, -in.Amount, newBalance, fromAcc, in.ToAccount, in.Reason, in.RefID, in.IdempotencyKey)
		var txID string
		var createdAt time.Time
		if err := row.Scan(&txID, &createdAt); err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}
		if _, err := dbtx.Exec(ctx, `
			UPDATE wallets SET balance = $1, updated_at = now() WHERE id = $2
		`, newBalance, w.id); err != nil {
			return fmt.Errorf("update wallet: %w", err)
		}
		tx = Transaction{
			ID:             txID,
			WalletID:       w.id,
			Delta:          -in.Amount,
			BalanceAfter:   newBalance,
			FromAccount:    fromAcc,
			ToAccount:      in.ToAccount,
			Reason:         in.Reason,
			RefID:          in.RefID,
			IdempotencyKey: in.IdempotencyKey,
			CreatedAt:      createdAt,
		}
		return nil
	})
	return tx, err
}

// Credit атомарно пополняет кошелёк.
func (s *Service) Credit(ctx context.Context, in CreditInput) (Transaction, error) {
	if in.Amount <= 0 {
		return Transaction{}, ErrInvalidAmount
	}
	currency := in.Currency
	if currency == "" {
		currency = DefaultCurrency
	}
	var tx Transaction
	err := s.db.InTx(ctx, func(ctx context.Context, dbtx pgx.Tx) error {
		w, err := lockOrCreateWallet(ctx, dbtx, in.UserID, currency)
		if err != nil {
			return err
		}
		if w.frozen {
			return ErrFrozen
		}
		newBalance := w.balance + in.Amount
		toAcc := fmt.Sprintf("wallet:user_%s:%s", in.UserID, currency)
		row := dbtx.QueryRow(ctx, `
			INSERT INTO transactions
				(wallet_id, delta, balance_after, from_account, to_account, reason, ref_id, idempotency_key)
			VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''))
			RETURNING id, created_at
		`, w.id, in.Amount, newBalance, in.FromAccount, toAcc, in.Reason, in.RefID, in.IdempotencyKey)
		var txID string
		var createdAt time.Time
		if err := row.Scan(&txID, &createdAt); err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}
		if _, err := dbtx.Exec(ctx, `
			UPDATE wallets SET balance = $1, updated_at = now() WHERE id = $2
		`, newBalance, w.id); err != nil {
			return fmt.Errorf("update wallet: %w", err)
		}
		tx = Transaction{
			ID:             txID,
			WalletID:       w.id,
			Delta:          in.Amount,
			BalanceAfter:   newBalance,
			FromAccount:    in.FromAccount,
			ToAccount:      toAcc,
			Reason:         in.Reason,
			RefID:          in.RefID,
			IdempotencyKey: in.IdempotencyKey,
			CreatedAt:      createdAt,
		}
		return nil
	})
	return tx, err
}

// Balance возвращает текущий баланс. Создаёт кошелёк (с balance=0), если его нет
// — это корректно: первый Balance-запрос для нового юзера должен давать 0,
// а не 404.
func (s *Service) Balance(ctx context.Context, userID, currency string) (Wallet, error) {
	if currency == "" {
		currency = DefaultCurrency
	}
	var w Wallet
	err := s.db.InTx(ctx, func(ctx context.Context, dbtx pgx.Tx) error {
		ws, err := lockOrCreateWallet(ctx, dbtx, userID, currency)
		if err != nil {
			return err
		}
		w = Wallet{
			ID:           ws.id,
			UserID:       userID,
			CurrencyCode: currency,
			Balance:      ws.balance,
			Frozen:       ws.frozen,
			UpdatedAt:    ws.updatedAt,
		}
		return nil
	})
	return w, err
}

// History возвращает последние limit транзакций кошелька.
func (s *Service) History(ctx context.Context, userID, currency string, limit, offset int) ([]Transaction, error) {
	if currency == "" {
		currency = DefaultCurrency
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT t.id, t.wallet_id, t.delta, t.balance_after,
		       t.from_account, t.to_account, t.reason,
		       COALESCE(t.ref_id, ''), COALESCE(t.idempotency_key, ''),
		       t.created_at
		FROM transactions t
		JOIN wallets w ON w.id = t.wallet_id
		WHERE w.user_id = $1 AND w.currency_code = $2
		ORDER BY t.created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, currency, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()
	var out []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Delta, &t.BalanceAfter,
			&t.FromAccount, &t.ToAccount, &t.Reason, &t.RefID, &t.IdempotencyKey,
			&t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// walletState — внутренний снапшот строки wallets под FOR UPDATE.
type walletState struct {
	id        string
	balance   int64
	frozen    bool
	updatedAt time.Time
}

// lockOrCreateWallet атомарно блокирует строку wallet (или создаёт, если нет).
//
// Реализация: INSERT ... ON CONFLICT DO UPDATE (UPSERT) — сразу получаем
// строку и блокировку. ON CONFLICT нужен для обработки race condition между
// SELECT и INSERT (когда два запроса одновременно создают кошелёк).
//
// После UPSERT делаем SELECT FOR UPDATE — на случай, если INSERT просто
// прочитал существующую строку (без RETURNING её не блокирует).
func lockOrCreateWallet(ctx context.Context, dbtx pgx.Tx, userID, currency string) (*walletState, error) {
	// Сначала пытаемся создать (idempotent INSERT).
	_, err := dbtx.Exec(ctx, `
		INSERT INTO wallets (user_id, currency_code, balance)
		VALUES ($1, $2, 0)
		ON CONFLICT (user_id, currency_code) DO NOTHING
	`, userID, currency)
	if err != nil {
		return nil, fmt.Errorf("upsert wallet: %w", err)
	}
	// Теперь FOR UPDATE — блокируем строку.
	var s walletState
	err = dbtx.QueryRow(ctx, `
		SELECT id, balance, frozen, updated_at
		FROM wallets
		WHERE user_id = $1 AND currency_code = $2
		FOR UPDATE
	`, userID, currency).Scan(&s.id, &s.balance, &s.frozen, &s.updatedAt)
	if err != nil {
		return nil, fmt.Errorf("lock wallet: %w", err)
	}
	return &s, nil
}
