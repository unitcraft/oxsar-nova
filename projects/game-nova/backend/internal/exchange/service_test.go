package exchange

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// fakeRepo — in-memory mock-реализация Repo для unit-тестов сервиса.
//
// Сохраняет минимальное достаточное состояние для проверки инвариантов:
//   - артефакты с unit_id и state ('held' / 'listed');
//   - оксариты на пользователях;
//   - лоты + items + history;
//   - home-планеты пользователей.
//
// Реализует Repo полностью (compile-time check внизу). Транзакция — no-op
// (всё in-memory), tx-параметр игнорируется.
type fakeRepo struct {
	// artefacts: id → {ownerID, unitID, state}
	artefacts map[string]*fakeArtefact
	// oxsarits: userID → balance
	oxsarits map[string]int64
	// lots: id → Lot
	lots map[string]*Lot
	// lotItems: lotID → []artefactID
	lotItems map[string][]string
	// history: lotID → []record (для тестов считаем bought-историю)
	history []historyRec
	// homes: userID → planetID
	homes map[string]string
	// avgOverride: при != nil — игнорирует историю и возвращает это значение.
	avgOverride *int64
}

type fakeArtefact struct {
	owner  string
	unitID int
	state  string
}

type historyRec struct {
	lotID     string
	eventKind string
	createdAt time.Time
	priceUnit int64 // для AVG-расчёта при event_kind='bought'
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		artefacts: map[string]*fakeArtefact{},
		oxsarits:  map[string]int64{},
		lots:      map[string]*Lot{},
		lotItems:  map[string][]string{},
		history:   nil,
		homes:     map[string]string{},
	}
}

func (f *fakeRepo) addUser(id string, oxsarits int64, homePlanet string) {
	f.oxsarits[id] = oxsarits
	f.homes[id] = homePlanet
}

func (f *fakeRepo) addArtefact(ownerID string, unitID int) string {
	id := ids.New()
	f.artefacts[id] = &fakeArtefact{owner: ownerID, unitID: unitID, state: "held"}
	return id
}

// totalQtyOfUnit возвращает сумму держателей+listed для unit_id.
// Используется в escrow-инварианте: эта величина не должна меняться при
// create/buy/cancel/expire.
func (f *fakeRepo) totalQtyOfUnit(unitID int) int {
	n := 0
	for _, a := range f.artefacts {
		if a.unitID == unitID {
			n++
		}
	}
	return n
}

func (f *fakeRepo) totalOxsarits() int64 {
	var s int64
	for _, v := range f.oxsarits {
		s += v
	}
	return s
}

// ---- Repo interface ---------------------------------------------------

func (f *fakeRepo) ListLots(ctx context.Context, fl ListFilters) ([]Lot, string, error) {
	out := make([]Lot, 0)
	status := "active"
	if fl.Status != nil && *fl.Status != "" {
		status = *fl.Status
	}
	for _, l := range f.lots {
		if l.Status != status {
			continue
		}
		if fl.ArtifactUnitID != nil && l.ArtifactUnitID != *fl.ArtifactUnitID {
			continue
		}
		if fl.SellerID != nil && l.SellerUserID != *fl.SellerID {
			continue
		}
		if fl.MinPrice != nil && l.PriceOxsarit < *fl.MinPrice {
			continue
		}
		if fl.MaxPrice != nil && l.PriceOxsarit > *fl.MaxPrice {
			continue
		}
		out = append(out, *l)
	}
	return out, "", nil
}

func (f *fakeRepo) GetLot(ctx context.Context, id string) (Lot, error) {
	l, ok := f.lots[id]
	if !ok {
		return Lot{}, ErrLotNotFound
	}
	return *l, nil
}

func (f *fakeRepo) GetLotItems(ctx context.Context, lotID string) ([]string, error) {
	items, ok := f.lotItems[lotID]
	if !ok {
		return nil, nil
	}
	cp := make([]string, len(items))
	copy(cp, items)
	return cp, nil
}

func (f *fakeRepo) CountActiveLotsBySeller(ctx context.Context, _ pgx.Tx, sellerID string) (int, error) {
	n := 0
	for _, l := range f.lots {
		if l.SellerUserID == sellerID && l.Status == "active" {
			n++
		}
	}
	return n, nil
}

