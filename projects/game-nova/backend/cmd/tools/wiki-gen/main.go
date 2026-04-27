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

	"oxsar/game-nova/internal/config"
	"gopkg.in/yaml.v3"
)

// wikiDescriptions — отдельный конфиг с описаниями. Не смешиваем с
// catalog: описания — это контент wiki, не баланс.
type wikiDescriptions struct {
	Descriptions map[string]struct {
		Short string `yaml:"short"`
		Long  string `yaml:"long"`
	} `yaml:"descriptions"`
}

var descCatalog wikiDescriptions

func loadDescriptions(configsDir string) error {
	path := filepath.Join(configsDir, "wiki-descriptions.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read wiki-descriptions: %w", err)
	}
	if err := yaml.Unmarshal(data, &descCatalog); err != nil {
		return fmt.Errorf("parse wiki-descriptions: %w", err)
	}
	return nil
}

// writeDescription добавляет на страницу секцию «Описание» из
// configs/wiki-descriptions.yml. Если описания нет — секцию не пишем.
func writeDescription(page *strings.Builder, key string) {
	d, ok := descCatalog.Descriptions[key]
	if !ok || d.Long == "" {
		return
	}
	fmt.Fprintln(page, "## Описание")
	fmt.Fprintln(page)
	fmt.Fprintln(page, strings.TrimRight(d.Long, "\n"))
	fmt.Fprintln(page)
}

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
	if err := loadDescriptions(*configsDir); err != nil {
		return err
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
	fmt.Fprintln(indexB, "Полный список построек. Каждое здание занимает одно поле планеты. Стоимость следующего уровня растёт по формуле `базовая × множитель^(уровень−1)`.")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "| Здание | Стоимость 1-го уровня | Макс. уровень |")
	fmt.Fprintln(indexB, "|---|---|---:|")

	for _, key := range keys {
		b := cat.Buildings.Buildings[key]
		costLine := costStr(b.CostBase.Metal, b.CostBase.Silicon, b.CostBase.Hydrogen)
		fmt.Fprintf(indexB, "| [[unit:%d]] | %s | %d |\n",
			b.ID, costLine, b.MaxLevel)

		// Отдельная страница.
		page := &strings.Builder{}
		fmt.Fprintln(page, "---")
		fmt.Fprintf(page, "title: %s\n", key)
		fmt.Fprintln(page, "category: buildings")
		fmt.Fprintf(page, "entity_key: %s\n", key)
		fmt.Fprintf(page, "unit_id: %d\n", b.ID)
		fmt.Fprintln(page, "---")
		fmt.Fprintln(page)
		// Имя и иконку показывает frontend по unit_id; собственный заголовок
		// в md не пишем — он бы дублировался.
		if b.MoonOnly {
			fmt.Fprintln(page, "> **Только на луне.**")
			fmt.Fprintln(page)
		}
		writeDescription(page, key)
		writeRequirements(page, cat, key)
		fmt.Fprintln(page, "## Стоимость и время")
		fmt.Fprintln(page)
		fmt.Fprintf(page, "- **Стоимость 1-го уровня**: %s.\n", costLine)
		fmt.Fprintf(page, "- **Каждый следующий уровень дороже в %.2f раза** по всем ресурсам.\n", b.CostFactor)
		fmt.Fprintf(page, "- **Базовое время постройки**: %s.\n", durationStr(b.TimeBaseSeconds))
		if b.MaxLevel > 0 {
			fmt.Fprintf(page, "- **Максимальный уровень**: %d.\n", b.MaxLevel)
		}
		fmt.Fprintln(page)

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
		fmt.Fprintln(page, "*Числа берутся из `configs/buildings.yml`.*")

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
	fmt.Fprintln(indexB, "Корабли строятся на [[unit:8]] и используются для атак, обороны, перевозки ресурсов и колонизации. Скорость в бою и в полёте усиливается двигательными исследованиями.")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "| Корабль | Атака | Корпус | Щит | Грузоподъёмность | Скорость | Стоимость |")
	fmt.Fprintln(indexB, "|---|---:|---:|---:|---:|---:|---|")

	for _, key := range keys {
		s := cat.Ships.Ships[key]
		costLine := costStr(s.Cost.Metal, s.Cost.Silicon, s.Cost.Hydrogen)
		fmt.Fprintf(indexB, "| [[unit:%d]] | %s | %s | %s | %s | %s | %s |\n",
			s.ID, fmtNum(int64(s.Attack)), fmtNum(int64(s.Shell)),
			fmtNum(int64(s.Shield)), fmtNum(int64(s.Cargo)),
			fmtNum(int64(s.Speed)), costLine)

		// Страница корабля.
		page := &strings.Builder{}
		fmt.Fprintln(page, "---")
		fmt.Fprintf(page, "title: %s\n", key)
		fmt.Fprintln(page, "category: ships")
		fmt.Fprintf(page, "entity_key: %s\n", key)
		fmt.Fprintf(page, "unit_id: %d\n", s.ID)
		fmt.Fprintln(page, "---")
		fmt.Fprintln(page)
		writeDescription(page, key)
		writeRequirements(page, cat, key)
		fmt.Fprintln(page, "## Характеристики")
		fmt.Fprintln(page)
		fmt.Fprintf(page, "**Стоимость постройки**: %s.\n\n", costLine)
		fmt.Fprintln(page, "| Параметр | Значение |")
		fmt.Fprintln(page, "|---|---:|")
		fmt.Fprintf(page, "| Атака | %s |\n", fmtNum(int64(s.Attack)))
		fmt.Fprintf(page, "| Корпус | %s |\n", fmtNum(int64(s.Shell)))
		fmt.Fprintf(page, "| Щит | %s |\n", fmtNum(int64(s.Shield)))
		fmt.Fprintf(page, "| Грузоподъёмность | %s |\n", fmtNum(int64(s.Cargo)))
		fmt.Fprintf(page, "| Скорость | %s |\n", fmtNum(int64(s.Speed)))
		fmt.Fprintf(page, "| Расход топлива | %s |\n", fmtNum(int64(s.Fuel)))
		fmt.Fprintln(page)

		// Rapidfire граф.
		if rfOut := rapidfireFrom(cat, s.ID); len(rfOut) > 0 {
			fmt.Fprintln(page, "## Скорострельность по целям")
			fmt.Fprintln(page)
			fmt.Fprintln(page, "Каждый указанный юнит этот корабль может обстрелять несколько раз за один раунд боя:")
			fmt.Fprintln(page)
			for _, ln := range rfOut {
				fmt.Fprintln(page, "- "+ln)
			}
			fmt.Fprintln(page)
		}
		if rfIn := rapidfireTo(cat, s.ID); len(rfIn) > 0 {
			fmt.Fprintln(page, "## Уязвим к скорострельности")
			fmt.Fprintln(page)
			fmt.Fprintln(page, "Эти юниты могут обстреливать данный корабль несколько раз за раунд:")
			fmt.Fprintln(page)
			for _, ln := range rfIn {
				fmt.Fprintln(page, "- "+ln)
			}
			fmt.Fprintln(page)
		}
		fmt.Fprintln(page, "*Числа берутся из `configs/ships.yml`, `construction.yml`, `rapidfire.yml`.*")

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
	fmt.Fprintln(indexB, "Стационарные сооружения, защищающие планету от атак. В отличие от флота, обломки от уничтоженной обороны после боя восстанавливаются на 70 %.")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "| Сооружение | Атака | Щит | Корпус | Стоимость |")
	fmt.Fprintln(indexB, "|---|---:|---:|---:|---|")

	for _, key := range keys {
		d := cat.Defense.Defense[key]
		costLine := costStr(d.Cost.Metal, d.Cost.Silicon, d.Cost.Hydrogen)
		fmt.Fprintf(indexB, "| [[unit:%d]] | %s | %s | %s | %s |\n",
			d.ID, fmtNum(int64(d.Attack)), fmtNum(int64(d.Shield)),
			fmtNum(int64(d.Shell)), costLine)

		page := &strings.Builder{}
		fmt.Fprintln(page, "---")
		fmt.Fprintf(page, "title: %s\n", key)
		fmt.Fprintln(page, "category: defense")
		fmt.Fprintf(page, "entity_key: %s\n", key)
		fmt.Fprintf(page, "unit_id: %d\n", d.ID)
		fmt.Fprintln(page, "---")
		fmt.Fprintln(page)
		writeDescription(page, key)
		writeRequirements(page, cat, key)
		fmt.Fprintln(page, "## Характеристики")
		fmt.Fprintln(page)
		fmt.Fprintf(page, "**Стоимость постройки**: %s.\n\n", costLine)
		fmt.Fprintln(page, "| Параметр | Значение |")
		fmt.Fprintln(page, "|---|---:|")
		fmt.Fprintf(page, "| Атака | %s |\n", fmtNum(int64(d.Attack)))
		fmt.Fprintf(page, "| Щит | %s |\n", fmtNum(int64(d.Shield)))
		fmt.Fprintf(page, "| Корпус | %s |\n", fmtNum(int64(d.Shell)))
		fmt.Fprintln(page)

		fmt.Fprintln(page, "*Числа берутся из `configs/defense.yml`.*")

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
	indexB := &strings.Builder{}
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB, "title: Исследования")
	fmt.Fprintln(indexB, "category: research")
	fmt.Fprintln(indexB, "order: 35")
	fmt.Fprintln(indexB, "---")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "# Исследования")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "Технологии открывают новые корабли, исследования и оборону. Все исследования проводятся в [[unit:12]]; чем выше её уровень, тем быстрее завершается работа.")
	fmt.Fprintln(indexB)
	fmt.Fprintln(indexB, "| Технология | Стоимость 1-го уровня | Каждый уровень дороже |")
	fmt.Fprintln(indexB, "|---|---|---:|")

	rKeys := sortedKeys(cat.Research.Research)
	for _, key := range rKeys {
		b := cat.Research.Research[key]
		costLine := costStr(b.CostBase.Metal, b.CostBase.Silicon, b.CostBase.Hydrogen)
		fmt.Fprintf(indexB, "| [[unit:%d]] | %s | в %.2f раза |\n", b.ID, costLine, b.CostFactor)

		// Отдельная страница на исследование.
		page := &strings.Builder{}
		fmt.Fprintln(page, "---")
		fmt.Fprintf(page, "title: %s\n", key)
		fmt.Fprintln(page, "category: research")
		fmt.Fprintf(page, "entity_key: %s\n", key)
		fmt.Fprintf(page, "unit_id: %d\n", b.ID)
		fmt.Fprintln(page, "---")
		fmt.Fprintln(page)
		writeDescription(page, key)
		writeRequirements(page, cat, key)
		fmt.Fprintln(page, "## Стоимость и время")
		fmt.Fprintln(page)
		fmt.Fprintf(page, "- **Стоимость 1-го уровня**: %s.\n", costLine)
		fmt.Fprintf(page, "- **Каждый следующий уровень дороже в %.2f раза** по всем ресурсам.\n\n", b.CostFactor)

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
		fmt.Fprintln(page, "*Числа берутся из `configs/research.yml` и `configs/construction.yml`.*")

		if err := writeIfChanged(filepath.Join(dir, key+".md"), page.String()); err != nil {
			return err
		}
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
	_ = cat // имя цели подставит frontend через nameOf(id)
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, fmt.Sprintf("[[unit:%d]] × %d", it.target, it.value))
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
	_ = cat // имя shooter подставит frontend через nameOf(id)
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, fmt.Sprintf("[[unit:%d]] × %d", it.shooter, it.value))
	}
	return out
}

