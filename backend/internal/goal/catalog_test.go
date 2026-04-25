package goal_test

import (
	"testing"

	"github.com/oxsar/nova/backend/internal/goal"
	// Подключаем conditions, чтобы зарегистрировались типы для валидации.
	_ "github.com/oxsar/nova/backend/internal/goal/conditions"
)

// Локальные алиасы для краткости.
type (
	Catalog          = goal.Catalog
	Category         = goal.Category
	Lifecycle        = goal.Lifecycle
	GoalDef          = goal.GoalDef
)

const (
	CategoryAchievement = goal.CategoryAchievement
)

var ParseCatalog = goal.ParseCatalog

func TestParseCatalog_Empty(t *testing.T) {
	c, err := ParseCatalog([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if c.Len() != 0 {
		t.Errorf("expected empty catalog, got %d", c.Len())
	}
}

func TestParseCatalog_Sample(t *testing.T) {
	yaml := []byte(`
goals:
  FIRST_METAL:
    title: "Первая шахта"
    category: achievement
    lifecycle: permanent
    condition:
      type: building_level
      params: { unit_id: 1, min_level: 1 }
    reward:
      credits: 10
`)
	c, err := ParseCatalog(yaml)
	if err != nil {
		t.Fatal(err)
	}
	if c.Len() != 1 {
		t.Fatalf("expected 1 goal, got %d", c.Len())
	}
	g, ok := c.Get("FIRST_METAL")
	if !ok {
		t.Fatal("FIRST_METAL not found")
	}
	if g.Title != "Первая шахта" {
		t.Errorf("title: %q", g.Title)
	}
	if g.Category != CategoryAchievement {
		t.Errorf("category: %q", g.Category)
	}
	if g.Reward.Credits != 10 {
		t.Errorf("credits: %d", g.Reward.Credits)
	}
}

func TestParseCatalog_RequiresValidation(t *testing.T) {
	yaml := []byte(`
goals:
  FIRST:
    title: A
    category: starter
    lifecycle: one-time
    condition: { type: building_level, params: { unit_id: 1, min_level: 1 } }
  SECOND:
    title: B
    category: starter
    lifecycle: one-time
    condition: { type: building_level, params: { unit_id: 2, min_level: 1 } }
    requires: [FIRST]
`)
	if _, err := ParseCatalog(yaml); err != nil {
		t.Errorf("valid graph should parse: %v", err)
	}
}

func TestParseCatalog_RequiresMissing(t *testing.T) {
	yaml := []byte(`
goals:
  A:
    title: A
    category: starter
    lifecycle: one-time
    condition: { type: building_level, params: { unit_id: 1, min_level: 1 } }
    requires: [GHOST]
`)
	if _, err := ParseCatalog(yaml); err == nil {
		t.Error("expected error for unknown requires")
	}
}

func TestParseCatalog_RequiresCycle(t *testing.T) {
	yaml := []byte(`
goals:
  A:
    title: A
    category: starter
    lifecycle: one-time
    condition: { type: building_level, params: { unit_id: 1, min_level: 1 } }
    requires: [B]
  B:
    title: B
    category: starter
    lifecycle: one-time
    condition: { type: building_level, params: { unit_id: 2, min_level: 1 } }
    requires: [A]
`)
	if _, err := ParseCatalog(yaml); err == nil {
		t.Error("expected cycle error")
	}
}

func TestParseCatalog_InvalidCategory(t *testing.T) {
	yaml := []byte(`
goals:
  A:
    title: A
    category: bogus
    lifecycle: permanent
    condition: { type: building_level, params: { unit_id: 1, min_level: 1 } }
`)
	if _, err := ParseCatalog(yaml); err == nil {
		t.Error("expected category validation error")
	}
}

func TestParseCatalog_SeasonalRequiresWindow(t *testing.T) {
	yaml := []byte(`
goals:
  S:
    title: S
    category: event
    lifecycle: seasonal
    condition: { type: building_level, params: { unit_id: 1, min_level: 1 } }
`)
	if _, err := ParseCatalog(yaml); err == nil {
		t.Error("seasonal without active_from/until should fail")
	}
}

func TestCatalog_ByEventKindIndex(t *testing.T) {
	yaml := []byte(`
goals:
  daily_attack:
    title: A
    category: daily
    lifecycle: daily
    condition: { type: event_count, params: { event_kind: 10 } }
    target: 1
  daily_spy:
    title: B
    category: daily
    lifecycle: daily
    condition: { type: event_count, params: { event_kind: 11 } }
    target: 1
`)
	c, err := ParseCatalog(yaml)
	if err != nil {
		t.Fatal(err)
	}
	got := c.ByEventKind(10)
	if len(got) != 1 || got[0] != "daily_attack" {
		t.Errorf("ByEventKind(10): %v", got)
	}
	got = c.ByEventKind(11)
	if len(got) != 1 || got[0] != "daily_spy" {
		t.Errorf("ByEventKind(11): %v", got)
	}
	if len(c.ByEventKind(99)) != 0 {
		t.Error("ByEventKind for unknown should be empty")
	}
}

func TestCatalog_ByCategorySorted(t *testing.T) {
	yaml := []byte(`
goals:
  Z:
    title: Z
    category: achievement
    lifecycle: permanent
    condition: { type: building_level, params: { unit_id: 1, min_level: 1 } }
  A:
    title: A
    category: achievement
    lifecycle: permanent
    condition: { type: building_level, params: { unit_id: 2, min_level: 1 } }
`)
	c, _ := ParseCatalog(yaml)
	keys := c.ByCategory(CategoryAchievement)
	if len(keys) != 2 || keys[0] != "A" || keys[1] != "Z" {
		t.Errorf("expected sorted [A Z], got %v", keys)
	}
}
