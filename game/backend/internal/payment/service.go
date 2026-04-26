package payment

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/i18n"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// ReferralProcessor начисляет реферальные бонусы после покупки.
// Интерфейс разрывает циклическую зависимость payment → referral.
type ReferralProcessor interface {
	ProcessPurchase(ctx context.Context, buyerID string, amount float64) error
}

// AutoMsgSender — узкий интерфейс к automsg.SendDirect, чтобы не
// плодить зависимости. folder=8 (MSG_FOLDER_CREDIT в legacy).
type AutoMsgSender interface {
	SendDirect(ctx context.Context, tx pgx.Tx, userID string, folder int, title, body string) error
}

// Purchase — одна запись из истории покупок.
type Purchase struct {
	ID           string
	PackageKey   string
	PackageLabel string
	Credits      int
	PriceRub     float64
	Status       string
	CreatedAt    time.Time
	PaidAt       *time.Time
}

// Service управляет жизненным циклом платёжных заказов.
type Service struct {
	db       repo.Exec
	cfg      config.PaymentConfig
	gateway  Gateway
	referral ReferralProcessor
	automsg  AutoMsgSender
	bundle   *i18n.Bundle
}

// NewService создаёт Service. Если PAYMENT_PROVIDER не задан — gateway равен nil,
// вызовы CreateOrder вернут ErrGatewayDisabled.
func NewService(db repo.Exec, cfg config.PaymentConfig) *Service {
	svc := &Service{db: db, cfg: cfg}
	switch cfg.Provider {
	case "robokassa":
		svc.gateway = NewRobokassaGateway(cfg.RobokassaLogin, cfg.RobokassaPass1, cfg.RobokassaPass2)
	case "enot":
		svc.gateway = NewEnotGateway(cfg.EnotShopID, cfg.EnotApiKey)
	case "mock":
		svc.gateway = NewMockGateway(cfg.MockBaseURL)
	}
	return svc
}

// IsMock возвращает true, если активен mock-шлюз (для UI-баннера «тестовый режим»).
func (s *Service) IsMock() bool {
	if s.gateway == nil {
		return false
	}
	_, ok := s.gateway.(*MockGateway)
	return ok
}

// ConfirmPaymentDirect — прямой вызов подтверждения без прохождения через
// VerifyWebhook. Используется mock pay-эндпоинтом, где подпись не нужна.
func (s *Service) ConfirmPaymentDirect(ctx context.Context, orderID, providerID string) error {
	return s.ConfirmPayment(ctx, orderID, providerID)
}

// WithReferral подключает реферальный сервис (опционально).
func (s *Service) WithReferral(r ReferralProcessor) *Service {
	s.referral = r
	return s
}

// WithAutoMsg подключает сервис системных сообщений (опционально).
func (s *Service) WithAutoMsg(a AutoMsgSender) *Service {
	s.automsg = a
	return s
}

func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
}

func (s *Service) tr(group, key string, vars map[string]string) string {
	if s.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return s.bundle.Tr(i18n.LangRu, group, key, vars)
}

// CreateOrder создаёт pending-запись в credit_purchases и возвращает orderID и URL оплаты.
func (s *Service) CreateOrder(ctx context.Context, userID, packageKey string) (string, string, error) {
	if s.gateway == nil {
		return "", "", ErrGatewayDisabled
	}

	pkg, ok := PackageByKey(packageKey)
	if !ok {
		return "", "", ErrPackageNotFound
	}

	oid := ids.New()
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO credit_purchases (id, user_id, package_key, amount_credits, price_rub, provider, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
	`, oid, userID, pkg.Key, pkg.TotalCredits(), pkg.PriceRub(), s.cfg.Provider)
	if err != nil {
		return "", "", fmt.Errorf("payment: insert order: %w", err)
	}

	payURL, err := s.gateway.BuildPayURL(ctx, oid, pkg.Label, pkg.PriceKop, s.cfg.ReturnURL)
	if err != nil {
		return "", "", fmt.Errorf("payment: build pay url: %w", err)
	}

	return oid, payURL, nil
}

// ConfirmPayment зачисляет кредиты по orderID. Идемпотентен.
func (s *Service) ConfirmPayment(ctx context.Context, orderID, providerID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var userID string
		var credits int
		var priceRub float64
		var status string

		err := tx.QueryRow(ctx, `
			SELECT user_id, amount_credits, price_rub, status
			FROM credit_purchases WHERE id=$1 FOR UPDATE
		`, orderID).Scan(&userID, &credits, &priceRub, &status)
		if err == pgx.ErrNoRows {
			return ErrOrderNotFound
		}
		if err != nil {
			return fmt.Errorf("payment: fetch order: %w", err)
		}

		if status != "pending" {
			return nil // идемпотентность
		}

		if _, err = tx.Exec(ctx, `
			UPDATE credit_purchases
			SET status='paid', paid_at=now(), provider_id=$2
			WHERE id=$1
		`, orderID, providerID); err != nil {
			return fmt.Errorf("payment: update order: %w", err)
		}

		if _, err = tx.Exec(ctx, `
			UPDATE users SET credit = credit + $1 WHERE id=$2
		`, credits, userID); err != nil {
			return fmt.Errorf("payment: credit user: %w", err)
		}

		// Системное сообщение о зачислении (folder=8 CREDIT).
		if s.automsg != nil {
			title := s.tr("payment", "credited.title", nil)
			body := s.tr("payment", "credited.body", map[string]string{
				"credits": fmt.Sprintf("%d", credits),
				"orderId": orderID,
			})
			if err := s.automsg.SendDirect(ctx, tx, userID, 8, title, body); err != nil {
				slog.Warn("payment: credit msg failed", "order_id", orderID, "err", err.Error())
			}
		}

		// Реферальный бонус — после транзакции (ошибки не критичны).
		go func() {
			if s.referral != nil {
				if refErr := s.referral.ProcessPurchase(context.Background(), userID, priceRub); refErr != nil {
					slog.Error("payment: referral bonus failed", "order_id", orderID, "err", refErr)
				}
			}
		}()

		return nil
	})
}

// ListPurchases возвращает историю покупок игрока (paid + pending, новые первыми).
func (s *Service) ListPurchases(ctx context.Context, userID string) ([]Purchase, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, package_key, amount_credits, price_rub, status, created_at, paid_at
		FROM credit_purchases
		WHERE user_id=$1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("payment: list purchases: %w", err)
	}
	defer rows.Close()

	var result []Purchase
	for rows.Next() {
		var p Purchase
		if err = rows.Scan(&p.ID, &p.PackageKey, &p.Credits, &p.PriceRub, &p.Status, &p.CreatedAt, &p.PaidAt); err != nil {
			return nil, fmt.Errorf("payment: scan purchase: %w", err)
		}
		if pkg, ok := PackageByKey(p.PackageKey); ok {
			p.PackageLabel = pkg.Label
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
