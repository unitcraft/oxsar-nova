// Command i18n-rename — массовое переименование ключей и групп в
// configs/i18n/*.yml из SCREAMING_SNAKE_CASE в lowerCamelCase +
// конвертация %s/%d плейсхолдеров в {{name}} (с именами из glossary).
//
// Использование:
//
//	go run ./cmd/tools/i18n-rename \
//	  --dir=../../../configs/i18n \
//	  --glossary=../../../docs/plans/33-i18n-placeholder-glossary.yml \
//	  --map-out=../../../configs/i18n/i18n-rename-map.json
//
// Алгоритм переименования ключей:
//   - split по "_"
//   - первый сегмент → lower
//   - остальные → Title
//   - если первый символ результата — цифра → panic (нужен ручной фикс)
//
// Алгоритм переименования групп: явная таблица (hardcoded), т.к. групп
// только 23 и автоматика даст неожиданные результаты (напр. "Main" → "main").
//
// Плейсхолдеры: если glossary не задан, скрипт конвертирует %s → {{arg1}},
// %d → {{num1}} с числовым суффиксом по порядку. Правильные имена нужно
// проставить вручную через glossary (YAML: "ASSAULT_TIME.%s.1": "time").
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "i18n-rename:", err)
		os.Exit(1)
	}
}

// groupRenameTable — явная таблица переименования групп.
var groupRenameTable = map[string]string{
	"Achievements":   "achievements",
	"Administrator":  "administrator",
	"Alliance":       "alliance",
	"ArtefactInfo":   "artefactInfo",
	"AssaultReport":  "assaultReport",
	"AutoMessages":   "autoMessages",
	"Buddylist":      "buddylist",
	"EspionageReport": "espionageReport",
	"Galaxy":         "galaxy",
	"Main":           "main",
	"Message":        "message",
	"Payment":        "payment",
	"Prefs":          "prefs",
	"Registration":   "registration",
	"Resource":       "resource",
	"Statistics":     "statistics",
	"UnitInfo":       "unitInfo",
	"UserAgreement":  "userAgreement",
	// Уже в нужном формате — оставляем.
	"buildings": "buildings",
	"error":     "error",
	"global":    "global",
	"info":      "info",
	"mission":   "mission",
}

// RenameMap хранит маппинг "OldGroup.OLD_KEY" → "newGroup.newKey".
type RenameMap map[string]string

var rePrintf = regexp.MustCompile(`%[sd]`)

func run() error {
	dir := flag.String("dir", "configs/i18n", "папка с *.yml локалями")
	glossaryPath := flag.String("glossary", "", "путь к glossary YAML (опционально)")
	mapOut := flag.String("map-out", "configs/i18n/i18n-rename-map.json", "выходной JSON rename-map")
	flag.Parse()

	// Загружаем glossary плейсхолдеров, если задан.
	glossary := map[string]string{}
	if *glossaryPath != "" {
		data, err := os.ReadFile(*glossaryPath)
		if err != nil {
			return fmt.Errorf("read glossary: %w", err)
		}
		if err := yaml.Unmarshal(data, &glossary); err != nil {
			return fmt.Errorf("parse glossary: %w", err)
		}
	}

	entries, err := os.ReadDir(*dir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	renameMap := RenameMap{}
	var firstLocale map[string]map[string]string

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yml") {
			continue
		}
		// Пропускаем уже сгенерированный rename-map если вдруг лежит рядом.
		if e.Name() == "i18n-rename-map.json" {
			continue
		}

		path := filepath.Join(*dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", e.Name(), err)
		}

		var raw map[string]map[string]string
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse %s: %w", e.Name(), err)
		}

		converted, rm, err := convertLocale(raw, glossary)
		if err != nil {
			return fmt.Errorf("convert %s: %w", e.Name(), err)
		}

		// Строим общий rename-map только один раз (все локали имеют одинаковые ключи).
		if firstLocale == nil {
			firstLocale = raw
			renameMap = rm
		}

		if err := writeLocale(path, converted); err != nil {
			return fmt.Errorf("write %s: %w", e.Name(), err)
		}
		fmt.Printf("ok %s\n", e.Name())
	}

	// Сохраняем rename-map.
	mapData, err := json.MarshalIndent(renameMap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rename-map: %w", err)
	}
	if err := os.WriteFile(*mapOut, mapData, 0o644); err != nil {
		return fmt.Errorf("write rename-map: %w", err)
	}
	fmt.Printf("rename-map: %s (%d entries)\n", *mapOut, len(renameMap))
	return nil
}

