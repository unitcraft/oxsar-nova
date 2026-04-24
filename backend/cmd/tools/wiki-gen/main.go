// Command wiki-gen — генерирует docs/wiki/ru/{buildings,ships,defense,research}/*.md
// из configs/*.yml. Числа живут в configs, вики — только текст + числа.
//
// Запуск:
//
//	go run ./cmd/tools/wiki-gen/ --configs=../configs --out=../docs/wiki/ru
//
// При необходимости перегенерировать — просто запустить ещё раз; файлы
// перезаписываются. Ручные статьи (getting-started/, combat/, missions/
// и другие *.md, не попадающие под генерируемые категории) не трогаются.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oxsar/nova/backend/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "wiki-gen:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configsDir = flag.String("configs", "../../../configs", "путь к configs/")
		outDir     = flag.String("out", "../../../docs/wiki/ru", "куда писать wiki/ru")
	)
	flag.Parse()

	cat, err := config.LoadCatalog(*configsDir)
	if err != nil {
		return fmt.Errorf("load catalog: %w", err)
	}

	if err := genBuildings(cat, *outDir); err != nil {
		return err
	}
	if err := genShips(cat, *outDir); err != nil {
		return err
	}
	if err := genDefense(cat, *outDir); err != nil {
		return err
	}
	if err := genResearch(cat, *outDir); err != nil {
		return err
	}
	return nil
}

// --- Buildings ---

func genBuildings(cat *config.Catalog, outDir string) error {
	dir := filepath.Join(outDir, "buildings")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Список ключей для стабильного порядка.
	keys := sortedKeys(cat.Buildings.Buildings)
	indexB := &strings.Builder{}
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB, "title: Здания")
	fmt.Fprintln(indexB, "category: buildings")
	fmt.Fprintln(indexB, "order: 20")
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "# Здания")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "Полный список построек. Числа берутся из `configs/buildings.yml` и `configs/construction.yml`.")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "| ID | Ключ | Базовая стоимость | Макс. уровень |")
	fmt.Fprintln(indexB, "|---:|---|---|---:|")

	for _, key := range keys {
		b := cat.Buildings.Buildings[key]
		costLine := costStr(b.CostBase.Metal, b.CostBase.Silicon, b.CostBase.Hydrogen)
		fmt.Fprintf(indexB, "| %d | [%s](./%s.md) | %s | %d |\n",
			b.ID, key, key, costLine, b.MaxLevel)

		// Отдельная страница.
		page := &strings.Builder{}
		fmt.Fprintln(page, "---")
		fmt.Fprintf(page, "title: %s\n", key)
		fmt.Fprintln(page, "category: buildings")
		fmt.Fprintf(page, "entity_id: %s\n", key)
		fmt.Fprintln(page, "---")
		fmt.Fprintln(page)
		fmt.Fprintf(page, "# %s (id=%d)\n", key, b.ID)
		fmt.Fprintln(page)
		if b.MoonOnly {
			fmt.Fprintln(page, "> **Только на луне.**")
			fmt.Fprintln(page)
		}
		fmt.Fprintf(page, "**Базовая стоимость**: %s\n\n", costLine)
		fmt.Fprintf(page, "**Множитель стоимости** (geometric): ×%.2f за уровень.\n\n", b.CostFactor)
		fmt.Fprintf(page, "**Базовое время постройки**: %d секунд.\n\n", b.TimeBaseSeconds)
		if b.MaxLevel > 0 {
			fmt.Fprintf(page, "**Максимальный уровень**: %d.\n\n", b.MaxLevel)
		}

		// Таблица первых 10 уровней.
		fmt.Fprintln(page, "## Стоимость по уровням")
		fmt.Fprintln(page)
		fmt.Fprintln(page, "| Уровень | Металл | Кремний | Водород |")
		fmt.Fprintln(page, "|---:|---:|---:|---:|")
		for lvl := 1; lvl <= 10; lvl++ {
			factor := math.Pow(b.CostFactor, float64(lvl-1))
			m := int64(float64(b.CostBase.Metal) * factor)
			si := int64(float64(b.CostBase.Silicon) * factor)
			h := int64(float64(b.CostBase.Hydrogen) * factor)
			fmt.Fprintf(page, "| %d | %s | %s | %s |\n", lvl, fmtNum(m), fmtNum(si), fmtNum(h))
		}
		fmt.Fprintln(page)
		fmt.Fprintln(page, "*Сгенерировано из `configs/buildings.yml`.*")

		if err := writeIfChanged(filepath.Join(dir, key+".md"), page.String()); err != nil {
			return err
		}
	}
	return writeIfChanged(filepath.Join(dir, "index.md"), indexB.String())
}

// --- Ships ---

