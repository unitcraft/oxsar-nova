package balance

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// configsRoot — путь к корневому configs/ из текущего пакета.
//
// Тесты загружают реальный nova-каталог (источник истины modern-баланса).
// Это критерий приёма Ф.1 (план 64): LoadDefaults() == текущая
// конфигурация nova, ничего не сломали.
const configsRoot = "../../../configs"

func TestLoadDefaults_ReturnsModernCatalog(t *testing.T) {
	t.Parallel()
	l := NewLoader(configsRoot)
	b, err := l.LoadDefaults()
	if err != nil {
		t.Fatalf("LoadDefaults: %v", err)
	}
	if b == nil {
		t.Fatal("LoadDefaults returned nil bundle")
	}
	if b.Catalog == nil {
		t.Fatal("Bundle.Catalog == nil")
	}
	if b.HasOverride {
		t.Error("HasOverride must be false for defaults")
	}
	if b.UniverseID != "" {
		t.Errorf("UniverseID for defaults must be empty, got %q", b.UniverseID)
	}

	// Modern-числа из buildings.yml: metal_mine cost_base 60/15/0.
	mm, ok := b.Catalog.Buildings.Buildings["metal_mine"]
	if !ok {
		t.Fatal("metal_mine missing from default catalog")
	}
	if mm.CostBase.Metal != 60 {
		t.Errorf("metal_mine.cost_base.metal = %d, want 60", mm.CostBase.Metal)
	}
	if mm.CostBase.Silicon != 15 {
		t.Errorf("metal_mine.cost_base.silicon = %d, want 15", mm.CostBase.Silicon)
	}
	if mm.CostFactor != 1.5 {
		t.Errorf("metal_mine.cost_factor = %v, want 1.5", mm.CostFactor)
	}

	// Modern-globals соответствуют existing economy/formulas.go.
	if b.Globals.MetalMineBasicProd != 30 {
		t.Errorf("Globals.MetalMineBasicProd = %v, want 30", b.Globals.MetalMineBasicProd)
	}
	if b.Globals.HydrogenTempCoefficient != -0.002 {
		t.Errorf("HydrogenTempCoefficient = %v, want -0.002", b.Globals.HydrogenTempCoefficient)
	}
	if b.Globals.HydrogenTempIntercept != 1.28 {
		t.Errorf("HydrogenTempIntercept = %v, want 1.28", b.Globals.HydrogenTempIntercept)
	}
}

func TestLoadDefaults_Cached(t *testing.T) {
	t.Parallel()
	l := NewLoader(configsRoot)
	b1, err := l.LoadDefaults()
	if err != nil {
		t.Fatal(err)
	}
	b2, err := l.LoadDefaults()
	if err != nil {
		t.Fatal(err)
	}
	if b1 != b2 {
		t.Error("LoadDefaults must return the same pointer on repeated calls")
	}
}

func TestLoadFor_NoOverride_ReturnsDefaults(t *testing.T) {
	t.Parallel()
	l := NewLoader(configsRoot)
	b, err := l.LoadFor("uni01")
	if err != nil {
		t.Fatalf("LoadFor uni01: %v", err)
	}
	if b.HasOverride {
		t.Error("uni01 must NOT have override (R0: modern-вселенные на дефолте)")
	}
	if b.UniverseID != "uni01" {
		t.Errorf("UniverseID = %q, want uni01", b.UniverseID)
	}
	def, _ := l.LoadDefaults()
	if b.Catalog != def.Catalog {
		t.Error("uni01 should share default catalog (no override)")
	}
}

func TestLoadFor_WithOverride(t *testing.T) {
	t.Parallel()

	// Готовим временный configs-tree с override-файлом.
	tmpRoot := tmpConfigsTree(t)

	l := NewLoader(tmpRoot)
	b, err := l.LoadFor("origin")
	if err != nil {
		t.Fatalf("LoadFor origin: %v", err)
	}
	if !b.HasOverride {
		t.Fatal("origin must have override")
	}
	if b.UniverseID != "origin" {
		t.Errorf("UniverseID = %q, want origin", b.UniverseID)
	}

	// Override переопределил metal_mine cost_factor 1.5 → 1.6.
	mm := b.Catalog.Buildings.Buildings["metal_mine"]
	if mm.CostFactor != 1.6 {
		t.Errorf("origin metal_mine.cost_factor = %v, want 1.6 (overridden)", mm.CostFactor)
	}
	// А cost_base остался дефолтным (override его не трогал).
	if mm.CostBase.Metal != 60 {
		t.Errorf("origin metal_mine.cost_base.metal = %d, want 60 (inherited)", mm.CostBase.Metal)
	}
	// Globals override: hydrogen_temp_intercept 1.28 → 1.30.
	if b.Globals.HydrogenTempIntercept != 1.30 {
		t.Errorf("origin Globals.HydrogenTempIntercept = %v, want 1.30 (overridden)", b.Globals.HydrogenTempIntercept)
	}
	// Globals не-overridden остался дефолтным.
	if b.Globals.MetalMineBasicProd != 30 {
		t.Errorf("origin Globals.MetalMineBasicProd = %v, want 30 (inherited)", b.Globals.MetalMineBasicProd)
	}

	// Проверяем, что default bundle НЕ испорчен override-merge'ем
	// (immutability default-Catalog).
	def, _ := l.LoadDefaults()
	if def.Catalog.Buildings.Buildings["metal_mine"].CostFactor != 1.5 {
		t.Errorf("default metal_mine cost_factor должен остаться 1.5 после origin override, got %v",
			def.Catalog.Buildings.Buildings["metal_mine"].CostFactor)
	}
	if def.Globals.HydrogenTempIntercept != 1.28 {
		t.Error("default Globals.HydrogenTempIntercept должен остаться 1.28")
	}
}