func (f *fakeRepo) AvgUnitPrice(ctx context.Context, _ pgx.Tx, unitID int, window time.Duration) (*int64, error) {
	if f.avgOverride != nil {
		v := *f.avgOverride
		return &v, nil
	}
	cutoff := time.Now().Add(-window)
	var sum int64
	var n int
	for _, h := range f.history {
		if h.eventKind != "bought" {
			continue
		}
		if h.createdAt.Before(cutoff) {
			continue
		}
		l, ok := f.lots[h.lotID]
		if !ok || l.ArtifactUnitID != unitID {
			continue
		}
		sum += h.priceUnit
		n++
	}
	if n == 0 {
		return nil, nil
	}
	avg := sum / int64(n)
	return &avg, nil
}

func (f *fakeRepo) SelectAvailableArtefacts(ctx context.Context, _ pgx.Tx,
	sellerID string, unitID int, n int) ([]string, error) {
	out := make([]string, 0, n)
	for id, a := range f.artefacts {
		if a.owner == sellerID && a.unitID == unitID && a.state == "held" {
			out = append(out, id)
			if len(out) >= n {
				break
			}
		}
	}
	return out, nil
}

func (f *fakeRepo) MarkArtefactsListed(ctx context.Context, _ pgx.Tx, ids []string) error {
	for _, id := range ids {
		if a, ok := f.artefacts[id]; ok {
			a.state = "listed"
		}
	}
	return nil
}

func (f *fakeRepo) MarkArtefactsHeld(ctx context.Context, _ pgx.Tx,
	ids []string, newOwner, newPlanet string) error {
	for _, id := range ids {
		if a, ok := f.artefacts[id]; ok {
			a.state = "held"
			if newOwner != "" {
				a.owner = newOwner
			}
		}
	}
	return nil
}

