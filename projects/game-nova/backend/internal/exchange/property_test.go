package exchange

import (
	"context"
	"testing"

	"pgregory.net/rapid"
)

// Property-based тесты биржи (R4).
//
// Инвариант 1 (escrow): total quantity артефактов unit_id в системе
// (sum across users + sum across active lot_items) constant при любой
// последовательности create/buy/cancel.
//
// Инвариант 2 (oxsarit): сумма oxsarit участников константна при buy
// (просто перетекает от buyer'а к seller'у).
//
// Инвариант 3 (price-cap detection): для одного и того же reference
// price_cap_exceeded — детерминированное условие
// (unitPrice > reference * multiplier).

const (
	maxUsers      = 5
	maxArtefacts  = 50
	maxOpsPerCase = 20
)

// genWorld создаёт случайную начальную ситуацию: N пользователей с
// артефактами и оксаритами.
type world struct {
	svc       *Service
	repo      *fakeRepo
	users     []string
	unitID    int
	startQty  int
	startOxsa int64
}

// newSvcRaw — newSvc без testing.TB-зависимости (для property-тестов).
func newSvcRaw() (*Service, *fakeRepo) {
	fr := newFakeRepo()
	db := &fakeExec{}
	svc := NewService(db, fr, DefaultConfig()).
		WithEventInserter(fakeEventInserter)
	return svc, fr
}

func makeWorld(t *rapid.T) *world {
	rng := rapid.IntRange(2, maxUsers).Draw(t, "users")
	users := make([]string, rng)
	for i := 0; i < rng; i++ {
		users[i] = "u" + rapid.StringMatching(`[a-z0-9]{3,5}`).Draw(t, "uid")
	}
	svc, fr := newSvcRaw()
	startOx := int64(rapid.IntRange(10000, 100000).Draw(t, "startOx"))
	for _, u := range users {
		fr.addUser(u, startOx, "p"+u)
	}
	unitID := 100
	totalArt := rapid.IntRange(rng*3, maxArtefacts).Draw(t, "art")
	for i := 0; i < totalArt; i++ {
		owner := users[i%len(users)]
		fr.addArtefact(owner, unitID)
	}
	return &world{
		svc:       svc,
		repo:      fr,
		users:     users,
		unitID:    unitID,
		startQty:  totalArt,
		startOxsa: startOx * int64(len(users)),
	}
}

// rapidT просто прокидывает rapid.T (тесты используют newSvc, которая
// принимает testing.TB). Так как testing.TB — закрытый интерфейс, мы
// не можем имплементировать его извне; вместо этого создаём service
// напрямую через NewService без хелпера.

// PropertyEscrowInvariant: total quantity артефактов unit_id неизменно
// при любой последовательности успешных create/buy/cancel.
func TestProperty_EscrowInvariant(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		w := makeWorld(rt)
		startTotal := w.repo.totalQtyOfUnit(w.unitID)
		ops := rapid.IntRange(1, maxOpsPerCase).Draw(rt, "ops")
		ctx := context.Background()
		for i := 0; i < ops; i++ {
			op := rapid.IntRange(0, 2).Draw(rt, "op")
			switch op {
			case 0:
				// Create.
				seller := w.users[rapid.IntRange(0, len(w.users)-1).Draw(rt, "seller")]
				qty := rapid.IntRange(1, 5).Draw(rt, "qty")
				price := int64(rapid.IntRange(1, 10000).Draw(rt, "price"))
				_, _ = w.svc.CreateLot(ctx, CreateLotInput{
					SellerUserID: seller, ArtifactUnitID: w.unitID,
					Quantity: qty, PriceOxsarit: price, ExpiresInHours: 24,
				})
			case 1:
				// Buy первый active lot.
				buyer := w.users[rapid.IntRange(0, len(w.users)-1).Draw(rt, "buyer")]
				lots, _, _ := w.svc.ListLots(ctx, ListFilters{Limit: 100})
				if len(lots) == 0 {
					continue
				}
				_, _ = w.svc.BuyLot(ctx, lots[0].ID, buyer)
			case 2:
				// Cancel первый active lot своего seller'а.
				lots, _, _ := w.svc.ListLots(ctx, ListFilters{Limit: 100})
				if len(lots) == 0 {
					continue
				}
				_ = w.svc.CancelLot(ctx, lots[0].ID, lots[0].SellerUserID)
			}
		}
		got := w.repo.totalQtyOfUnit(w.unitID)
		if got != startTotal {
			rt.Fatalf("escrow invariant violated: total %d → %d", startTotal, got)
		}
	})
}

// PropertyOxsaritInvariant: сумма oxsarit участников константна (она не
// уходит в null/system).
func TestProperty_OxsaritInvariant(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		w := makeWorld(rt)
		startSum := w.repo.totalOxsarits()
		ops := rapid.IntRange(1, maxOpsPerCase).Draw(rt, "ops")
		ctx := context.Background()
		for i := 0; i < ops; i++ {
			op := rapid.IntRange(0, 2).Draw(rt, "op")
			switch op {
			case 0:
				seller := w.users[rapid.IntRange(0, len(w.users)-1).Draw(rt, "seller")]
				qty := rapid.IntRange(1, 5).Draw(rt, "qty")
				price := int64(rapid.IntRange(1, 5000).Draw(rt, "price"))
				_, _ = w.svc.CreateLot(ctx, CreateLotInput{
					SellerUserID: seller, ArtifactUnitID: w.unitID,
					Quantity: qty, PriceOxsarit: price, ExpiresInHours: 24,
				})
			case 1:
				buyer := w.users[rapid.IntRange(0, len(w.users)-1).Draw(rt, "buyer")]
				lots, _, _ := w.svc.ListLots(ctx, ListFilters{Limit: 100})
				if len(lots) == 0 {
					continue
				}
				_, _ = w.svc.BuyLot(ctx, lots[0].ID, buyer)
			case 2:
				lots, _, _ := w.svc.ListLots(ctx, ListFilters{Limit: 100})
				if len(lots) == 0 {
					continue
				}
				_ = w.svc.CancelLot(ctx, lots[0].ID, lots[0].SellerUserID)
			}
		}
		got := w.repo.totalOxsarits()
		if got != startSum {
			rt.Fatalf("oxsarit invariant violated: sum %d → %d", startSum, got)
		}
	})
}

// PropertyPriceCapDeterministic: для одного и того же reference price-cap
// detection не зависит от порядка проверок — детерминирован.
func TestProperty_PriceCapDeterministic(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		svc, fr := newSvcRaw()
		seller := "seller"
		fr.addUser(seller, 0, "p")
		fr.addArtefact(seller, 100)

		ref := int64(rapid.IntRange(10, 1000).Draw(rt, "ref"))
		fr.avgOverride = &ref
		mult := svc.cfg.PriceCapMultiplier
		cap := int64(float64(ref) * mult)

		price := int64(rapid.IntRange(1, int(cap*2)).Draw(rt, "price"))
		// При quantity=1 unit_price = price.
		_, err := svc.CreateLot(context.Background(), CreateLotInput{
			SellerUserID: seller, ArtifactUnitID: 100,
			Quantity: 1, PriceOxsarit: price, ExpiresInHours: 24,
		})
		shouldFail := price > cap
		got := err != nil
		if shouldFail != got {
			rt.Fatalf("price=%d cap=%d ref=%d: shouldFail=%v gotErr=%v",
				price, cap, ref, shouldFail, err)
		}
	})
}
