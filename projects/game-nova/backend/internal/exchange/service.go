// Package exchange — биржа артефактов player-to-player (план 68).
//
// Биржа меняет N штук артефактов одного unit_id на оксариты (users.credit;
// см. ADR-0009). Эскроу: артефакты переходят в state='listed' при создании
// лота (атомарно с INSERT в exchange_lots / exchange_lot_items) и
// возвращаются в 'held' при cancel/expire или переходят к buyer'у при buy.
//
// События expire реализованы через event-loop: при создании лота вставляется
// KindExchangeExpire (план 65 event-loop) с fire_at = expires_at;
// handler возвращает escrow в случае истечения. При buy/cancel связанный
// event переводится в state='ok' (см. CancelExpireEvent).
package exchange

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/metrics"
)

// PermitChecker — проверка permit «Знак торговца».
//
// MVP-реализация AlwaysAllowPermit возвращает true для всех (gating
// отключён). В будущем (план премиум-вселенных) появится DBPermitChecker,
// проверяющий artefact_type='merchant_permit' у seller'а в active-state.
//
// DI через конструктор Service позволяет переключить реализацию без
// изменения логики биржи.
type PermitChecker interface {
	HasMerchantPermit(ctx context.Context, tx pgx.Tx, userID string) (bool, error)
}

// AlwaysAllowPermit — заглушка PermitChecker, возвращает true всегда.
// MVP-реализация (см. simplifications.md «План 68: permit-gating отключён»).
type AlwaysAllowPermit struct{}

func (AlwaysAllowPermit) HasMerchantPermit(_ context.Context, _ pgx.Tx, _ string) (bool, error) {
	return true, nil
}

// Config — параметры биржи (загружаются из configs/balance/default.yaml +
// override origin.yaml).
type Config struct {
	MaxQuantityPerLot     int           // 100
	MaxActiveLotsPerUser  int           // 10
	PriceCapMultiplier    float64       // 10.0 (1000% от reference)
	ReferenceWindow       time.Duration // 30 * 24h
	ExpiresInHoursMin     int           // 1
	ExpiresInHoursMax     int           // 168 (7 дней)
}

// DefaultConfig — fallback на случай если configs/balance не загрузился.
func DefaultConfig() Config {
	return Config{
		MaxQuantityPerLot:    100,
		MaxActiveLotsPerUser: 10,
		PriceCapMultiplier:   10.0,
		ReferenceWindow:      30 * 24 * time.Hour,
		ExpiresInHoursMin:    1,
		ExpiresInHoursMax:    168,
	}
}

// EventInserter — DI для event.Insert. Позволяет тестам подменять real
// pgx-вызов фейковой функцией без поднятия БД.
type EventInserter func(ctx context.Context, tx pgx.Tx, opts event.InsertOpts) (string, error)

// defaultEventInserter — обёртка над event.Insert. Используется в проде.
func defaultEventInserter(ctx context.Context, tx pgx.Tx, opts event.InsertOpts) (string, error) {
	return event.Insert(ctx, tx, opts)
}

type Service struct {
	db          repo.Exec
	repo        Repo
	cfg         Config
	permit      PermitChecker
	insertEvent EventInserter
}

func NewService(db repo.Exec, r Repo, cfg Config) *Service {
	return &Service{
		db:          db,
		repo:        r,
		cfg:         cfg,
		permit:      AlwaysAllowPermit{},
		insertEvent: defaultEventInserter,
	}
}

// WithPermitChecker — DI для тестов и будущей реальной реализации.
func (s *Service) WithPermitChecker(p PermitChecker) *Service {
	s.permit = p
	return s
}

// WithEventInserter — DI для тестов (mock event.Insert).
func (s *Service) WithEventInserter(ei EventInserter) *Service {
	s.insertEvent = ei
	return s
}

// CreateLotInput — входные параметры создания лота.
type CreateLotInput struct {
	SellerUserID   string
	ArtifactUnitID int
	Quantity       int
	PriceOxsarit   int64
	ExpiresInHours int
	IdempotencyKey string
}