func (f *fakeRepo) InsertLot(ctx context.Context, _ pgx.Tx, l Lot) (Lot, error) {
	if l.ID == "" {
		l.ID = ids.New()
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	l.Status = "active"
	l.UnitPriceOxsarit = l.PriceOxsarit / int64(l.Quantity)
	cp := l
	f.lots[l.ID] = &cp
	return l, nil
}

func (f *fakeRepo) InsertLotItems(ctx context.Context, _ pgx.Tx, lotID string, items []string) error {
	cp := make([]string, len(items))
	copy(cp, items)
	f.lotItems[lotID] = cp
	return nil
}

func (f *fakeRepo) SetLotExpireEvent(ctx context.Context, _ pgx.Tx, lotID, eventID string) error {
	if l, ok := f.lots[lotID]; ok {
		l.ExpireEventID = &eventID
	}
	return nil
}

func (f *fakeRepo) LockLotForUpdate(ctx context.Context, _ pgx.Tx, id string) (Lot, error) {
	return f.GetLot(ctx, id)
}

func (f *fakeRepo) MarkLotSold(ctx context.Context, _ pgx.Tx, lotID, buyerID string, soldAt time.Time) error {
	if l, ok := f.lots[lotID]; ok && l.Status == "active" {
		l.Status = "sold"
		l.BuyerUserID = &buyerID
		l.SoldAt = &soldAt
	}
	return nil
}

func (f *fakeRepo) MarkLotCancelled(ctx context.Context, _ pgx.Tx, lotID string) error {
	if l, ok := f.lots[lotID]; ok && l.Status == "active" {
		l.Status = "cancelled"
	}
	return nil
}

func (f *fakeRepo) MarkLotExpired(ctx context.Context, _ pgx.Tx, lotID string) error {
	if l, ok := f.lots[lotID]; ok && l.Status == "active" {
		l.Status = "expired"
	}
	return nil
}

func (f *fakeRepo) CancelExpireEvent(ctx context.Context, _ pgx.Tx, eventID, reason string) error {
	return nil
}

func (f *fakeRepo) InsertHistory(ctx context.Context, _ pgx.Tx,
	lotID, kind string, actor *string, payload []byte) error {
	rec := historyRec{lotID: lotID, eventKind: kind, createdAt: time.Now()}
	if l, ok := f.lots[lotID]; ok && kind == "bought" {
		rec.priceUnit = l.PriceOxsarit / int64(l.Quantity)
	}
	f.history = append(f.history, rec)
	return nil
}

func (f *fakeRepo) SelectHomePlanet(ctx context.Context, _ pgx.Tx, userID string) (string, error) {
	p, ok := f.homes[userID]
	if !ok || p == "" {
		return "", ErrUserHasNoPlanet
	}
	return p, nil
}

func (f *fakeRepo) SpendOxsarits(ctx context.Context, _ pgx.Tx, userID string, amount int64) error {
	bal := f.oxsarits[userID]
	if bal < amount {
		return ErrInsufficientOxsarits
	}
	f.oxsarits[userID] = bal - amount
	return nil
}

func (f *fakeRepo) AddOxsarits(ctx context.Context, _ pgx.Tx, userID string, amount int64) error {
	f.oxsarits[userID] += amount
	return nil
}

func (f *fakeRepo) SelectActiveLotsBySeller(ctx context.Context, _ pgx.Tx, sellerID string) ([]Lot, error) {
	var out []Lot
	for _, l := range f.lots {
		if l.SellerUserID == sellerID && l.Status == "active" {
			out = append(out, *l)
		}
	}
	return out, nil
}

func (f *fakeRepo) Stats(ctx context.Context, window time.Duration) ([]StatsRow, error) {
	return nil, nil
}

// План 72.1.27: Premium + Ban (stubs для тестов).

func (f *fakeRepo) CountActiveFeaturedLots(_ context.Context, _ pgx.Tx, _ time.Duration) (int, error) {
	n := 0
	for _, l := range f.lots {
		if l.Status == "active" && l.BannedAt == nil && l.FeaturedAt != nil {
			n++
		}
	}
	return n, nil
}

func (f *fakeRepo) MarkLotFeatured(_ context.Context, _ pgx.Tx, lotID string, at time.Time) error {
	for i := range f.lots {
		if f.lots[i].ID == lotID && f.lots[i].Status == "active" && f.lots[i].BannedAt == nil {
			t := at
			f.lots[i].FeaturedAt = &t
			return nil
		}
	}
	return ErrLotNotActive
}

func (f *fakeRepo) MarkLotBanned(_ context.Context, _ pgx.Tx, lotID string, at time.Time) error {
	for i := range f.lots {
		if f.lots[i].ID == lotID && f.lots[i].Status == "active" {
			f.lots[i].Status = "banned"
			t := at
			f.lots[i].BannedAt = &t
			return nil
		}
	}
	return ErrLotNotActive
}

func (f *fakeRepo) CheckIsAdmin(_ context.Context, _ pgx.Tx, userID string) (bool, error) {
	// fakeRepo: admin если userID начинается с "admin-".
	return strings.HasPrefix(userID, "admin-"), nil
}

// compile-check
var _ Repo = (*fakeRepo)(nil)

// ---- helpers ----------------------------------------------------------

// newSvc создаёт сервис с fakeRepo, DefaultConfig и mock event-inserter.
// Возвращает (svc, fakeRepo) — состояние можно проверять напрямую.
//
// Принимает testing.TB (а не *testing.T), чтобы работать в обоих режимах:
// обычные тесты и rapid property-тесты (через testingTAdapter).
func newSvc(t testing.TB) (*Service, *fakeRepo) {
	t.Helper()
	fr := newFakeRepo()
	db := &fakeExec{}
	svc := NewService(db, fr, DefaultConfig()).
		WithEventInserter(fakeEventInserter)
	return svc, fr
}

// fakeEventInserter возвращает синтетический event_id без реального INSERT.
func fakeEventInserter(_ context.Context, _ pgx.Tx, _ event.InsertOpts) (string, error) {
	return ids.New(), nil
}

// fakeExec — repo.Exec без реальной БД. InTx вызывает fn с tx=nil
// (fakeRepo игнорирует tx). Pool() возвращает nil — service не вызывает
// его в этих тестах (только repo_pgx использует Pool).
type fakeExec struct{}

func (f *fakeExec) InTx(ctx context.Context, fn repo.TxFunc) error {
	return fn(ctx, nil)
}
func (f *fakeExec) Pool() *pgxpool.Pool { return nil }

// ---- happy paths ------------------------------------------------------

const (
	unitA = 100
	unitB = 200
)

func TestCreateLot_HappyPath(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	for i := 0; i < 5; i++ {
		fr.addArtefact(seller, unitA)
	}
	totalBefore := fr.totalQtyOfUnit(unitA)

	lot, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID:   seller,
		ArtifactUnitID: unitA,
		Quantity:       3,
		PriceOxsarit:   3000,
		ExpiresInHours: 24,
		IdempotencyKey: "k1",
	})
	if err != nil {
		t.Fatalf("create lot: %v", err)
	}
	if lot.Quantity != 3 {
		t.Errorf("quantity = %d, want 3", lot.Quantity)
	}
	if lot.Status != "active" {
		t.Errorf("status = %s, want active", lot.Status)
	}
	// Эскроу-инвариант: total quantity артефактов не изменился.
	if fr.totalQtyOfUnit(unitA) != totalBefore {
		t.Errorf("total quantity changed after CreateLot: %d → %d",
			totalBefore, fr.totalQtyOfUnit(unitA))
	}
	// 3 артефакта в state='listed', 2 в 'held'.
	listed := 0
	held := 0
	for _, a := range fr.artefacts {
		switch a.state {
		case "listed":
			listed++
		case "held":
			held++
		}
	}
	if listed != 3 || held != 2 {
		t.Errorf("listed=%d held=%d, want 3/2", listed, held)
	}
}

