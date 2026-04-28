package exchange

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Lot — доменная запись лота. Соответствует таблице exchange_lots.
//
// SellerUsername заполняется через JOIN users только в read-операциях
// (ListLots, GetLot); в write-сценариях остаётся пустым.
//
// JSON-теги — snake_case по контракту OpenAPI (R1, R12).
type Lot struct {
	ID               string     `json:"id"`
	SellerUserID     string     `json:"seller_user_id"`
	SellerUsername   string     `json:"seller_username,omitempty"`
	ArtifactUnitID   int        `json:"artifact_unit_id"`
	Quantity         int        `json:"quantity"`
	PriceOxsarit     int64      `json:"price_oxsarit"`
	UnitPriceOxsarit int64      `json:"unit_price_oxsarit"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	BuyerUserID      *string    `json:"buyer_user_id,omitempty"`
	SoldAt           *time.Time `json:"sold_at,omitempty"`
	ExpireEventID    *string    `json:"expire_event_id,omitempty"`
}

// ListFilters — фильтры для ListLots.
type ListFilters struct {
	ArtifactUnitID *int
	MinPrice       *int64
	MaxPrice       *int64
	SellerID       *string
	Status         *string // default — "active"
	Cursor         string  // opaque, формат: "<created_at_unix_nanos>:<id>"
	Limit          int     // 1..100, default 50
}

// StatsRow — агрегат по одному unit_id (для GET /api/exchange/stats).
type StatsRow struct {
	ArtifactUnitID int
	ActiveLots     int
	AvgUnitPrice   *int64 // nil если нет 'bought'-истории за окно
	Last30dVolume  int64
}

// Repo — узкий интерфейс БД-доступа биржи. Реализация PgRepo (repo_pgx.go)
// поверх pgx; mock-реализация в service_test.go для unit-тестов без БД.
//
// Все методы, которые пишут — принимают tx (часть транзакции, вызванной
// сервисом). Read-методы используют pool напрямую (consistent-read не
// нужен для list/get/stats).
type Repo interface {
	// ListLots — список лотов по фильтрам, с JOIN seller_username.
	// Возвращает (lots, nextCursor) — nextCursor пустой, если страница последняя.
	ListLots(ctx context.Context, f ListFilters) ([]Lot, string, error)

	// GetLot — детали одного лота с username.
	GetLot(ctx context.Context, id string) (Lot, error)

	// GetLotItems — id артефактов в лоте (для GET /lots/{id}).
	GetLotItems(ctx context.Context, lotID string) ([]string, error)

	// CountActiveLotsBySeller — для проверки лимита max_active_lots_per_user.
	CountActiveLotsBySeller(ctx context.Context, tx pgx.Tx, sellerID string) (int, error)

	// AvgUnitPrice — rolling-30d AVG(price/quantity) по 'bought' для антифрода.
	// Возвращает (nil, nil) если истории нет.
	AvgUnitPrice(ctx context.Context, tx pgx.Tx, artifactUnitID int, window time.Duration) (*int64, error)

	// SelectAvailableArtefacts — N штук артефактов sellerID/unit_id в state='held',
	// не находящихся в active-лотах биржи и не выставленных в artefact_offers.
	// FOR UPDATE — блокировка строк до конца транзакции.
	// Возвращает ровно min(N, available) ID артефактов.
	SelectAvailableArtefacts(ctx context.Context, tx pgx.Tx,
		sellerID string, artifactUnitID int, n int) ([]string, error)

	// MarkArtefactsListed — UPDATE artefacts_user.state='listed' для перечня ID.
	MarkArtefactsListed(ctx context.Context, tx pgx.Tx, artefactIDs []string) error

	// MarkArtefactsHeld — UPDATE state='held' (возврат при cancel/expire).
	// Опционально переписывает user_id и planet_id (для buy → buyer + buyer_planet).
	// Если newOwnerID == "" → owner не меняется (cancel/expire).
	MarkArtefactsHeld(ctx context.Context, tx pgx.Tx,
		artefactIDs []string, newOwnerID, newPlanetID string) error

	// InsertLot — вставка нового лота. Возвращает финальный Lot с заполненным id.
	InsertLot(ctx context.Context, tx pgx.Tx, l Lot) (Lot, error)

	// InsertLotItems — записи artefact_id ↔ lot_id для escrow.
	InsertLotItems(ctx context.Context, tx pgx.Tx, lotID string, artefactIDs []string) error

	// SetLotExpireEvent — связать лот с event_id для KindExchangeExpire
	// (после Insert event'а можем закольцевать ссылку).
	SetLotExpireEvent(ctx context.Context, tx pgx.Tx, lotID, eventID string) error

	// LockLotForUpdate — SELECT ... FOR UPDATE для buy/cancel пути.
	// Возвращает ErrLotNotFound, если строка отсутствует.
	LockLotForUpdate(ctx context.Context, tx pgx.Tx, id string) (Lot, error)

	// MarkLotSold / MarkLotCancelled / MarkLotExpired — обновление статуса.
	MarkLotSold(ctx context.Context, tx pgx.Tx, lotID, buyerID string, soldAt time.Time) error
	MarkLotCancelled(ctx context.Context, tx pgx.Tx, lotID string) error
	MarkLotExpired(ctx context.Context, tx pgx.Tx, lotID string) error

	// CancelExpireEvent — закрыть KindExchangeExpire-событие как 'ok' с
	// last_error='cancelled by buy/cancel'. Чтобы worker не подобрал.
	CancelExpireEvent(ctx context.Context, tx pgx.Tx, eventID, reason string) error

	// InsertHistory — запись в exchange_history. Универсальная для всех
	// kind'ов; payload — уже сериализованный JSONB.
	InsertHistory(ctx context.Context, tx pgx.Tx,
		lotID, eventKind string, actorUserID *string, payload []byte) error

	// SelectHomePlanet — самая старая планета пользователя (не луна,
	// not destroyed). Используется при buy для назначения planet_id новых
	// артефактов покупателя.
	SelectHomePlanet(ctx context.Context, tx pgx.Tx, userID string) (string, error)

	// SpendOxsarits — UPDATE users SET credit=credit-amount WHERE id=$1 AND credit>=$2.
	// Если 0 строк затронуто → ErrInsufficientOxsarits.
	SpendOxsarits(ctx context.Context, tx pgx.Tx, userID string, amount int64) error

	// AddOxsarits — UPDATE users SET credit=credit+amount.
	AddOxsarits(ctx context.Context, tx pgx.Tx, userID string, amount int64) error

	// SelectActiveLotsBySeller — для KindExchangeBan: все active-лоты seller'а
	// с lock (FOR UPDATE) — нужны для отзыва. Возвращает Lot с заполненным
	// expire_event_id (для CancelExpireEvent).
	SelectActiveLotsBySeller(ctx context.Context, tx pgx.Tx, sellerID string) ([]Lot, error)

	// Stats — агрегаты для GET /api/exchange/stats. Возвращает по unit_id:
	//   active_lots count, avg_unit_price (rolling 30d), last_30d_volume.
	Stats(ctx context.Context, window time.Duration) ([]StatsRow, error)
}
