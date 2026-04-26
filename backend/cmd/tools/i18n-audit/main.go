// Command i18n-audit сканирует frontend/src и backend на string-литералы
// с кириллицей и строит отчёт: файл:строка → литерал → предложенный
// i18n-ключ → найден ли в ru.yml.
//
// Исключает:
//   - строки-комментарии (// ... и /* ... */)
//   - test-файлы (*_test.go, *.test.ts, *.spec.tsx)
//   - slog.*/log.* вызовы (internal logging)
//   - fmt.Errorf / errors.New (internal error messages)
//   - panic(
//   - console.error / console.warn
//
// Использование:
//
//	go run ./cmd/tools/i18n-audit \
//	  --root=../../.. \
//	  --dict=../../../configs/i18n/ru.yml \
//	  --out=../../../docs/plans/33-i18n-audit-report.md
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "i18n-audit:", err)
		os.Exit(1)
	}
}

var (
	reCyrillic    = regexp.MustCompile(`[А-Яа-яЁё]`)
	reGoString    = regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)"`)
	reTSString    = regexp.MustCompile(`(?:'([^'\\]*(?:\\.[^'\\]*)*)'|"([^"\\]*(?:\\.[^"\\]*)*)")`)
	reTSTemplate  = regexp.MustCompile("`([^`]*)`")
	reJSXText     = regexp.MustCompile(`>([^<>{]+)<`)

	// Паттерны строк, которые нужно исключать.
	skipLinePatterns = []*regexp.Regexp{
		regexp.MustCompile(`^\s*//`),
		regexp.MustCompile(`^\s*\*`),
		regexp.MustCompile(`slog\.(Info|Warn|Error|Debug|InfoContext|WarnContext|ErrorContext|DebugContext)\(`),
		regexp.MustCompile(`log\.(Info|Warn|Error|Debug|Printf|Println|Fatal)\(`),
		regexp.MustCompile(`fmt\.Errorf\(`),
		regexp.MustCompile(`errors\.(New|As|Is)\(`),
		regexp.MustCompile(`panic\(`),
		regexp.MustCompile(`console\.(error|warn|log)\(`),
		regexp.MustCompile(`t\.Error|t\.Fatal|t\.Log|t\.Skip`),
	}
)

type Hit struct {
	File        string
	Line        int
	Literal     string
	SuggestKey  string
	InDict      bool
}

func run() error {
	root := flag.String("root", ".", "корень проекта (где frontend/ и backend/)")
	dictPath := flag.String("dict", "configs/i18n/ru.yml", "путь к ru.yml")
	out := flag.String("out", "docs/plans/33-i18n-audit-report.md", "выходной файл отчёта")
	flag.Parse()

	dict, err := loadDict(*dictPath)
	if err != nil {
		return fmt.Errorf("load dict: %w", err)
	}

	var hits []Hit

	// Сканируем backend Go-файлы.
	backendDir := filepath.Join(*root, "backend")
	if err := scanGo(backendDir, dict, &hits); err != nil {
		return fmt.Errorf("scan backend: %w", err)
	}

	// Сканируем frontend TS/TSX-файлы.
	frontendDir := filepath.Join(*root, "frontend", "src")
	if err := scanTS(frontendDir, dict, &hits); err != nil {
		return fmt.Errorf("scan frontend: %w", err)
	}

	// Сортируем по файлу и строке.
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].File != hits[j].File {
			return hits[i].File < hits[j].File
		}
		return hits[i].Line < hits[j].Line
	})

	if err := writeReport(*out, hits, *root); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	// Статистика.
	total := len(hits)
	found := 0
	for _, h := range hits {
		if h.InDict {
			found++
		}
	}
	fmt.Printf("Всего хардкод-строк: %d\n", total)
	fmt.Printf("Уже есть в ru.yml: %d (%.0f%%)\n", found, float64(found)/float64(total)*100)
	fmt.Printf("Нужно добавить: %d\n", total-found)
	fmt.Printf("Отчёт: %s\n", *out)
	return nil
}

// loadDict читает ru.yml и возвращает плоский set всех значений (для
// поиска по тексту, не по ключу).
func loadDict(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]map[string]string
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, group := range raw {
		for _, val := range group {
			if val != "" {
				set[strings.TrimSpace(val)] = true
			}
		}
	}
	return set, nil
}

