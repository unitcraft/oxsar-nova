package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// writeOriginOverride сериализует OverrideDoc в configs/balance/origin.yaml.
//
// Сериализация ручная (не yaml.Marshal), чтобы гарантировать:
//   - стабильный порядок ключей (alphabetical для buildings/research);
//   - отсутствие научной нотации (1e+06 вместо 1000000);
//   - комментарии-метаданные сверху файла + рядом с динамическими
//     формулами (HasDynamicProd → ссылка на internal/origin/economy/).
func writeOriginOverride(path string, doc *OverrideDoc) error {
	var b strings.Builder

	fmt.Fprintf(&b, "# configs/balance/%s.yaml\n", doc.Universe)
	fmt.Fprintln(&b, "# Override-файл для вселенной origin (oxsar2-classic balance).")
	fmt.Fprintln(&b, "# Применяется поверх дефолтных configs/{buildings,units,rapidfire,...}.yml.")
	fmt.Fprintln(&b, "# Источник: импорт из docker-mysql-1 (origin-стенд) таблиц")
	fmt.Fprintln(&b, "# na_construction, na_ship_datasheet, na_rapidfire.")
	fmt.Fprintln(&b, "# Сгенерировано cmd/tools/import-legacy-balance/main.go (план 64 Ф.2).")
	fmt.Fprintf(&b, "# Дата импорта: %s\n", doc.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintln(&b, "#")
	fmt.Fprintln(&b, "# R0: дефолтный configs/buildings.yml не модифицируется. Этот файл —")
	fmt.Fprintln(&b, "# единственный источник чисел origin, отличающихся от nova-defaults.")
	fmt.Fprintln(&b, "# Для динамических формул (prod_*, cons_* с {temp}/{tech=N}) числа НЕ")
	fmt.Fprintln(&b, "# предвычисляются — реализация остаётся в Go (internal/economy/")
	fmt.Fprintln(&b, "# formulas.go и internal/origin/economy/), а коэффициенты — в Globals")
	fmt.Fprintln(&b, "# bundle (internal/balance.Globals).")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "version: %d\n", doc.Version)
	fmt.Fprintf(&b, "universe: %s\n", doc.Universe)
	fmt.Fprintln(&b)

	// Globals (если есть override).
	if len(doc.Globals) > 0 {
		fmt.Fprintln(&b, "# Глобальные коэффициенты, отличающиеся от ModernGlobals().")
		fmt.Fprintln(&b, "# Поля без override унаследуются из internal/balance.ModernGlobals().")
		fmt.Fprintln(&b, "globals:")
		keys := make([]string, 0, len(doc.Globals))
		for k := range doc.Globals {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "  %s: %v\n", k, doc.Globals[k])
		}
		fmt.Fprintln(&b)
	} else {
		fmt.Fprintln(&b, "# globals: --- secret weapon. Origin-формулы prod_* совпадают с nova")
		fmt.Fprintln(&b, "# (verify 2026-04-28 против live origin docker-mysql-1). Поэтому")
		fmt.Fprintln(&b, "# Globals по умолчанию == ModernGlobals(); ничего перекрывать не нужно.")
		fmt.Fprintln(&b)
	}

	// Buildings.
	if len(doc.Buildings) > 0 {
		fmt.Fprintln(&b, "# Здания. Override применяется поверх configs/buildings.yml.")
		fmt.Fprintln(&b, "# Поля, не указанные здесь — наследуются из nova-defaults.")
		fmt.Fprintln(&b, "buildings:")
		keys := make([]string, 0, len(doc.Buildings))
		for k := range doc.Buildings {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			writeBuildingBlock(&b, k, doc.Buildings[k])
		}
		fmt.Fprintln(&b)
	}

	// Research.
	if len(doc.Research) > 0 {
		fmt.Fprintln(&b, "# Исследования. Override применяется поверх configs/research.yml.")
		fmt.Fprintln(&b, "research:")
		keys := make([]string, 0, len(doc.Research))
		for k := range doc.Research {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			writeResearchBlock(&b, k, doc.Research[k])
		}
		fmt.Fprintln(&b)
	}

	// Ships override (Lancer, Shadow, planet shields, alien fleet — для
	// origin их числа отличаются от nova).
	if len(doc.Ships) > 0 {
		fmt.Fprintln(&b, "# Корабли. Override применяется поверх configs/ships.yml.")
		fmt.Fprintln(&b, "# Числа nova-default — другие (план 22/26 + ADR-0007/0008);")
		fmt.Fprintln(&b, "# для origin восстанавливаем legacy-числа из na_construction +")
		fmt.Fprintln(&b, "# na_ship_datasheet.")
		fmt.Fprintln(&b, "ships:")
		keys := make([]string, 0, len(doc.Ships))
		for k := range doc.Ships {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			writeShipBlock(&b, k, doc.Ships[k])
		}
		fmt.Fprintln(&b)
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeShipBlock(b *strings.Builder, key string, spec ShipOverrideOut) {
	fmt.Fprintf(b, "  %s:\n", key)
	fmt.Fprintf(b, "    # origin: id=%d name=%s\n", spec.OriginID, spec.OriginName)
	fmt.Fprintf(b, "    cost: { metal: %d, silicon: %d, hydrogen: %d }\n",
		spec.Cost.Metal, spec.Cost.Silicon, spec.Cost.Hydrogen)
	fmt.Fprintf(b, "    attack: %d\n", spec.Attack)
	fmt.Fprintf(b, "    shield: %d\n", spec.Shield)
	if spec.Shell > 0 {
		fmt.Fprintf(b, "    shell:  %d\n", spec.Shell)
	}
	if spec.Cargo > 0 {
		fmt.Fprintf(b, "    cargo:  %d\n", spec.Cargo)
	}
	if spec.Speed > 0 {
		fmt.Fprintf(b, "    speed:  %d\n", spec.Speed)
	}
	if spec.Fuel > 0 {
		fmt.Fprintf(b, "    fuel:   %d\n", spec.Fuel)
	}
	if spec.Front > 0 {
		fmt.Fprintf(b, "    front:  %d\n", spec.Front)
	}
}

func writeBuildingBlock(b *strings.Builder, key string, spec BuildingOverrideOut) {
	fmt.Fprintf(b, "  %s:\n", key)
	fmt.Fprintf(b, "    # origin: id=%d name=%s\n", spec.OriginID, spec.OriginName)
	if spec.HasBasic {
		fmt.Fprintln(b, "    cost_base:")
		fmt.Fprintf(b, "      metal:    %d\n", spec.BasicMetal)
		fmt.Fprintf(b, "      silicon:  %d\n", spec.BasicSilicon)
		fmt.Fprintf(b, "      hydrogen: %d\n", spec.BasicHydrogen)
	}
	if spec.HasCostFactor {
		fmt.Fprintf(b, "    cost_factor: %v\n", spec.CostFactor)
	}
	if spec.HasDynamicProd {
		fmt.Fprintf(b, "    # has_dynamic_production: true\n")
		fmt.Fprintf(b, "    # origin prod formula: %s\n", oneline(spec.OriginProdFormula))
		fmt.Fprintf(b, "    # реализация — internal/origin/economy/buildings.go\n")
	}
	if spec.HasDynamicCons {
		fmt.Fprintf(b, "    # has_dynamic_consumption: true\n")
		fmt.Fprintf(b, "    # origin cons formula: %s\n", oneline(spec.OriginConsFormula))
	}
}

func writeResearchBlock(b *strings.Builder, key string, spec ResearchOverrideOut) {
	fmt.Fprintf(b, "  %s:\n", key)
	fmt.Fprintf(b, "    # origin: id=%d name=%s\n", spec.OriginID, spec.OriginName)
	if spec.HasBasic {
		fmt.Fprintln(b, "    cost_base:")
		fmt.Fprintf(b, "      metal:    %d\n", spec.BasicMetal)
		fmt.Fprintf(b, "      silicon:  %d\n", spec.BasicSilicon)
		fmt.Fprintf(b, "      hydrogen: %d\n", spec.BasicHydrogen)
	}
	if spec.HasCostFactor {
		fmt.Fprintf(b, "    cost_factor: %v\n", spec.CostFactor)
	}
}

func oneline(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

// appendDefaultUnits добавляет алиен/спец-юниты в configs/units.yml.
// Идемпотентно: если ключ уже присутствует — не дублирует.
//
// План 64 R0-исключение: эти юниты доступны во всех вселенных
// (uni01/uni02/origin), потому идут в дефолтный реестр, не в origin.yaml.
func appendDefaultUnits(path string, ext *DefaultExtensions) error {
	if len(ext.UnitsAppend) == 0 && len(ext.UnitsDefenseAppend) == 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	for _, e := range ext.UnitsAppend {
		if !unitKeyAlreadyPresent(content, e.Key) {
			block := fmt.Sprintf("  - { id: %d, key: %s, name: \"%s\" }    # план 64 R0-exception\n",
				e.ID, e.Key, e.Name)
			content, err = appendToYAMLSection(content, "fleet:", block)
			if err != nil {
				return err
			}
		}
	}
	for _, e := range ext.UnitsDefenseAppend {
		if !unitKeyAlreadyPresent(content, e.Key) {
			block := fmt.Sprintf("  - { id: %d, key: %s, name: \"%s\" }    # план 64 R0-exception\n",
				e.ID, e.Key, e.Name)
			content, err = appendToYAMLSection(content, "defense:", block)
			if err != nil {
				return err
			}
		}
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// appendDefaultShips добавляет балансовые записи для алиен/спец-юнитов
// в configs/ships.yml.
func appendDefaultShips(path string, ext *DefaultExtensions) error {
	if len(ext.ShipsAppend) == 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	keys := make([]string, 0, len(ext.ShipsAppend))
	for k := range ext.ShipsAppend {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		s := ext.ShipsAppend[key]
		if shipKeyAlreadyPresent(content, key) {
			continue
		}
		block := fmt.Sprintf(`
  %s:    # план 64 R0-exception
    id:     %d
    cost:   { metal: %d, silicon: %d, hydrogen: %d }
    attack: %d
    shield: %d
    shell:  %d
    cargo:  %d
    speed:  %d
    fuel:   %d
    front:  %d
`, key, s.ID, s.Cost.Metal, s.Cost.Silicon, s.Cost.Hydrogen, s.Attack, s.Shield, s.Shell, s.Cargo, s.Speed, s.Fuel, s.Front)
		content += block
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// appendDefaultRapidfire добавляет RF-записи алиен/спец-юнитов в
// configs/rapidfire.yml in-place, сохраняя комментарии исходного файла.
//
// Стратегия — текстовый merge:
//   - читаем существующие shooter ID через YAML-парсинг (для
//     идемпотентности и проверки существующих target'ов)
//   - для каждого new shooter ID:
//     · если shooter не существует → добавляем новый блок в конец
//       (с комментарием план-64)
//     · если shooter существует → находим его блок в текстовом
//       представлении и добавляем недостающие target'ы внутрь блока
//       (комментарии других строк не трогаем)
func appendDefaultRapidfire(path string, ext *DefaultExtensions) error {
	if len(ext.RapidfireAppend) == 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	hasCRLF := strings.Contains(content, "\r\n")

	// Парсим текущее содержимое чтобы знать какие shooter'ы и target'ы
	// уже есть (план 64 R0: существующие RF-числа nova не трогаем).
	var doc rapidfireDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parse existing rapidfire.yml: %w", err)
	}

	// Сортируем shooters для детерминированного порядка добавлений.
	shooters := make([]int, 0, len(ext.RapidfireAppend))
	for s := range ext.RapidfireAppend {
		shooters = append(shooters, s)
	}
	sort.Ints(shooters)

	var newBlock strings.Builder
	for _, shooter := range shooters {
		newTargets := ext.RapidfireAppend[shooter]
		existingTargets := doc.Rapidfire[shooter]
		if existingTargets == nil {
			// Новый shooter — добавляем целиком в конец.
			fmt.Fprintf(&newBlock, "  %d:    # план 64 R0-exception\n", shooter)
			ts := make([]int, 0, len(newTargets))
			for t := range newTargets {
				ts = append(ts, t)
			}
			sort.Ints(ts)
			for _, t := range ts {
				fmt.Fprintf(&newBlock, "    %d: %d\n", t, newTargets[t])
			}
			continue
		}
		// shooter уже есть — добавляем недостающие target'ы внутрь
		// существующего блока. Делаем текстовый insert после последней
		// строки блока (перед следующим shooter'ом или EOF).
		ts := make([]int, 0)
		for t, v := range newTargets {
			if _, has := existingTargets[t]; has {
				continue // R0
			}
			_ = v
			ts = append(ts, t)
		}
		if len(ts) == 0 {
			continue
		}
		sort.Ints(ts)
		insertion := ""
		for _, t := range ts {
			insertion += fmt.Sprintf("    %d: %d    # план 64 R0-exception\n", t, newTargets[t])
		}
		newContent, err := insertIntoExistingShooter(content, shooter, insertion, hasCRLF)
		if err != nil {
			return fmt.Errorf("merge into shooter %d: %w", shooter, err)
		}
		content = newContent
	}

	if newBlock.Len() == 0 {
		// Только in-place inserts были.
		out := content
		if hasCRLF {
			// content уже содержит \r\n; insert'ы добавились с \n —
			// нормализуем.
			out = strings.ReplaceAll(strings.ReplaceAll(out, "\r\n", "\n"), "\n", "\r\n")
		}
		return os.WriteFile(path, []byte(out), 0o644)
	}

	// Добавление новых shooter'ов в конец. Гарантируем перевод строки
	// перед нашим блоком.
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n  # План 64 R0-exception: RF алиен/спец-юнитов.\n"
	content += "  # Источник: oxsar2-mysql-1 / oxsar2/sql/new-for-dm/data.sql.\n"
	content += newBlock.String()
	if hasCRLF {
		content = strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", "\n"), "\n", "\r\n")
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// rapidfireDoc — DTO для парсинга configs/rapidfire.yml.
type rapidfireDoc struct {
	Rapidfire map[int]map[int]int `yaml:"rapidfire"`
}

// insertIntoExistingShooter находит блок shooter ID в content и
// вставляет insertion перед границей блока (следующая строка с тем же
// или меньшим отступом — следующий shooter или EOF).
//
// Поддерживает CRLF и LF.
func insertIntoExistingShooter(content string, shooter int, insertion string, hasCRLF bool) (string, error) {
	headerLF := fmt.Sprintf("\n  %d:", shooter)
	normalized := content
	if hasCRLF {
		normalized = strings.ReplaceAll(content, "\r\n", "\n")
	}
	idx := strings.Index(normalized, headerLF)
	if idx < 0 {
		return content, fmt.Errorf("shooter %d header not found", shooter)
	}
	// Курсор на начало строки после "  N:".
	cursor := idx + len(headerLF)
	// Пропускаем до конца строки.
	if nl := strings.IndexByte(normalized[cursor:], '\n'); nl >= 0 {
		cursor += nl + 1
	} else {
		cursor = len(normalized)
	}
	// Ищем следующую строку с отступом ≤ 2 (новый shooter или
	// top-level). Курсор уже на первой строке тела блока.
	rest := normalized[cursor:]
	lines := strings.SplitAfter(rest, "\n")
	insertOffsetN := len(rest)
	for i, line := range lines {
		trimmed := strings.TrimRight(line, "\n")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		ind := 0
		for ind < len(line) && line[ind] == ' ' {
			ind++
		}
		if ind <= 2 {
			off := 0
			for j := 0; j < i; j++ {
				off += len(lines[j])
			}
			insertOffsetN = off
			break
		}
	}
	nInsertPos := cursor + insertOffsetN
	origPos := nInsertPos
	if hasCRLF {
		origPos = mapNormalizedToOriginal(content, nInsertPos)
		insertion = strings.ReplaceAll(insertion, "\n", "\r\n")
	}
	return content[:origPos] + insertion + content[origPos:], nil
}

// --- helpers ---

func unitKeyAlreadyPresent(content, key string) bool {
	// Грубая проверка: ищем "key: <key>" или "key: <key>," в YAML.
	return strings.Contains(content, "key: "+key+",") || strings.Contains(content, "key: "+key+" ") || strings.Contains(content, "key: "+key+"\n") || strings.Contains(content, "key: "+key+"}") || strings.Contains(content, "key: "+key+" }")
}

func shipKeyAlreadyPresent(content, key string) bool {
	return strings.Contains(content, "  "+key+":\n") || strings.Contains(content, "  "+key+":    ")
}

func rapidfireShooterAlreadyPresent(content string, shooter int) bool {
	return strings.Contains(content, fmt.Sprintf("\n  %d:\n", shooter)) ||
		strings.Contains(content, fmt.Sprintf("\n  %d:    ", shooter))
}

// appendToYAMLSection вставляет block в конец YAML-секции,
// отмеченной headerLine (например "fleet:" на топ-уровне или "  fleet:"
// с отступом). Поддерживает CRLF и LF line endings (existing nova-
// конфиги — CRLF на Windows). Возвращает ошибку если секция не найдена.
//
// Логика: ищем строку headerLine, потом ищем следующую строку с
// меньшим или равным отступом (= следующая секция или EOF). block
// вставляется ПЕРЕД границей.
func appendToYAMLSection(content, headerLine, block string) (string, error) {
	// Нормализуем поиск: уберём \r, чтобы не зависеть от line endings.
	// Сами content и block оставляем как есть — индексы пересчитываем
	// через map позиций.
	normalized := strings.ReplaceAll(content, "\r\n", "\n")

	headerWithBoundary := "\n" + headerLine + "\n"
	nIdx := strings.Index(normalized, headerWithBoundary)
	if nIdx < 0 && strings.HasPrefix(normalized, headerLine+"\n") {
		// Секция в самом начале файла.
		nIdx = -1 // обработать как position 0
	}
	if nIdx < 0 && !strings.HasPrefix(normalized, headerLine+"\n") {
		return content, fmt.Errorf("section %q not found in YAML", headerLine)
	}

	// Определяем отступ заголовка.
	headerIndent := 0
	for headerIndent < len(headerLine) && headerLine[headerIndent] == ' ' {
		headerIndent++
	}

	// Курсор после заголовка в нормализованной строке.
	var nCursor int
	if nIdx < 0 {
		nCursor = len(headerLine) + 1 // после "<header>:\n"
	} else {
		nCursor = nIdx + len(headerWithBoundary)
	}
	rest := normalized[nCursor:]
	lines := strings.SplitAfter(rest, "\n")
	insertOffsetN := len(rest)
	for i, line := range lines {
		trimmed := strings.TrimRight(line, "\n")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		ind := 0
		for ind < len(line) && line[ind] == ' ' {
			ind++
		}
		if ind <= headerIndent {
			off := 0
			for j := 0; j < i; j++ {
				off += len(lines[j])
			}
			insertOffsetN = off
			break
		}
	}
	nInsert := nCursor + insertOffsetN

	// Конвертируем позицию из normalized → original (учитываем
	// число '\r' до этой позиции).
	origPos := mapNormalizedToOriginal(content, nInsert)

	// block уже в LF; если в исходнике CRLF — конвертируем block тоже.
	if strings.Contains(content, "\r\n") {
		block = strings.ReplaceAll(block, "\n", "\r\n")
	}
	return content[:origPos] + block + content[origPos:], nil
}

// mapNormalizedToOriginal — позиция nPos указывает на normalized
// (CRLF→LF) строку; возвращает соответствующий offset в исходной
// content. Идём вперёд по content, пока не пройдём nPos LF-байт.
func mapNormalizedToOriginal(content string, nPos int) int {
	pos := 0
	npos := 0
	for pos < len(content) && npos < nPos {
		if pos+1 < len(content) && content[pos] == '\r' && content[pos+1] == '\n' {
			pos += 2
			npos++ // в normalized это один '\n'
			continue
		}
		pos++
		npos++
	}
	return pos
}