func TestCreateLot_InsufficientArtefacts(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	fr.addArtefact(seller, unitA)
	fr.addArtefact(seller, unitA)

	_, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 5, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	if !errors.Is(err, ErrInsufficientArtefacts) {
		t.Fatalf("err = %v, want ErrInsufficientArtefacts", err)
	}
}

func TestCreateLot_InvalidQuantity(t *testing.T) {
	svc, _ := newSvc(t)
	_, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: "u", ArtifactUnitID: unitA,
		Quantity: 0, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("err = %v, want ErrInvalidQuantity", err)
	}
}

func TestCreateLot_MaxQuantity(t *testing.T) {
	svc, _ := newSvc(t)
	_, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: "u", ArtifactUnitID: unitA,
		Quantity: 101, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	if !errors.Is(err, ErrMaxQuantity) {
		t.Fatalf("err = %v, want ErrMaxQuantity", err)
	}
}

func TestCreateLot_InvalidPrice(t *testing.T) {
	svc, _ := newSvc(t)
	_, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: "u", ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 0, ExpiresInHours: 24,
	})
	if !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("err = %v, want ErrInvalidPrice", err)
	}
}

func TestCreateLot_InvalidExpiry(t *testing.T) {
	svc, _ := newSvc(t)
	for _, h := range []int{0, 200} {
		_, err := svc.CreateLot(context.Background(), CreateLotInput{
			SellerUserID: "u", ArtifactUnitID: unitA,
			Quantity: 1, PriceOxsarit: 100, ExpiresInHours: h,
		})
		if !errors.Is(err, ErrInvalidExpiry) {
			t.Fatalf("ExpiresInHours=%d → err=%v, want ErrInvalidExpiry", h, err)
		}
	}
}

func TestCreateLot_PriceCapExceeded(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	fr.addArtefact(seller, unitA)
	// avg = 100, cap = 100*10 = 1000.
	avg := int64(100)
	fr.avgOverride = &avg

	_, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 1500, ExpiresInHours: 24,
	})
	if !errors.Is(err, ErrPriceCapExceeded) {
		t.Fatalf("err = %v, want ErrPriceCapExceeded", err)
	}
}

