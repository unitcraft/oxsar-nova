// Package artmarket — продажа/покупка артефактов за credit.
//
// Credit — отдельная валюта, не связана с metal/silicon/hydrogen.
// Начальный баланс 0. Зарабатывается продажей артефактов и (в будущем)
// достижениями/платежами.
//
// Состояние артефакта при листинге: 'held' → 'listed'. При покупке:
// DELETE оффер + UPDATE artefacts_user (user_id=buyer, state='held').
// При снятии с продажи (cancel): 'listed' → 'held'.
package artmarket

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

var (
	ErrArtefactNotFound = errors.New("artmarket: artefact not found")
	ErrArtefactNotHeld  = errors.New("artmarket: artefact must be in 'held' state to sell")
	ErrNotOwner         = errors.New("artmarket: artefact not owned by user")
	ErrOfferNotFound    = errors.New("artmarket: offer not found")
	ErrOwnOffer         = errors.New("artmarket: cannot buy own offer")
	ErrNotEnoughCredit  = errors.New("artmarket: not enough credit")
	ErrInvalidPrice     = errors.New("artmarket: price must be > 0")
)

// Offer — листинг в market.
type Offer struct {
	ID           string    `json:"id"`
	ArtefactID   string    `json:"artefact_id"`
	SellerUserID string    `json:"seller_user_id"`
	SellerName   string    `json:"seller_name,omitempty"`
	UnitID       int       `json:"unit_id"`
	PriceCredit  int64     `json:"price_credit"`
	ListedAt     time.Time `json:"listed_at"`
}

// ListOffers возвращает все активные офферы (lowest price first).
// Фильтры типа «мои / не мои» — клиентская логика.
func (s *Service) ListOffers(ctx context.Context) ([]Offer, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT o.id, o.artefact_id, o.seller_user_id, COALESCE(u.username, ''),
		       o.unit_id, o.price_credit, o.listed_at
		FROM artefact_offers o
		LEFT JOIN users u ON u.id = o.seller_user_id
		ORDER BY o.price_credit ASC, o.listed_at ASC
		LIMIT 200
	`)
	if err != nil {
		return nil, fmt.Errorf("list offers: %w", err)
	}
	defer rows.Close()
	var out []Offer
	for rows.Next() {
		var o Offer
		if err := rows.Scan(&o.ID, &o.ArtefactID, &o.SellerUserID, &o.SellerName,
			&o.UnitID, &o.PriceCredit, &o.ListedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// Credit возвращает текущий баланс пользователя.
func (s *Service) Credit(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := s.db.Pool().QueryRow(ctx,
		`SELECT credit FROM users WHERE id = $1`, userID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("credit: %w", err)
	}
	return n, nil
}

// ListForSale выставляет артефакт на продажу. Требования:
//   - артефакт принадлежит userID;
//   - state='held' (нельзя продать активированный / в delayed / expired).
func (s *Service) ListForSale(ctx context.Context, userID, artefactID string, price int64) (Offer, error) {
	if price <= 0 {
		return Offer{}, ErrInvalidPrice
	}
	var out Offer
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			ownerID string
			state   string
			unitID  int
		)
		err := tx.QueryRow(ctx, `
			SELECT user_id, state, unit_id
			FROM artefacts_user WHERE id = $1 FOR UPDATE
		`, artefactID).Scan(&ownerID, &state, &unitID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrArtefactNotFound
			}
			return fmt.Errorf("read artefact: %w", err)
		}
		if ownerID != userID {
			return ErrNotOwner
		}
		if state != "held" {
			return ErrArtefactNotHeld
		}
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state='listed' WHERE id=$1`, artefactID); err != nil {
			return fmt.Errorf("update artefact state: %w", err)
		}
		offerID := ids.New()
		now := time.Now().UTC()
		if _, err := tx.Exec(ctx, `
			INSERT INTO artefact_offers
				(id, artefact_id, seller_user_id, unit_id, price_credit, listed_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, offerID, artefactID, userID, unitID, price, now); err != nil {
			return fmt.Errorf("insert offer: %w", err)
		}
		out = Offer{
			ID: offerID, ArtefactID: artefactID, SellerUserID: userID,
			UnitID: unitID, PriceCredit: price, ListedAt: now,
		}
		return nil
	})
	return out, err
}

// Cancel — снятие собственного оффера. state 'listed' → 'held'.
func (s *Service) Cancel(ctx context.Context, userID, offerID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			sellerID   string
			artefactID string
		)
		err := tx.QueryRow(ctx,
			`SELECT seller_user_id, artefact_id FROM artefact_offers WHERE id=$1 FOR UPDATE`,
			offerID).Scan(&sellerID, &artefactID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrOfferNotFound
			}
			return fmt.Errorf("read offer: %w", err)
		}
		if sellerID != userID {
			return ErrNotOwner
		}
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state='held' WHERE id=$1`, artefactID); err != nil {
			return fmt.Errorf("revert state: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`DELETE FROM artefact_offers WHERE id=$1`, offerID); err != nil {
			return fmt.Errorf("delete offer: %w", err)
		}
		return nil
	})
}

// Buy — покупка артефакта. Переводит credit, меняет владельца,
// удаляет оффер.
func (s *Service) Buy(ctx context.Context, buyerID, offerID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			sellerID    string
			artefactID  string
			price       int64
		)
		err := tx.QueryRow(ctx, `
			SELECT seller_user_id, artefact_id, price_credit
			FROM artefact_offers WHERE id=$1 FOR UPDATE
		`, offerID).Scan(&sellerID, &artefactID, &price)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrOfferNotFound
			}
			return fmt.Errorf("read offer: %w", err)
		}
		if sellerID == buyerID {
			return ErrOwnOffer
		}
		// Проверка баланса покупателя.
		var buyerCredit int64
		if err := tx.QueryRow(ctx,
			`SELECT credit FROM users WHERE id=$1 FOR UPDATE`, buyerID).Scan(&buyerCredit); err != nil {
			return fmt.Errorf("buyer credit: %w", err)
		}
		if buyerCredit < price {
			return ErrNotEnoughCredit
		}
		// Перевод credit.
		if _, err := tx.Exec(ctx,
			`UPDATE users SET credit = credit - $1 WHERE id=$2`, price, buyerID); err != nil {
			return fmt.Errorf("debit buyer: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET credit = credit + $1 WHERE id=$2`, price, sellerID); err != nil {
			return fmt.Errorf("credit seller: %w", err)
		}
		// Смена владельца + state=held.
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET user_id=$1, state='held' WHERE id=$2`,
			buyerID, artefactID); err != nil {
			return fmt.Errorf("transfer artefact: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`DELETE FROM artefact_offers WHERE id=$1`, offerID); err != nil {
			return fmt.Errorf("delete offer: %w", err)
		}
		return nil
	})
}