func genShips(cat *config.Catalog, outDir string) error {
	dir := filepath.Join(outDir, "ships")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	keys := sortedKeys(cat.Ships.Ships)
	indexB := &strings.Builder{}
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB, "title: Корабли")
	fmt.Fprintln(indexB, "category: ships")
	fmt.Fprintln(indexB, "order: 30")
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "# Корабли")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "| ID | Ключ | Attack | Shell | Shield | Cargo | Speed | Стоимость |")
	fmt.Fprintln(indexB, "|---:|---|---:|---:|---:|---:|---:|---|")

	for _, key := range keys {
		s := cat.Ships.Ships[key]
		costLine := costStr(s.Cost.Metal, s.Cost.Silicon, s.Cost.Hydrogen)
		fmt.Fprintf(indexB, "| %d | [%s](./%s.md) | %d | %d | %d | %d | %d | %s |\n",
			s.ID, key, key, s.Attack, s.Shell, s.Shield, s.Cargo, s.Speed, costLine)

		// Страница корабля.
		page := &strings.Builder{}
		fmt.Fprintln(page, "---")
		fmt.Fprintf(page, "title: %s\n", key)
		fmt.Fprintln(page, "category: ships")
		fmt.Fprintf(page, "entity_id: %s\n", key)
		fmt.Fprintln(page, "---")
		fmt.Fprintln(page)
		fmt.Fprintf(page, "# %s (id=%d)\n", key, s.ID)
		fmt.Fprintln(page)
		fmt.Fprintf(page, "**Стоимость**: %s\n\n", costLine)
		fmt.Fprintln(page, "## Характеристики")
		fmt.Fprintln(page)
		fmt.Fprintln(page, "| Параметр | Значение |")
		fmt.Fprintln(page, "|---|---:|")
		fmt.Fprintf(page, "| Attack | %d |\n", s.Attack)
		fmt.Fprintf(page, "| Shell | %d |\n", s.Shell)
		fmt.Fprintf(page, "| Shield | %d |\n", s.Shield)
		fmt.Fprintf(page, "| Cargo | %d |\n", s.Cargo)
		fmt.Fprintf(page, "| Speed | %d |\n", s.Speed)
		fmt.Fprintf(page, "| Fuel | %d |\n", s.Fuel)
		fmt.Fprintln(page)

		// Rapidfire граф.
		if rfOut := rapidfireFrom(cat, s.ID); len(rfOut) > 0 {
			fmt.Fprintln(page, "## Rapidfire: контрит")
			fmt.Fprintln(page)
			for _, ln := range rfOut {
				fmt.Fprintln(page, "- "+ln)
			}
			fmt.Fprintln(page)
		}
		if rfIn := rapidfireTo(cat, s.ID); len(rfIn) > 0 {
			fmt.Fprintln(page, "## Rapidfire: контрится")
			fmt.Fprintln(page)
			for _, ln := range rfIn {
				fmt.Fprintln(page, "- "+ln)
			}
			fmt.Fprintln(page)
		}
		fmt.Fprintln(page, "*Сгенерировано из `configs/ships.yml` + `construction.yml` + `rapidfire.yml`.*")

		if err := writeIfChanged(filepath.Join(dir, key+".md"), page.String()); err != nil {
			return err
		}
	}
	return writeIfChanged(filepath.Join(dir, "index.md"), indexB.String())
}

// --- Defense ---

func genDefense(cat *config.Catalog, outDir string) error {
	dir := filepath.Join(outDir, "defense")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	keys := sortedKeys(cat.Defense.Defense)
	indexB := &strings.Builder{}
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB, "title: Оборона")
	fmt.Fprintln(indexB, "category: defense")
	fmt.Fprintln(indexB, "order: 40")
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "# Оборона")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "| ID | Ключ | Attack | Shield | Shell | Стоимость |")
	fmt.Fprintln(indexB, "|---:|---|---:|---:|---:|---|")

	for _, key := range keys {
		d := cat.Defense.Defense[key]
		costLine := costStr(d.Cost.Metal, d.Cost.Silicon, d.Cost.Hydrogen)
		fmt.Fprintf(indexB, "| %d | [%s](./%s.md) | %d | %d | %d | %s |\n",
			d.ID, key, key, d.Attack, d.Shield, d.Shell, costLine)

		page := &strings.Builder{}
		fmt.Fprintln(page, "---")
		fmt.Fprintf(page, "title: %s\n", key)
		fmt.Fprintln(page, "category: defense")
		fmt.Fprintf(page, "entity_id: %s\n", key)
		fmt.Fprintln(page, "---")
		fmt.Fprintln(page)
		fmt.Fprintf(page, "# %s (id=%d)\n", key, d.ID)
		fmt.Fprintln(page)
		fmt.Fprintf(page, "**Стоимость**: %s\n\n", costLine)
		fmt.Fprintln(page, "| Параметр | Значение |")
		fmt.Fprintln(page, "|---|---:|")
		fmt.Fprintf(page, "| Attack | %d |\n", d.Attack)
		fmt.Fprintf(page, "| Shield | %d |\n", d.Shield)
		fmt.Fprintf(page, "| Shell | %d |\n", d.Shell)
		fmt.Fprintln(page)
		fmt.Fprintln(page, "*Сгенерировано из `configs/defense.yml`.*")

		if err := writeIfChanged(filepath.Join(dir, key+".md"), page.String()); err != nil {
			return err
		}
	}
	return writeIfChanged(filepath.Join(dir, "index.md"), indexB.String())
}