// CreateLot — escrow + INSERT lot + INSERT items + INSERT history +
// INSERT event KindExchangeExpire.
//
// Атомарно: при любой ошибке транзакция откатывается и артефакты
// остаются в state='held'.
func (s *Service) CreateLot(ctx context.Context, in CreateLotInput) (Lot, error) {
	t0 := time.Now()
	defer func() {
		if metrics.ExchangeActionDuration != nil {
			metrics.ExchangeActionDuration.WithLabelValues("create").Observe(time.Since(t0).Seconds())
		}
	}()

	if in.Quantity <= 0 {
		return Lot{}, ErrInvalidQuantity
	}
	if in.Quantity > s.cfg.MaxQuantityPerLot {
		return Lot{}, ErrMaxQuantity
	}
	if in.PriceOxsarit <= 0 {
		return Lot{}, ErrInvalidPrice
	}
	if in.ExpiresInHours < s.cfg.ExpiresInHoursMin || in.ExpiresInHours > s.cfg.ExpiresInHoursMax {
		return Lot{}, ErrInvalidExpiry
	}

	var out Lot
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Permit-check.
		ok, err := s.permit.HasMerchantPermit(ctx, tx, in.SellerUserID)
		if err != nil {
			return fmt.Errorf("permit check: %w", err)
		}
		if !ok {
			return ErrPermitRequired
		}

		// 2. Лимит активных лотов.
		n, err := s.repo.CountActiveLotsBySeller(ctx, tx, in.SellerUserID)
		if err != nil {
			return err
		}
		if n >= s.cfg.MaxActiveLotsPerUser {
			return ErrMaxActiveLots
		}

		// 3. Price-cap (антифрод).
		ref, err := s.repo.AvgUnitPrice(ctx, tx, in.ArtifactUnitID, s.cfg.ReferenceWindow)
		if err != nil {
			return err
		}
		if ref != nil && *ref > 0 {
			cap := int64(float64(*ref) * s.cfg.PriceCapMultiplier)
			unitPrice := in.PriceOxsarit / int64(in.Quantity)
			if unitPrice > cap {
				return ErrPriceCapExceeded
			}
		}
		// Если ref == nil (нет истории за окно) — без cap'а (новый артефакт).

		// 4. Escrow: SELECT N штук FOR UPDATE.
		artefactIDs, err := s.repo.SelectAvailableArtefacts(ctx, tx,
			in.SellerUserID, in.ArtifactUnitID, in.Quantity)
		if err != nil {
			return err
		}
		if len(artefactIDs) < in.Quantity {
			return ErrInsufficientArtefacts
		}
		if err := s.repo.MarkArtefactsListed(ctx, tx, artefactIDs); err != nil {
			return err
		}

		// 5. INSERT lot.
		now := time.Now().UTC()
		expiresAt := now.Add(time.Duration(in.ExpiresInHours) * time.Hour)
		lot, err := s.repo.InsertLot(ctx, tx, Lot{
			SellerUserID:   in.SellerUserID,
			ArtifactUnitID: in.ArtifactUnitID,
			Quantity:       in.Quantity,
			PriceOxsarit:   in.PriceOxsarit,
			CreatedAt:      now,
			ExpiresAt:      expiresAt,
		})
		if err != nil {
			return err
		}

		// 6. INSERT items (artefact_id ↔ lot_id).
		if err := s.repo.InsertLotItems(ctx, tx, lot.ID, artefactIDs); err != nil {
			return err
		}

		// 7. INSERT event KindExchangeExpire с fire_at=expires_at.
		sellerCopy := in.SellerUserID
		eventID, err := s.insertEvent(ctx, tx, event.InsertOpts{
			UserID: &sellerCopy,
			Kind:   event.KindExchangeExpire,
			FireAt: expiresAt,
			Payload: map[string]any{
				"lot_id": lot.ID,
			},
		})
		if err != nil {
			return fmt.Errorf("insert expire event: %w", err)
		}
		if err := s.repo.SetLotExpireEvent(ctx, tx, lot.ID, eventID); err != nil {
			return err
		}
		lot.ExpireEventID = &eventID

		// 8. INSERT history.
		hp := HistoryPayloadCreated{
			ArtifactUnitID: in.ArtifactUnitID,
			Quantity:       in.Quantity,
			PriceOxsarit:   in.PriceOxsarit,
			ExpiresInHours: in.ExpiresInHours,
			IdempotencyKey: in.IdempotencyKey,
		}
		raw, _ := json.Marshal(hp)
		actor := in.SellerUserID
		if err := s.repo.InsertHistory(ctx, tx, lot.ID, "created", &actor, raw); err != nil {
			return err
		}

		out = lot
		return nil
	})
	if err != nil {
		recordExchangeAction("create", err)
		return Lot{}, err
	}
	recordExchangeAction("create", nil)
	slog.InfoContext(ctx, "exchange_lot_created",
		slog.String("lot_id", out.ID),
		slog.String("seller_user_id", out.SellerUserID),
		slog.Int("artifact_unit_id", out.ArtifactUnitID),
		slog.Int("quantity", out.Quantity),
		slog.Int64("price_oxsarit", out.PriceOxsarit))
	return out, nil
}

