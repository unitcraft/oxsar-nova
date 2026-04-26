package planet

import "testing"

// seedFromUserID — стабильна: для того же userID тот же seed, и разные
// userID почти всегда дают разные seed'ы. Это не для криптографии —
// просто чтобы тесты шли предсказуемо.
func TestSeedFromUserID_Stable(t *testing.T) {
	t.Parallel()
	a := seedFromUserID("user-1")
	b := seedFromUserID("user-1")
	if a != b {
		t.Fatalf("seed not stable: %d vs %d", a, b)
	}
}

func TestSeedFromUserID_DifferentUsersDiverge(t *testing.T) {
	t.Parallel()
	ids := []string{
		"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta",
	}
	seen := map[uint64]string{}
	for _, id := range ids {
		s := seedFromUserID(id)
		if existing, ok := seen[s]; ok {
			t.Fatalf("seed collision for %q and %q", existing, id)
		}
		seen[s] = id
	}
}
