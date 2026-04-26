package economy

import "testing"

func TestStorageCapacity_Level0_BaseOnly(t *testing.T) {
	t.Parallel()
	got := StorageCapacity(5000, 0, 1.0)
	if got != 5000 {
		t.Fatalf("expected 5000 at level 0, got %v", got)
	}
}

func TestStorageCapacity_Monotonic(t *testing.T) {
	t.Parallel()
	prev := StorageCapacity(5000, 0, 1.0)
	for lvl := 1; lvl <= 20; lvl++ {
		cur := StorageCapacity(5000, lvl, 1.0)
		if cur <= prev {
			t.Fatalf("storage cap not monotonic at level %d: %v <= %v", lvl, cur, prev)
		}
		prev = cur
	}
}

func TestStorageCapacity_FactorApplies(t *testing.T) {
	t.Parallel()
	base := StorageCapacity(5000, 5, 1.0)
	boosted := StorageCapacity(5000, 5, 1.15) // ATOMIC_DENSIFIER активен
	if boosted <= base {
		t.Fatalf("factor must increase cap: base=%v boosted=%v", base, boosted)
	}
	// Проверяем точный множитель.
	want := base * 1.15
	if got := boosted; got < want-1e-6 || got > want+1e-6 {
		t.Fatalf("factor must be linear: want %v, got %v", want, got)
	}
}
