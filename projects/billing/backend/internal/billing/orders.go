package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/billing/internal/payment"
)

// Errors.
var (
	ErrOrderNotFound = errors.New("billing: order not found")
	ErrOrderExpired  = errors.New("billing: order expired")
	ErrOrderClosed   = errors.New("billing: order already closed")
)

// Order — публичная проекция payment_order.
type Order struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	PackageID  string    `json:"package_id"`
	AmountKop  int64     `json:"amount_kop"`
	Credits    int64     `json:"credits"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// CreateOrder создаёт payment_order и возвращает PayURL для редиректа клиента.
//
// Алгоритм:
//   1. Найти Package по id.
//   2. INSERT payment_orders (status='pending', expires_at=now+1h).
//   3. Gateway.BuildPayURL(...) → ссылка для редиректа.
func (s *Service) CreateOrder(ctx context.Context, userID, packageID, returnURL string, gw payment.Gateway) (Order, string, error) {
	pkg, err := payment.FindPackage(packageID)
	if err != nil {
		return Order{}, "", err
	}
	totalCredits := pkg.TotalCredits()
	var orderID string
	var createdAt, expiresAt time.Time
	row := s.db.Pool().QueryRow(ctx, `
		INSERT INTO payment_orders
			(user_id, provider, package_id, amount_kop, credits, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, created_at, expires_at
	`, userID, gw.Name(), pkg.ID, pkg.AmountKop, totalCredits)
	if err := row.Scan(&orderID, &createdAt, &expiresAt); err != nil {
		return Order{}, "", fmt.Errorf("insert order: %w", err)
	}
	payURL, err := gw.BuildPayURL(ctx, orderID, userID, pkg.AmountKop, returnURL)
	if err != nil {
		return Order{}, "", fmt.Errorf("build payurl: %w", err)
	}
	return Order{
		ID:        orderID,
		UserID:    userID,
		Provider:  gw.Name(),
		PackageID: pkg.ID,
		AmountKop: pkg.AmountKop,
		Credits:   totalCredits,
		Status:    "pending",
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}, payURL, nil
}

// PayOrder помечает order как paid и пополняет кошелёк юзера.
// Вызывается из webhook-handler-а после verify подписи.
//
// Идемпотентен по order_id: если order уже paid — ничего не делает (no-op).
// Это критично для Robokassa, которая может слать webhook несколько раз.
func (s *Service) PayOrder(ctx context.Context, orderID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, dbtx pgx.Tx) error {
		var userID, packageID, status string
		var amountKop, credits int64
		var expiresAt time.Time
		err := dbtx.QueryRow(ctx, `
			SELECT user_id, package_id, amount_kop, credits, status, expires_at
			FROM payment_orders
			WHERE id = $1
			FOR UPDATE
		`, orderID).Scan(&userID, &packageID, &amountKop, &credits, &status, &expiresAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrOrderNotFound
			}
			return fmt.Errorf("lock order: %w", err)
		}
		if status == "paid" {
			// Idempotent: повторный webhook ничего не делает.
			return nil
		}
		if status != "pending" {
			return ErrOrderClosed
		}
		if time.Now().After(expiresAt) {
			// Order протух — помечаем как expired, не зачисляем.
			_, _ = dbtx.Exec(ctx, `UPDATE payment_orders SET status = 'expired' WHERE id = $1`, orderID)
			return ErrOrderExpired
		}
		// Пометить как paid.
		_, err = dbtx.Exec(ctx, `
			UPDATE payment_orders
			SET status = 'paid', paid_at = now()
			WHERE id = $1
		`, orderID)
		if err != nil {
			return fmt.Errorf("update order: %w", err)
		}
		// Пополнить кошелёк. Используем тот же tx (атомарно с UPDATE order):
		// либо обе записи появятся, либо ни одной.
		w, err := lockOrCreateWallet(ctx, dbtx, userID, DefaultCurrency)
		if err != nil {
			return err
		}
		newBalance := w.balance + credits
		fromAcc := fmt.Sprintf("payment:order_%s", orderID)
		toAcc := fmt.Sprintf("wallet:user_%s:%s", userID, DefaultCurrency)
		_, err = dbtx.Exec(ctx, `
			INSERT INTO transactions
				(wallet_id, delta, balance_after, from_account, to_account, reason, ref_id)
			VALUES ($1, $2, $3, $4, $5, 'top_up', $6)
		`, w.id, credits, newBalance, fromAcc, toAcc, orderID)
		if err != nil {
			return fmt.Errorf("insert tx: %w", err)
		}
		_, err = dbtx.Exec(ctx, `
			UPDATE wallets SET balance = $1, updated_at = now() WHERE id = $2
		`, newBalance, w.id)
		if err != nil {
			return fmt.Errorf("update wallet: %w", err)
		}
		return nil
	})
}

// GetOrder читает order по id (для admin / debug). Возвращает ErrOrderNotFound.
func (s *Service) GetOrder(ctx context.Context, orderID string) (Order, error) {
	var o Order
	var paidAt *time.Time
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, provider, package_id, amount_kop, credits,
		       status, created_at, paid_at, expires_at
		FROM payment_orders WHERE id = $1
	`, orderID).Scan(&o.ID, &o.UserID, &o.Provider, &o.PackageID, &o.AmountKop, &o.Credits,
		&o.Status, &o.CreatedAt, &paidAt, &o.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, fmt.Errorf("get order: %w", err)
	}
	o.PaidAt = paidAt
	return o, nil
}