func shouldSkipLine(line string) bool {
	for _, re := range skipLinePatterns {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

func isTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test.go") ||
		strings.HasSuffix(base, ".test.ts") ||
		strings.HasSuffix(base, ".test.tsx") ||
		strings.HasSuffix(base, ".spec.ts") ||
		strings.HasSuffix(base, ".spec.tsx")
}

func hasCyrillic(s string) bool {
	return reCyrillic.MatchString(s)
}

// suggestKey предлагает i18n-ключ по пути файла и тексту.
func suggestKey(relPath, text string) string {
	// Группа из папки домена.
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	group := "global"
	for _, p := range parts {
		switch p {
		case "fleet", "auth", "galaxy", "market", "battle", "alien",
			"rocket", "alliance", "chat", "payment", "admin", "wiki",
			"artefact", "shipyard", "buildings", "research", "officer",
			"expedition", "colonize", "repair", "referral", "achievement",
			"score", "settings", "economy":
			group = p
		case "features":
			// следующий элемент — домен
		}
	}

	// Ключ из первых значимых слов текста.
	words := strings.Fields(text)
	if len(words) == 0 {
		return group + ".unknown"
	}
	// Берём первые 3 слова, убираем знаки.
	var keyParts []string
	for _, w := range words {
		clean := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return -1
		}, w)
		if clean != "" {
			keyParts = append(keyParts, clean)
		}
		if len(keyParts) >= 3 {
			break
		}
	}
	key := strings.Join(keyParts, "_")
	// Упрощаем (убираем двойные подчёркивания).
	for strings.Contains(key, "__") {
		key = strings.ReplaceAll(key, "__", "_")
	}
	key = strings.ToLower(key)
	return group + "." + key
}

func addHit(hits *[]Hit, root, file string, line int, literal string, dict map[string]bool) {
	literal = strings.TrimSpace(literal)
	if !hasCyrillic(literal) {
		return
	}
	// Слишком короткие строки (1-2 символа) — пропускаем.
	if len([]rune(literal)) < 3 {
		return
	}
	rel, _ := filepath.Rel(root, file)
	inDict := dict[literal]
	*hits = append(*hits, Hit{
		File:       rel,
		Line:       line,
		Literal:    literal,
		SuggestKey: suggestKey(filepath.ToSlash(rel), literal),
		InDict:     inDict,
	})
}

func scanGo(dir string, dict map[string]bool, hits *[]Hit) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Пропускаем vendor и generated.
			base := info.Name()
			if base == "vendor" || base == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || isTestFile(path) {
			return nil
		}
		return scanGoFile(path, dict, hits)
	})
}

func scanGoFile(path string, dict map[string]bool, hits *[]Hit) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path)))) // backend/../..
	// Используем путь файла напрямую.
	scanner := bufio.NewScanner(f)
	lineNum := 0
	inBlockComment := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Блочные комментарии.
		if strings.Contains(line, "/*") {
			inBlockComment = true
		}
		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
			}
			continue
		}

		if shouldSkipLine(line) {
			continue
		}

		// Извлекаем строковые литералы.
		for _, m := range reGoString.FindAllStringSubmatch(line, -1) {
			addHit(hits, root, path, lineNum, m[1], dict)
		}
	}
	return scanner.Err()
}

func scanTS(dir string, dict map[string]bool, hits *[]Hit) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if (ext != ".ts" && ext != ".tsx") || isTestFile(path) {
			return nil
		}
		return scanTSFile(path, dict, hits)
	})
}

func scanTSFile(path string, dict map[string]bool, hits *[]Hit) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	root := filepath.Dir(filepath.Dir(filepath.Dir(path))) // frontend/src/../..
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if shouldSkipLine(line) {
			continue
		}

		// Строковые литералы в кавычках.
		for _, m := range reTSString.FindAllStringSubmatch(line, -1) {
			lit := m[1]
			if lit == "" {
				lit = m[2]
			}
			addHit(hits, root, path, lineNum, lit, dict)
		}

		// Template literals.
		for _, m := range reTSTemplate.FindAllStringSubmatch(line, -1) {
			addHit(hits, root, path, lineNum, m[1], dict)
		}

		// JSX-текст между тегами.
		for _, m := range reJSXText.FindAllStringSubmatch(line, -1) {
			addHit(hits, root, path, lineNum, strings.TrimSpace(m[1]), dict)
		}
	}
	return scanner.Err()
}

func writeReport(outPath string, hits []Hit, root string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "# i18n Audit Report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Всего: **%d** хардкод-строк с кириллицей.\n", len(hits))
	inDict := 0
	for _, h := range hits {
		if h.InDict {
			inDict++
		}
	}
	fmt.Fprintf(w, "Уже в ru.yml: **%d**. Нужно добавить: **%d**.\n", inDict, len(hits)-inDict)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| Файл | Строка | Литерал | Предложенный ключ | В ru.yml? |")
	fmt.Fprintln(w, "|---|---|---|---|---|")

	for _, h := range hits {
		rel := strings.TrimPrefix(filepath.ToSlash(h.File), filepath.ToSlash(root)+"/")
		lit := strings.ReplaceAll(h.Literal, "|", "\\|")
		if len([]rune(lit)) > 60 {
			lit = string([]rune(lit)[:57]) + "…"
		}
		inDictMark := "❌"
		if h.InDict {
			inDictMark = "✅"
		}
		fmt.Fprintf(w, "| %s | %d | %s | `%s` | %s |\n",
			rel, h.Line, lit, h.SuggestKey, inDictMark)
	}

	return w.Flush()
}
