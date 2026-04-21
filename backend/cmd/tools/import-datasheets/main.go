// Command import-datasheets — конвертер sql/table_dump/*.sql в
// configs/*.yml.
//
// Это реализация M0.1 из §16 ТЗ. Источник балансов — не наш «на глаз»
// YAML, а реальные данные прода oxsar2-srv-01 (2011 г.), зафиксированные
// в table_dump.
//
// Использование:
//     import-datasheets \
//         --input=d:/Sources/oxsar2/sql/table_dump \
//         --output=d:/Sources/oxsar-nova/configs
//
// Скрипт идемпотентен: повторный прогон без изменений во входе даёт
// тот же выход. Поэтому выход можно коммитить в git.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type converter func(inputDir, outputDir string) error

// converters — реестр «входной SQL-файл → конвертер». Порядок не важен.
var converters = map[string]converter{
	"na_construction.sql":      convertConstruction,
	"na_ship_datasheet.sql":    convertShips,
	"na_requirements.sql":      convertRequirements,
	"na_artefact_datasheet.sql": convertArtefacts,
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "import-datasheets:", err)
		os.Exit(1)
	}
}

func run() error {
	input := flag.String("input", "", "папка с sql/table_dump/*.sql (обязательно)")
	output := flag.String("output", "", "папка с configs/ для записи *.yml (обязательно)")
	dryRun := flag.Bool("dry-run", false, "не писать файлы, только распарсить и показать статистику")
	flag.Parse()

	if *input == "" || *output == "" {
		return fmt.Errorf("нужны --input и --output")
	}

	for file, conv := range converters {
		path := filepath.Join(*input, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("skip %s (not found)\n", file)
			continue
		}
		if *dryRun {
			fmt.Printf("would convert %s\n", file)
			continue
		}
		if err := conv(*input, *output); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		fmt.Printf("ok %s\n", file)
	}
	return nil
}

func readInputSQL(inputDir, file string) (string, error) {
	data, err := os.ReadFile(filepath.Join(inputDir, file))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
