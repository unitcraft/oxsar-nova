package referral

import (
	"testing"
)

func TestConstants(t *testing.T) {
	t.Parallel()
	if CreditPercent != 0.20 {
		t.Fatalf("CreditPercent must be 0.20, got %v", CreditPercent)
	}
	if BonusPoints != 3000 {
		t.Fatalf("BonusPoints must be 3000, got %d", BonusPoints)
	}
	if MaxBonusPoints != 500000 {
		t.Fatalf("MaxBonusPoints must be 500000, got %d", MaxBonusPoints)
	}
}

func TestStartingResources(t *testing.T) {
	t.Parallel()
	if StartingMetal != 10 || StartingSilicon != 5 || StartingHydrogen != 2 {
		t.Fatalf("starting resources: metal=%d silicon=%d hydrogen=%d",
			StartingMetal, StartingSilicon, StartingHydrogen)
	}
}

func TestProcessPurchaseBonus(t *testing.T) {
	t.Parallel()
	// 20% от 1000 = 200.
	amount := 1000.0
	bonus := amount * CreditPercent
	if bonus != 200.0 {
		t.Fatalf("bonus from 1000cr purchase should be 200, got %v", bonus)
	}
}

func TestErrReferrerNotFound_IsSentinel(t *testing.T) {
	t.Parallel()
	if ErrReferrerNotFound == nil {
		t.Fatal("ErrReferrerNotFound must not be nil")
	}
}