// idByKey ищет числовой id юнита по ключу — нужно для перекрёстных ссылок
// `[[unit:N]]` на здания и исследования из requirements.yml.
func idByKey(cat *config.Catalog, key string) int {
	if b, ok := cat.Buildings.Buildings[key]; ok {
		return b.ID
	}
	if r, ok := cat.Research.Research[key]; ok {
		return r.ID
	}
	if s, ok := cat.Ships.Ships[key]; ok {
		return s.ID
	}
	if d, ok := cat.Defense.Defense[key]; ok {
		return d.ID
	}
	return 0
}

// writeRequirements добавляет на страницу секцию «Требования» — список
// зданий и технологий, без которых юнит/исследование недоступны. Каждое
// требование — кликабельная ссылка `[[unit:N]]` на свою страницу вики.
func writeRequirements(page *strings.Builder, cat *config.Catalog, key string) {
	reqs := cat.Requirements.Requirements[key]
	if len(reqs) == 0 {
		return
	}
	// Стабильный порядок: сначала здания, потом исследования; внутри —
	// по уровню убыванием, при равенстве по ключу.
	sorted := make([]config.Requirement, len(reqs))
	copy(sorted, reqs)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Kind != sorted[j].Kind {
			return sorted[i].Kind == "building"
		}
		if sorted[i].Level != sorted[j].Level {
			return sorted[i].Level > sorted[j].Level
		}
		return sorted[i].Key < sorted[j].Key
	})
	fmt.Fprintln(page, "## Требования")
	fmt.Fprintln(page)
	fmt.Fprintln(page, "Чтобы открыть постройку, нужно достроить или изучить:")
	fmt.Fprintln(page)
	for _, r := range sorted {
		id := idByKey(cat, r.Key)
		if id == 0 {
			fmt.Fprintf(page, "- %s — уровень %d\n", r.Key, r.Level)
			continue
		}
		fmt.Fprintf(page, "- [[unit:%d]] — уровень %d\n", id, r.Level)
	}
	fmt.Fprintln(page)
}

// durationStr форматирует секунды в человекочитаемый вид: «1 ч 30 мин».
func durationStr(seconds int) string {
	if seconds <= 0 {
		return "мгновенно"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	parts := []string{}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%d ч", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%d мин", m))
	}
	if s > 0 && h == 0 {
		parts = append(parts, fmt.Sprintf("%d сек", s))
	}
	if len(parts) == 0 {
		return "0 сек"
	}
	return strings.Join(parts, " ")
}

func costStr(m, si, h int64) string {
	parts := []string{}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%s металла", fmtNum(m)))
	}
	if si > 0 {
		parts = append(parts, fmt.Sprintf("%s кремния", fmtNum(si)))
	}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%s водорода", fmtNum(h)))
	}
	if len(parts) == 0 {
		return "бесплатно"
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
