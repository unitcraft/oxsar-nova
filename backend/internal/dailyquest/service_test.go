package dailyquest

import (
	"math/rand"
	"testing"
)

func TestPickWeighted_AllItemsIfNGE(t *testing.T) {
	items := []defRow{{1, 100}, {2, 50}}
	r := rand.New(rand.NewSource(1))
	got := pickWeighted(items, 5, r)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestPickWeighted_ReturnsExactlyN(t *testing.T) {
	items := []defRow{{1, 10}, {2, 10}, {3, 10}, {4, 10}, {5, 10}}
	r := rand.New(rand.NewSource(42))
	got := pickWeighted(items, 3, r)
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d", len(got))
	}
	// Все должны быть уникальны.
	seen := map[int]bool{}
	for _, it := range got {
		if seen[it.id] {
			t.Errorf("duplicate id %d", it.id)
		}
		seen[it.id] = true
	}
}

func TestPickWeighted_RespectsWeight(t *testing.T) {
	// Item 1 имеет вес 1000, item 2 — 1. На 1000 прогонов
	// item 1 должен попасть в выборку почти всегда.
	items := []defRow{{1, 1000}, {2, 1}}
	hit1 := 0
	for s := int64(0); s < 1000; s++ {
		r := rand.New(rand.NewSource(s))
		got := pickWeighted(items, 1, r)
		if len(got) > 0 && got[0].id == 1 {
			hit1++
		}
	}
	if hit1 < 950 {
		t.Errorf("expected item 1 to be picked >=950 times, got %d", hit1)
	}
}