func TestCreateLot_PriceWithinCap(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	fr.addArtefact(seller, unitA)
	avg := int64(100)
	fr.avgOverride = &avg

	_, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 999, ExpiresInHours: 24,
	})
	if err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestCreateLot_MaxActiveLots(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	for i := 0; i < 11; i++ {
		fr.addArtefact(seller, unitA)
	}
	// Создаём 10 максимально допустимых.
	for i := 0; i < 10; i++ {
		_, err := svc.CreateLot(context.Background(), CreateLotInput{
			SellerUserID: seller, ArtifactUnitID: unitA,
			Quantity: 1, PriceOxsarit: 100, ExpiresInHours: 24,
		})
		if err != nil {
			t.Fatalf("create #%d: %v", i, err)
		}
	}
	// 11-й — отказ.
	_, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	if !errors.Is(err, ErrMaxActiveLots) {
		t.Fatalf("err = %v, want ErrMaxActiveLots", err)
	}
}

func TestCreateLot_PermitDenied(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	fr.addArtefact(seller, unitA)
	svc = svc.WithPermitChecker(denyPermit{})

	_, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	if !errors.Is(err, ErrPermitRequired) {
		t.Fatalf("err = %v, want ErrPermitRequired", err)
	}
}

type denyPermit struct{}

func (denyPermit) HasMerchantPermit(_ context.Context, _ pgx.Tx, _ string) (bool, error) {
	return false, nil
}

func TestBuyLot_HappyPath(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	buyer := "user-B"
	fr.addUser(seller, 0, "planet-S")
	fr.addUser(buyer, 5000, "planet-B")
	fr.addArtefact(seller, unitA)
	fr.addArtefact(seller, unitA)

	totalArtBefore := fr.totalQtyOfUnit(unitA)
	totalOxBefore := fr.totalOxsarits()

	lot, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 2, PriceOxsarit: 1000, ExpiresInHours: 24,
	})
	if err != nil {
		t.Fatal(err)
	}

	bought, err := svc.BuyLot(context.Background(), lot.ID, buyer)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if bought.Status != "sold" {
		t.Errorf("status = %s, want sold", bought.Status)
	}
	if bought.BuyerUserID == nil || *bought.BuyerUserID != buyer {
		t.Errorf("buyer mismatch")
	}

	// Эскроу-инвариант.
	if fr.totalQtyOfUnit(unitA) != totalArtBefore {
		t.Errorf("artefacts disappeared")
	}
	// План 72.1.46 P1#1: per-broker fee удерживается с продавца
	// (DefaultBrokerFee=5%). Buyer платит 1000, seller получает
	// 1000 × (100-5)/100 = 950, 50 oxsarits — системная комиссия
	// (legacy `Exchange.class.php`: exchange_profit идёт брокеру; в
	// origin/nova brokers как отдельных юзеров нет → системная).
	expectedFee := int64(50)
	expectedSellerProfit := int64(1000) - expectedFee
	if fr.totalOxsarits() != totalOxBefore-expectedFee {
		t.Errorf("oxsarits sum: %d → %d, want fee=%d удержан",
			totalOxBefore, fr.totalOxsarits(), expectedFee)
	}
	// Buyer списан полностью, seller получил с вычетом fee.
	if fr.oxsarits[buyer] != 4000 {
		t.Errorf("buyer balance = %d, want 4000", fr.oxsarits[buyer])
	}
	if fr.oxsarits[seller] != expectedSellerProfit {
		t.Errorf("seller balance = %d, want %d", fr.oxsarits[seller], expectedSellerProfit)
	}
	// Артефакты у buyer'а в state='held'.
	heldByBuyer := 0
	for _, a := range fr.artefacts {
		if a.owner == buyer && a.state == "held" {
			heldByBuyer++
		}
	}
	if heldByBuyer != 2 {
		t.Errorf("buyer holds %d, want 2", heldByBuyer)
	}
}

func TestBuyLot_InsufficientOxsarits(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	buyer := "user-B"
	fr.addUser(seller, 0, "planet-S")
	fr.addUser(buyer, 100, "planet-B")
	fr.addArtefact(seller, unitA)

	lot, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 1000, ExpiresInHours: 24,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.BuyLot(context.Background(), lot.ID, buyer)
	if !errors.Is(err, ErrInsufficientOxsarits) {
		t.Fatalf("err = %v, want ErrInsufficientOxsarits", err)
	}
	// Лот должен остаться active.
	if fr.lots[lot.ID].Status != "active" {
		t.Errorf("lot status changed despite failed buy: %s", fr.lots[lot.ID].Status)
	}
}