// BuyLot — атомарная покупка: блокировка лота, проверки, oxsarit-перевод,
// transfer артефактов на buyer'а, history, отмена expire-event.
func (s *Service) BuyLot(ctx context.Context, lotID, buyerID string) (Lot, error) {
	t0 := time.Now()
	defer func() {
		if metrics.ExchangeActionDuration != nil {
			metrics.ExchangeActionDuration.WithLabelValues("buy").Observe(time.Since(t0).Seconds())
		}
	}()

	var out Lot
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Lock lot.
		lot, err := s.repo.LockLotForUpdate(ctx, tx, lotID)
		if err != nil {
			return err
		}
		if lot.Status != "active" {
			return ErrLotNotActive
		}
		if !time.Now().Before(lot.ExpiresAt) {
			// Лот истёк, но expire-event ещё не сработал.
			return ErrLotNotActive
		}
		if lot.SellerUserID == buyerID {
			return ErrCannotBuyOwnLot
		}

		// 2. Buyer's home planet.
		homePlanet, err := s.repo.SelectHomePlanet(ctx, tx, buyerID)
		if err != nil {
			return err
		}

		// 3. Atomic oxsarit transfer: spend buyer (с проверкой баланса), add seller.
		if err := s.repo.SpendOxsarits(ctx, tx, buyerID, lot.PriceOxsarit); err != nil {
			return err
		}
		if err := s.repo.AddOxsarits(ctx, tx, lot.SellerUserID, lot.PriceOxsarit); err != nil {
			return err
		}

		// 4. Transfer artefacts (state='held', user_id=buyer, planet_id=home).
		items, err := s.repo.GetLotItems(ctx, lot.ID)
		if err != nil {
			return err
		}
		if err := s.repo.MarkArtefactsHeld(ctx, tx, items, buyerID, homePlanet); err != nil {
			return err
		}

		// 5. Lot → sold.
		soldAt := time.Now().UTC()
		if err := s.repo.MarkLotSold(ctx, tx, lot.ID, buyerID, soldAt); err != nil {
			return err
		}
		lot.Status = "sold"
		lot.BuyerUserID = &buyerID
		lot.SoldAt = &soldAt

		// 6. Cancel expire-event.
		if lot.ExpireEventID != nil {
			if err := s.repo.CancelExpireEvent(ctx, tx, *lot.ExpireEventID, "lot_bought"); err != nil {
				return err
			}
		}

		// 7. History.
		hp := HistoryPayloadBought{
			BuyerUserID:  buyerID,
			SellerUserID: lot.SellerUserID,
			Quantity:     lot.Quantity,
			PriceOxsarit: lot.PriceOxsarit,
		}
		raw, _ := json.Marshal(hp)
		actor := buyerID
		if err := s.repo.InsertHistory(ctx, tx, lot.ID, "bought", &actor, raw); err != nil {
			return err
		}

		out = lot
		return nil
	})
	if err != nil {
		recordExchangeAction("buy", err)
		return Lot{}, err
	}
	recordExchangeAction("buy", nil)
	if metrics.ExchangeOxsaritsVolume != nil {
		metrics.ExchangeOxsaritsVolume.Add(float64(out.PriceOxsarit))
	}
	slog.InfoContext(ctx, "exchange_lot_bought",
		slog.String("lot_id", out.ID),
		slog.String("seller_user_id", out.SellerUserID),
		slog.String("buyer_user_id", buyerID),
		slog.Int64("price_oxsarit", out.PriceOxsarit))
	return out, nil
}