func TestLoadFor_OverrideUnknownBuildingFails(t *testing.T) {
	t.Parallel()

	tmpRoot := tmpConfigsTreeRaw(t, `version: 1
universe: bogus
buildings:
  unknown_building:
    cost_factor: 2.0
`, "bogus")

	l := NewLoader(tmpRoot)
	_, err := l.LoadFor("bogus")
	if err == nil {
		t.Fatal("expected error for unknown building in override, got nil")
	}
	if !errors.Is(err, ErrInvalidOverride) {
		t.Errorf("error must wrap ErrInvalidOverride, got %v", err)
	}
}

func TestLoadFor_OverrideUniverseMismatchFails(t *testing.T) {
	t.Parallel()

	tmpRoot := tmpConfigsTreeRaw(t, `version: 1
universe: not_origin
`, "origin")

	l := NewLoader(tmpRoot)
	_, err := l.LoadFor("origin")
	if err == nil {
		t.Fatal("expected error for universe mismatch, got nil")
	}
	if !errors.Is(err, ErrInvalidOverride) {
		t.Errorf("error must wrap ErrInvalidOverride, got %v", err)
	}
}

func TestLoadFor_OverrideMalformedFails(t *testing.T) {
	t.Parallel()

	tmpRoot := tmpConfigsTreeRaw(t, "version: not-a-number\n", "broken")
	l := NewLoader(tmpRoot)
	_, err := l.LoadFor("broken")
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !errors.Is(err, ErrInvalidOverride) {
		t.Errorf("error must wrap ErrInvalidOverride, got %v", err)
	}
}

func TestLoadFor_Cached(t *testing.T) {
	t.Parallel()
	l := NewLoader(configsRoot)
	a, err := l.LoadFor("uni01")
	if err != nil {
		t.Fatal(err)
	}
	b, err := l.LoadFor("uni01")
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Error("LoadFor must return cached bundle")
	}
}

func TestLoadFor_ConcurrentSafe(t *testing.T) {
	t.Parallel()
	l := NewLoader(configsRoot)

	const n = 32
	var wg sync.WaitGroup
	bundles := make([]*Bundle, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			b, err := l.LoadFor("uni01")
			if err != nil {
				t.Errorf("goroutine %d: %v", idx, err)
				return
			}
			bundles[idx] = b
		}(i)
	}
	wg.Wait()

	for i := 1; i < n; i++ {
		if bundles[i] != bundles[0] {
			t.Errorf("concurrent LoadFor returned different bundles: %p vs %p", bundles[i], bundles[0])
			break
		}
	}
}

func TestLoadForCtx_LogsAndMetrics(t *testing.T) {
	t.Parallel()
	l := NewLoader(configsRoot)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	b, err := l.LoadForCtx(context.Background(), log, "uni01")
	if err != nil {
		t.Fatal(err)
	}
	if b == nil || b.UniverseID != "uni01" {
		t.Errorf("expected uni01 bundle, got %+v", b)
	}
	// Метрики не валидируем по значениям (тест зависит от глобального
	// registry); важно что вызов не упал.
}

// --- helpers ---

// tmpConfigsTree собирает временный configs/ + override-файл с
// валидным содержимым для TestLoadFor_WithOverride.
func tmpConfigsTree(t *testing.T) string {
	t.Helper()
	override := `version: 1
universe: origin
globals:
  hydrogen_temp_intercept: 1.30
buildings:
  metal_mine:
    cost_factor: 1.6
`
	return tmpConfigsTreeRaw(t, override, "origin")
}

// tmpConfigsTreeRaw создаёт tmp/configs/ с симлинком на реальный
// nova-configs (или копией) + balance/<id>.yaml = body.
func tmpConfigsTreeRaw(t *testing.T, body, id string) string {
	t.Helper()
	dir := t.TempDir()

	// Копируем все YAML из настоящего configs/ — Loader читает их все.
	abs, err := filepath.Abs(configsRoot)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(abs, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// balance/ субдиректория.
	balDir := filepath.Join(dir, "balance")
	if err := os.MkdirAll(balDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(balDir, id+".yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}