func TestBuyLot_CannotBuyOwnLot(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	fr.addUser(seller, 5000, "planet-S")
	fr.addArtefact(seller, unitA)

	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 1000, ExpiresInHours: 24,
	})
	_, err := svc.BuyLot(context.Background(), lot.ID, seller)
	if !errors.Is(err, ErrCannotBuyOwnLot) {
		t.Fatalf("err = %v, want ErrCannotBuyOwnLot", err)
	}
}

func TestBuyLot_NotFound(t *testing.T) {
	svc, _ := newSvc(t)
	_, err := svc.BuyLot(context.Background(), "nonexistent", "buyer")
	if !errors.Is(err, ErrLotNotFound) {
		t.Fatalf("err = %v, want ErrLotNotFound", err)
	}
}

func TestBuyLot_LotAlreadySold(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	buyer1 := "user-B1"
	buyer2 := "user-B2"
	fr.addUser(seller, 0, "planet-S")
	fr.addUser(buyer1, 5000, "planet-B1")
	fr.addUser(buyer2, 5000, "planet-B2")
	fr.addArtefact(seller, unitA)

	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 1000, ExpiresInHours: 24,
	})
	if _, err := svc.BuyLot(context.Background(), lot.ID, buyer1); err != nil {
		t.Fatal(err)
	}
	_, err := svc.BuyLot(context.Background(), lot.ID, buyer2)
	if !errors.Is(err, ErrLotNotActive) {
		t.Fatalf("err = %v, want ErrLotNotActive", err)
	}
}

func TestCancelLot_HappyPath(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	fr.addUser(seller, 0, "planet-S")
	fr.addArtefact(seller, unitA)
	fr.addArtefact(seller, unitA)

	totalBefore := fr.totalQtyOfUnit(unitA)
	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 2, PriceOxsarit: 500, ExpiresInHours: 24,
	})
	if err := svc.CancelLot(context.Background(), lot.ID, seller); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	// Все артефакты обратно в 'held' у seller'а.
	heldBySeller := 0
	for _, a := range fr.artefacts {
		if a.owner == seller && a.state == "held" {
			heldBySeller++
		}
	}
	if heldBySeller != totalBefore {
		t.Errorf("expected all %d artefacts back to held, got %d", totalBefore, heldBySeller)
	}
	if fr.lots[lot.ID].Status != "cancelled" {
		t.Errorf("status = %s, want cancelled", fr.lots[lot.ID].Status)
	}
}

func TestCancelLot_NotASeller(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	stranger := "user-X"
	fr.addUser(seller, 0, "planet-S")
	fr.addUser(stranger, 0, "planet-X")
	fr.addArtefact(seller, unitA)

	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	err := svc.CancelLot(context.Background(), lot.ID, stranger)
	if !errors.Is(err, ErrNotASeller) {
		t.Fatalf("err = %v, want ErrNotASeller", err)
	}
}

func TestCancelLot_AlreadyCancelled(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	fr.addUser(seller, 0, "planet-S")
	fr.addArtefact(seller, unitA)
	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	if err := svc.CancelLot(context.Background(), lot.ID, seller); err != nil {
		t.Fatal(err)
	}
	err := svc.CancelLot(context.Background(), lot.ID, seller)
	if !errors.Is(err, ErrLotNotActive) {
		t.Fatalf("err = %v, want ErrLotNotActive", err)
	}
}

func TestBuyLot_BuyerHasNoPlanet(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	buyer := "user-B"
	fr.addUser(seller, 0, "planet-S")
	fr.addUser(buyer, 5000, "") // home==""
	fr.addArtefact(seller, unitA)
	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: unitA,
		Quantity: 1, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	_, err := svc.BuyLot(context.Background(), lot.ID, buyer)
	if !errors.Is(err, ErrUserHasNoPlanet) {
		t.Fatalf("err = %v, want ErrUserHasNoPlanet", err)
	}
}
