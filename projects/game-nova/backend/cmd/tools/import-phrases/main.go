// Command import-phrases — конвертер
//   d:\Sources\oxsar2\sql\table_dump\na_phrases.sql
// в
//   configs/i18n/ru.yml
// с заготовкой
//   configs/i18n/en.yml
//
// Реализация M0.2 (см. §10.3 ТЗ). Берём na_phrasesgroups как индекс
// имён групп, na_phrases — как строки с переводами. languageid=1 —
// русский (единственный в legacy).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"oxsar/game-nova/pkg/sqldump"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "import-phrases:", err)
		os.Exit(1)
	}
}

func run() error {
	input := flag.String("input", "", "путь к na_phrases.sql (обязательно)")
	output := flag.String("output", "", "папка configs/i18n/ для записи *.yml")
	flag.Parse()
	if *input == "" || *output == "" {
		return fmt.Errorf("нужны --input и --output")
	}

	data, err := os.ReadFile(*input)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	src := string(data)

	// Группы.
	groupsData, err := sqldump.ParseInserts(src, "na_phrasesgroups")
	if err != nil {
		return fmt.Errorf("parse groups: %w", err)
	}
	gcol := sqldump.IndexColumns(groupsData.Columns)
	groups := map[int64]string{}
	for _, row := range groupsData.Rows {
		var id int64
		if err := sqldump.AssignInt(gcol, row, "phrasegroupid", &id); err != nil {
			return err
		}
		title, err := sqldump.GetStr(gcol, row, "title")
		if err != nil {
			return err
		}
		groups[id] = title
	}
	if len(groups) == 0 {
		return fmt.Errorf("na_phrasesgroups: no rows parsed")
	}

	// Фразы.
	phrasesData, err := sqldump.ParseInserts(src, "na_phrases")
	if err != nil {
		return fmt.Errorf("parse phrases: %w", err)
	}
	pcol := sqldump.IndexColumns(phrasesData.Columns)

	// i18n[groupTitle][phraseKey] = content
	i18n := map[string]map[string]string{}
	skipped := 0
	for _, row := range phrasesData.Rows {
		var lang, groupID int64
		if err := sqldump.AssignInt(pcol, row, "languageid", &lang); err != nil {
			return err
		}
		if lang != 1 {
			// В legacy другие языки не заведены; защита на будущее.
			skipped++
			continue
		}
		if err := sqldump.AssignInt(pcol, row, "phrasegroupid", &groupID); err != nil {
			return err
		}
		key, err := sqldump.GetStr(pcol, row, "title")
		if err != nil {
			return err
		}
		content, err := sqldump.GetStr(pcol, row, "content")
		if err != nil {
			return err
		}
		group, ok := groups[groupID]
		if !ok {
			group = fmt.Sprintf("group_%d", groupID)
		}
		if _, exists := i18n[group]; !exists {
			i18n[group] = map[string]string{}
		}
		i18n[group][key] = content
	}

	if err := writeLocale(filepath.Join(*output, "ru.yml"), i18n, true); err != nil {
		return fmt.Errorf("write ru.yml: %w", err)
	}
	// en.yml — заготовка с пустыми значениями, чтобы структуру
	// переводчик увидел сразу. Если файл уже есть, НЕ перезаписываем —
	// чтобы не потерять переводы.
	enPath := filepath.Join(*output, "en.yml")
	if _, err := os.Stat(enPath); os.IsNotExist(err) {
		if err := writeLocale(enPath, i18n, false); err != nil {
			return fmt.Errorf("write en.yml: %w", err)
		}
	} else {
		fmt.Println("en.yml already exists, leaving untouched")
	}

	fmt.Printf("ok ru.yml: %d groups, %d phrases (skipped non-ru: %d)\n",
		len(i18n), countAll(i18n), skipped)
	return nil
}

func countAll(m map[string]map[string]string) int {
	n := 0
	for _, g := range m {
		n += len(g)
	}
	return n
}

// writeLocale сериализует с отсортированными группами и ключами
// (стабильный git-diff). fillValues=false — пишет пустые строки
// (заготовка en.yml).
func writeLocale(path string, i18n map[string]map[string]string, fillValues bool) error {
	groupKeys := make([]string, 0, len(i18n))
	for k := range i18n {
		groupKeys = append(groupKeys, k)
	}
	sort.Strings(groupKeys)

	root := &yaml.Node{Kind: yaml.MappingNode}
	for _, g := range groupKeys {
		phraseKeys := make([]string, 0, len(i18n[g]))
		for k := range i18n[g] {
			phraseKeys = append(phraseKeys, k)
		}
		sort.Strings(phraseKeys)

		groupNode := &yaml.Node{Kind: yaml.MappingNode}
		for _, k := range phraseKeys {
			val := ""
			if fillValues {
				val = i18n[g][k]
			}
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

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
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

// scalarStyleFor выбирает стиль кавычек, чтобы yaml.v3 не ломал
// русские строки с двоеточиями/спецсимволами. Для плейсхолдеров
// типа %s и HTML <b> лучше всегда двойные кавычки.
func scalarStyleFor(s string) yaml.Style {
	if s == "" {
		return yaml.DoubleQuotedStyle
	}
	if strings.ContainsAny(s, ":#'\"<>%\n") {
		return yaml.DoubleQuotedStyle
	}
	return 0
}