// --- Research ---

func genResearch(cat *config.Catalog, outDir string) error {
	dir := filepath.Join(outDir, "research")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Research идёт из Buildings (type=Research mode=2 в legacy-трактовке).
	// У нас он лежит в cat.Research — но если его нет, соберём из тех
	// construction-entries, у которых mode=2. Для MVP ограничимся Research
	// из construction-каталога.
	indexB := &strings.Builder{}
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB, "title: Исследования")
	fmt.Fprintln(indexB, "category: research")
	fmt.Fprintln(indexB, "order: 35")
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "# Исследования")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "Технологии, требующие Research Lab.")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "| ID | Ключ | Базовая стоимость | Множитель |")
	fmt.Fprintln(indexB, "|---:|---|---|---:|")

	// Стабильный порядок.
	rKeys := sortedKeys(cat.Research.Research)
	count := 0
	for _, key := range rKeys {
		b := cat.Research.Research[key]
		costLine := costStr(b.CostBase.Metal, b.CostBase.Silicon, b.CostBase.Hydrogen)
		fmt.Fprintf(indexB, "| %d | %s | %s | ×%.2f |\n", b.ID, key, costLine, b.CostFactor)
		count++
	}
	if count == 0 {
		fmt.Fprintln(indexB)
		fmt.Fprintln(indexB, "*(Каталог research не загружен — требуется проверка конфига.)*")
	}
	return writeIfChanged(filepath.Join(dir, "index.md"), indexB.String())
}

// --- helpers ---

func rapidfireFrom(cat *config.Catalog, shooterID int) []string {
	row, ok := cat.Rapidfire.Rapidfire[shooterID]
	if !ok {
		return nil
	}
	// Сортируем по значению desc, потом по id asc.
	type rfItem struct {
		target int
		value  int
	}
	items := make([]rfItem, 0, len(row))
	for t, v := range row {
		items = append(items, rfItem{t, v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].value != items[j].value {
			return items[i].value > items[j].value
		}
		return items[i].target < items[j].target
	})
	out := make([]string, 0, len(items))
	for _, it := range items {
		name := entityNameByID(cat, it.target)
		out = append(out, fmt.Sprintf("**%s** (id=%d) × %d", name, it.target, it.value))
	}
	return out
}

func rapidfireTo(cat *config.Catalog, targetID int) []string {
	type rfItem struct {
		shooter int
		value   int
	}
	var items []rfItem
	for s, row := range cat.Rapidfire.Rapidfire {
		if v, ok := row[targetID]; ok {
			items = append(items, rfItem{s, v})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].value != items[j].value {
			return items[i].value > items[j].value
		}
		return items[i].shooter < items[j].shooter
	})
	out := make([]string, 0, len(items))
	for _, it := range items {
		name := entityNameByID(cat, it.shooter)
		out = append(out, fmt.Sprintf("**%s** (id=%d) × %d", name, it.shooter, it.value))
	}
	return out
}

func entityNameByID(cat *config.Catalog, id int) string {
	for key, s := range cat.Ships.Ships {
		if s.ID == id {
			return key
		}
	}
	for key, d := range cat.Defense.Defense {
		if d.ID == id {
			return key
		}
	}
	return fmt.Sprintf("unit_%d", id)
}

func costStr(m, si, h int64) string {
	parts := []string{}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%s M", fmtNum(m)))
	}
	if si > 0 {
		parts = append(parts, fmt.Sprintf("%s Si", fmtNum(si)))
	}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%s H", fmtNum(h)))
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, " + ")
}

func fmtNum(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	// разделитель тысяч — узкий неразрывный пробел
	s := fmt.Sprintf("%d", n)
	var out strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out.WriteRune(' ')
		}
		out.WriteRune(c)
	}
	return out.String()
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// writeIfChanged пишет файл только если содержимое отличается.
// Сохраняет mtime и git-diff минимальным при запусках-noop.
func writeIfChanged(path, content string) error {
	if existing, err := os.ReadFile(path); err == nil && string(existing) == content {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