// CancelLot — отзыв лота seller'ом. state='listed' → 'held' для всех
// items, lot.status='cancelled', expire-event отменяется.
func (s *Service) CancelLot(ctx context.Context, lotID, sellerID string) error {
	t0 := time.Now()
	defer func() {
		if metrics.ExchangeActionDuration != nil {
			metrics.ExchangeActionDuration.WithLabelValues("cancel").Observe(time.Since(t0).Seconds())
		}
	}()

	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		lot, err := s.repo.LockLotForUpdate(ctx, tx, lotID)
		if err != nil {
			return err
		}
		if lot.SellerUserID != sellerID {
			return ErrNotASeller
		}
		if lot.Status != "active" {
			return ErrLotNotActive
		}

		items, err := s.repo.GetLotItems(ctx, lot.ID)
		if err != nil {
			return err
		}
		if err := s.repo.MarkArtefactsHeld(ctx, tx, items, "", ""); err != nil {
			return err
		}
		if err := s.repo.MarkLotCancelled(ctx, tx, lot.ID); err != nil {
			return err
		}
		if lot.ExpireEventID != nil {
			if err := s.repo.CancelExpireEvent(ctx, tx, *lot.ExpireEventID, "seller_cancel"); err != nil {
				return err
			}
		}
		hp := HistoryPayloadCancelled{Reason: "seller_cancel"}
		raw, _ := json.Marshal(hp)
		actor := sellerID
		return s.repo.InsertHistory(ctx, tx, lot.ID, "cancelled", &actor, raw)
	})
	if err != nil {
		recordExchangeAction("cancel", err)
		return err
	}
	recordExchangeAction("cancel", nil)
	slog.InfoContext(ctx, "exchange_lot_cancelled",
		slog.String("lot_id", lotID),
		slog.String("seller_user_id", sellerID))
	return nil
}

// ListLots — обёртка с метриками.
func (s *Service) ListLots(ctx context.Context, f ListFilters) ([]Lot, string, error) {
	t0 := time.Now()
	defer func() {
		if metrics.ExchangeActionDuration != nil {
			metrics.ExchangeActionDuration.WithLabelValues("list").Observe(time.Since(t0).Seconds())
		}
	}()
	return s.repo.ListLots(ctx, f)
}

// GetLotWithItems — детали лота + список ID артефактов.
func (s *Service) GetLotWithItems(ctx context.Context, lotID string) (Lot, []string, error) {
	t0 := time.Now()
	defer func() {
		if metrics.ExchangeActionDuration != nil {
			metrics.ExchangeActionDuration.WithLabelValues("get").Observe(time.Since(t0).Seconds())
		}
	}()
	lot, err := s.repo.GetLot(ctx, lotID)
	if err != nil {
		return Lot{}, nil, err
	}
	items, err := s.repo.GetLotItems(ctx, lotID)
	if err != nil {
		return Lot{}, nil, err
	}
	return lot, items, nil
}

// Stats — для GET /api/exchange/stats.
func (s *Service) Stats(ctx context.Context) ([]StatsRow, error) {
	t0 := time.Now()
	defer func() {
		if metrics.ExchangeActionDuration != nil {
			metrics.ExchangeActionDuration.WithLabelValues("stats").Observe(time.Since(t0).Seconds())
		}
	}()
	return s.repo.Stats(ctx, s.cfg.ReferenceWindow)
}

// recordExchangeAction обновляет ExchangeLotsTotal по action+status.
func recordExchangeAction(action string, err error) {
	if metrics.ExchangeLotsTotal == nil {
		return
	}
	status := "ok"
	switch {
	case err == nil:
		status = "ok"
	case errors.Is(err, ErrInsufficientArtefacts), errors.Is(err, ErrInsufficientOxsarits),
		errors.Is(err, ErrMaxActiveLots), errors.Is(err, ErrMaxQuantity),
		errors.Is(err, ErrPriceCapExceeded), errors.Is(err, ErrPermitRequired):
		status = "insufficient"
	case errors.Is(err, ErrLotNotActive):
		status = "conflict"
	case errors.Is(err, ErrNotASeller), errors.Is(err, ErrCannotBuyOwnLot):
		status = "forbidden"
	case errors.Is(err, ErrLotNotFound):
		status = "not_found"
	default:
		status = "error"
	}
	metrics.ExchangeLotsTotal.WithLabelValues(action, status).Inc()
}