// convertLocale переименовывает группы и ключи, конвертирует плейсхолдеры.
func convertLocale(raw map[string]map[string]string, glossary map[string]string) (map[string]map[string]string, RenameMap, error) {
	result := map[string]map[string]string{}
	rm := RenameMap{}

	for oldGroup, keys := range raw {
		newGroup, ok := groupRenameTable[oldGroup]
		if !ok {
			return nil, nil, fmt.Errorf("группа %q не найдена в groupRenameTable — добавьте её", oldGroup)
		}

		if _, exists := result[newGroup]; exists {
			return nil, nil, fmt.Errorf("конфликт: группа %q уже существует после переименования %q", newGroup, oldGroup)
		}
		result[newGroup] = map[string]string{}

		for oldKey, val := range keys {
			newKey, err := toCamelCase(oldKey)
			if err != nil {
				return nil, nil, fmt.Errorf("группа %s ключ %s: %w", oldGroup, oldKey, err)
			}

			// Проверка дубликатов.
			if _, exists := result[newGroup][newKey]; exists {
				return nil, nil, fmt.Errorf("дубликат после rename: %s.%s (из %s.%s)", newGroup, newKey, oldGroup, oldKey)
			}

			// Конвертируем плейсхолдеры %s/%d → {{name}}.
			newVal, err := convertPlaceholders(val, oldGroup+"."+oldKey, glossary)
			if err != nil {
				return nil, nil, fmt.Errorf("placeholder %s.%s: %w", oldGroup, oldKey, err)
			}

			result[newGroup][newKey] = newVal
			rm[oldGroup+"."+oldKey] = newGroup+"."+newKey
		}
	}

	return result, rm, nil
}

// toCamelCase конвертирует SCREAMING_SNAKE_CASE в lowerCamelCase.
func toCamelCase(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	// Уже lowerCamelCase или lower — оставляем.
	if !strings.Contains(s, "_") && s[0] == strings.ToLower(s)[0] {
		return s, nil
	}

	parts := strings.Split(s, "_")
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(p))
		} else {
			// Capitalize first rune.
			runes := []rune(p)
			b.WriteString(strings.ToUpper(string(runes[0])) + strings.ToLower(string(runes[1:])))
		}
	}

	result := b.String()
	if result == "" {
		return "", fmt.Errorf("пустой результат для %q", s)
	}
	// Проверяем, что не начинается с цифры.
	if result[0] >= '0' && result[0] <= '9' {
		return "", fmt.Errorf("ключ %q → %q начинается с цифры, нужен ручной фикс", s, result)
	}
	return result, nil
}

// convertPlaceholders заменяет %s/%d на {{name}} используя glossary.
// Glossary-ключ: "GROUP.KEY.N" (1-based) → имя параметра.
// Если glossary не задан или ключ не найден — использует argN/numN.
func convertPlaceholders(val, fullKey string, glossary map[string]string) (string, error) {
	matches := rePrintf.FindAllStringIndex(val, -1)
	if len(matches) == 0 {
		return val, nil
	}

	counter := 0
	result := rePrintf.ReplaceAllStringFunc(val, func(m string) string {
		counter++
		glossaryKey := fmt.Sprintf("%s.%d", fullKey, counter)
		if name, ok := glossary[glossaryKey]; ok {
			return "{{" + name + "}}"
		}
		// Fallback: arg1, arg2, … для %s; num1, num2, … для %d.
		if m == "%d" {
			return fmt.Sprintf("{{num%d}}", counter)
		}
		return fmt.Sprintf("{{arg%d}}", counter)
	})

	return result, nil
}

func writeLocale(path string, data map[string]map[string]string) error {
	// Сортированный вывод для стабильного git-diff.
	groupKeys := make([]string, 0, len(data))
	for k := range data {
		groupKeys = append(groupKeys, k)
	}
	sort.Strings(groupKeys)

	root := &yaml.Node{Kind: yaml.MappingNode}
	for _, g := range groupKeys {
		phraseKeys := make([]string, 0, len(data[g]))
		for k := range data[g] {
			phraseKeys = append(phraseKeys, k)
		}
		sort.Strings(phraseKeys)

		groupNode := &yaml.Node{Kind: yaml.MappingNode}
		for _, k := range phraseKeys {
			val := data[g][k]
			groupNode.Content = append(groupNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: k},
				&yaml.Node{Kind: yaml.ScalarNode, Value: val, Style: scalarStyleFor(val)},
			)
		}
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: g},
			groupNode,
		)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return err
	}
	return enc.Close()
}

func scalarStyleFor(s string) yaml.Style {
	if s == "" {
		return yaml.DoubleQuotedStyle
	}
	if strings.ContainsAny(s, ":#'\"<>{}\n") {
		return yaml.DoubleQuotedStyle
	}
	return 0
}
