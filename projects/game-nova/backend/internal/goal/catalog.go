package goal

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// Catalog — загруженный набор GoalDef. Immutable после Load.
//
// Включает вспомогательные индексы:
//   - byCategory:  для UI List по фильтру (achievement / daily / ...).
//   - byEventKind: для counter-целей: на event_kind=10 (KindAttackSingle)
//                  быстро найти все интересующиеся goals без скана всех.
type Catalog struct {
	byKey       map[string]GoalDef
	byCategory  map[Category][]string
	byEventKind map[int][]string // для counter-целей с условием event_count
	keysSorted  []string         // стабильный порядок для UI
}

// LoadCatalog читает configs/goals.yml и валидирует структуру.
//
// Если файла нет — возвращает пустой Catalog (без ошибки), как
// features.Load — фича может быть отключена feature flag'ом, и пустой
// каталог — валидное состояние для startup.
func LoadCatalog(path string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Catalog{
				byKey:       map[string]GoalDef{},
				byCategory:  map[Category][]string{},
				byEventKind: map[int][]string{},
			}, nil
		}
		return nil, fmt.Errorf("goal: read %s: %w", path, err)
	}
	return ParseCatalog(data)
}

// ParseCatalog — низкоуровневый парсинг (используется и в тестах).
func ParseCatalog(data []byte) (*Catalog, error) {
	var raw struct {
		Goals map[string]GoalDef `yaml:"goals"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("goal: parse yaml: %w", err)
	}
	if raw.Goals == nil {
		raw.Goals = map[string]GoalDef{}
	}

	c := &Catalog{
		byKey:       make(map[string]GoalDef, len(raw.Goals)),
		byCategory:  map[Category][]string{},
		byEventKind: map[int][]string{},
	}

	// 1. Заполняем byKey, проверяем минимальные инварианты.
	for key, def := range raw.Goals {
		def.Key = key
		if err := validateDef(def); err != nil {
			return nil, fmt.Errorf("goal %q: %w", key, err)
		}
		c.byKey[key] = def
	}

	// 2. Проверяем requires (все ключи существуют, нет циклов).
	for key, def := range c.byKey {
		for _, dep := range def.Requires {
			if _, ok := c.byKey[dep]; !ok {
				return nil, fmt.Errorf("goal %q: requires unknown goal %q", key, dep)
			}
		}
	}
	if cycle := findCycle(c.byKey); cycle != "" {
		return nil, fmt.Errorf("goal: dependency cycle detected at %q", cycle)
	}

	// 3. Индексы.
	for key, def := range c.byKey {
		c.byCategory[def.Category] = append(c.byCategory[def.Category], key)
		// event_count условия — индексируем по event_kind для быстрого
		// OnEvent (см. goal/conditions/event_count.go).
		if def.Condition.Type == "event_count" {
			if kind, ok := extractEventKind(def.Condition.Params); ok {
				c.byEventKind[kind] = append(c.byEventKind[kind], key)
			}
		}
		c.keysSorted = append(c.keysSorted, key)
	}
	sort.Strings(c.keysSorted)
	for cat := range c.byCategory {
		sort.Strings(c.byCategory[cat])
	}

	return c, nil
}

// All возвращает копию map (защита от мутаций извне).
func (c *Catalog) All() map[string]GoalDef {
	out := make(map[string]GoalDef, len(c.byKey))
	for k, v := range c.byKey {
		out[k] = v
	}
	return out
}

// Get — найти определение по ключу.
func (c *Catalog) Get(key string) (GoalDef, bool) {
	def, ok := c.byKey[key]
	return def, ok
}

// ByCategory — отсортированный список ключей в категории.
func (c *Catalog) ByCategory(cat Category) []string {
	out := make([]string, len(c.byCategory[cat]))
	copy(out, c.byCategory[cat])
	return out
}

// ByEventKind — список ключей counter-целей, реагирующих на event_kind.
// Используется в Engine.OnEvent.
func (c *Catalog) ByEventKind(kind int) []string {
	out := make([]string, len(c.byEventKind[kind]))
	copy(out, c.byEventKind[kind])
	return out
}

// Keys — все ключи в каталоге (отсортированы).
func (c *Catalog) Keys() []string {
	out := make([]string, len(c.keysSorted))
	copy(out, c.keysSorted)
	return out
}

// Len — сколько целей в каталоге.
func (c *Catalog) Len() int { return len(c.byKey) }

// validateDef проверяет минимальные инварианты GoalDef.
func validateDef(def GoalDef) error {
	if def.Title == "" {
		return errors.New("title is required")
	}
	switch def.Category {
	case CategoryAchievement, CategoryStarter, CategoryDaily, CategoryWeekly, CategoryEvent:
	default:
		return fmt.Errorf("invalid category %q", def.Category)
	}
	switch def.Lifecycle {
	case LifecyclePermanent, LifecycleOneTime, LifecycleDaily, LifecycleWeekly,
		LifecycleSeasonal, LifecycleRepeatable:
	default:
		return fmt.Errorf("invalid lifecycle %q", def.Lifecycle)
	}
	if def.Condition.Type == "" {
		return errors.New("condition.type is required")
	}
	if def.Lifecycle == LifecycleSeasonal && (def.ActiveFrom == nil || def.ActiveUntil == nil) {
		return errors.New("seasonal lifecycle requires active_from and active_until")
	}
	return nil
}

// findCycle ищет цикл в графе requires. Возвращает первый ключ цикла,
// либо "" если циклов нет.
func findCycle(goals map[string]GoalDef) string {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	state := make(map[string]int, len(goals))
	var dfs func(string) string
	dfs = func(node string) string {
		state[node] = gray
		for _, dep := range goals[node].Requires {
			switch state[dep] {
			case gray:
				return node
			case white:
				if c := dfs(dep); c != "" {
					return c
				}
			}
		}
		state[node] = black
		return ""
	}
	for key := range goals {
		if state[key] == white {
			if c := dfs(key); c != "" {
				return c
			}
		}
	}
	return ""
}

// extractEventKind — достать event_kind из params условия event_count.
// Используется при индексировании Catalog. Если params неправильного
// формата — возвращает (0, false), goal не попадает в byEventKind, но
// всё равно работает через общий путь Recompute.
func extractEventKind(params map[string]any) (int, bool) {
	if params == nil {
		return 0, false
	}
	v, ok := params["event_kind"]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, n > 0
	case int64:
		return int(n), n > 0
	case float64:
		return int(n), n > 0
	default:
		return 0, false
	}
}
