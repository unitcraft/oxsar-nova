package referral

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
)

// Параметры реферальной системы (Dominator consts.php).
const (
	CreditPercent   = 0.20 // 20% от суммы покупки кредитов реферала
	BonusPoints     = 3000 // очки за каждого реферала
	MaxBonusPoints  = 500000
	StartingMetal   = int64(10)
	StartingSilicon = int64(5)
	StartingHydrogen = int64(2)
)

var ErrReferrerNotFound = errors.New("referral: referrer not found")

// AutoMsgSender — узкий интерфейс к automsg.SendDirect.
type AutoMsgSender interface {
	SendDirect(ctx context.Context, tx pgx.Tx, userID string, folder int, title, body string) error
}

type Service struct {
	db      repo.Exec
	automsg AutoMsgSender
}

func NewService(db repo.Exec) *Service {
	return &Service{db: db}
}

// WithAutoMsg подключает сервис системных сообщений (опционально).
func (s *Service) WithAutoMsg(a AutoMsgSender) *Service {
	s.automsg = a
	return s
}

// ProcessRegistration записывает referred_by и начисляет стартовые ресурсы
// новому игроку. Вызывается после создания пользователя.
// referrerID может быть пустым (тогда ничего не делается).
func (s *Service) ProcessRegistration(ctx context.Context, newUserID, referrerID string) error {
	if referrerID == "" {
		return nil
	}
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверить что реферер существует.
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, referrerID).Scan(&exists); err != nil {
			return fmt.Errorf("referral: check referrer: %w", err)
		}
		if !exists {
			return ErrReferrerNotFound
		}

		// Записать referred_by.
		if _, err := tx.Exec(ctx, `UPDATE users SET referred_by=$1 WHERE id=$2`, referrerID, newUserID); err != nil {
			return fmt.Errorf("referral: set referred_by: %w", err)
		}

		// Начислить стартовые ресурсы новому игроку на его домашнюю планету.
		if _, err := tx.Exec(ctx, `
			UPDATE planets
			SET metal    = metal    + $1,
			    silicon  = silicon  + $2,
			    hydrogen = hydrogen + $3
			WHERE user_id = $4
			  AND id = (SELECT id FROM planets WHERE user_id=$4 ORDER BY created_at LIMIT 1)
		`, StartingMetal, StartingSilicon, StartingHydrogen, newUserID); err != nil {
			return fmt.Errorf("referral: bonus resources: %w", err)
		}

		// Начислить реферальные очки рефереру (не более MaxBonusPoints суммарно).
		if _, err := tx.Exec(ctx, `
			UPDATE users
			SET e_points = LEAST(e_points + $1, $2)
			WHERE id = $3
		`, BonusPoints, MaxBonusPoints, referrerID); err != nil {
			return fmt.Errorf("referral: bonus points: %w", err)
		}

		return nil
	})
}

// ProcessPurchase начисляет рефереру CreditPercent от суммы покупки.
// Вызывается из payment webhook после успешной покупки кредитов.
// amount — число купленных кредитов.
func (s *Service) ProcessPurchase(ctx context.Context, buyerID string, amount float64) error {
	bonus := amount * CreditPercent
	if bonus <= 0 {
		return nil
	}
	// Получаем ID реферера и имя покупателя, чтобы уведомить реферера.
	var referrerID *string
	var buyerName string
	err := s.db.Pool().QueryRow(ctx, `
		SELECT referred_by, username FROM users WHERE id = $1
	`, buyerID).Scan(&referrerID, &buyerName)
	if err != nil {
		return fmt.Errorf("referral: lookup buyer: %w", err)
	}
	if referrerID == nil {
		return nil
	}

	_, err = s.db.Pool().Exec(ctx, `
		UPDATE users SET credit = credit + $1 WHERE id = $2
	`, bonus, *referrerID)
	if err != nil {
		return fmt.Errorf("referral: purchase bonus: %w", err)
	}

	// Системное сообщение рефереру (folder=8 CREDIT).
	if s.automsg != nil {
		title := "Реферальный бонус"
		body := fmt.Sprintf("Ваш реферал %s совершил покупку. На ваш счёт зачислено %.2f кредитов.", buyerName, bonus)
		if err := s.automsg.SendDirect(ctx, nil, *referrerID, 8, title, body); err != nil {
			// Не критично — покупка уже проведена.
			return nil
		}
	}
	return nil
}
