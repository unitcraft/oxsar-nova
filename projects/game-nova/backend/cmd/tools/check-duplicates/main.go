// Command check-duplicates сравнивает содержимое DUPLICATE-файлов
// между Go-модулями oxsar/{game-nova,identity,portal,billing}.
//
// Каждый DUPLICATE-файл начинается с шапки вида:
//
//	// DUPLICATE: этот файл скопирован между Go-модулями ...
//	//   - projects/game-nova/backend/pkg/ids/ids.go
//	//   - projects/identity/backend/pkg/ids/ids.go
//	// Причина дубля: ...
//
// Утилита:
//  1. Walk'ает projects/*/backend/, исключая vendor/dist/node_modules.
//  2. Находит файлы с маркером `// DUPLICATE: этот файл скопирован`.
//  3. Парсит шапку каждого: список путей-копий + сам блок маркера.
//  4. Группирует копии по списку путей (если списки расходятся —
//     это уже drift самой шапки).
//  5. Срезает шапку (от строки `// DUPLICATE:` до первой пустой
//     строки после `// Причина дубля:`) и хеширует остаток (SHA256).
//  6. Все хеши в группе должны совпасть. Иначе — печатает unified-diff
//     между эталоном (первый путь в списке) и расходящимися копиями,
//     exit 1.
//
// Флаг -fix: тиражирует тело эталона в каждую копию (с сохранением
// её собственной шапки). Использовать вручную после сознательной
// правки эталона, не в CI.
//
// Запуск:
//
//	go run ./projects/game-nova/backend/cmd/tools/check-duplicates
//	go run ./projects/game-nova/backend/cmd/tools/check-duplicates -fix
//
// План 85.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	markerStart  = "// DUPLICATE: этот файл скопирован"
	markerPath   = "//   - "
	markerReason = "// Причина дубля:"
)

// duplicateFile — одна найденная копия.
type duplicateFile struct {
	// path — относительный к корню репо путь со слешами `/`.
	path string
	// module — имя модуля (game-nova / identity / portal / billing),
	// извлечённое из path как projects/<module>/backend/...
	module string
	// paths — список путей-копий, как объявлен в шапке файла.
	paths []string
	// headerEnd — индекс первой строки ПОСЛЕ шапки (тело начинается с него).
	headerEnd int
	// lines — все строки файла (без trailing '\n').
	lines []string
	// bodyHash — SHA256 от нормализованного тела (см. normalizeBody).
	bodyHash string
}

// modulePathPlaceholder — общая замена для `oxsar/<module>/` в import-строках,
// чтобы per-module различия не считались drift'ом. Конкретный <module> в
// каждой копии — необходимое следствие отдельных go.mod, а не drift.
const modulePathPlaceholder = "oxsar/__MODULE__/"

// normalizeBody заменяет `"oxsar/<module>/` на placeholder, чтобы хеши
// совпали при идентичной семантике, но per-module import-prefix.
// Применяется только к строкам, начинающимся с `"oxsar/` (typical Go-import).
func normalizeBody(lines []string, module string) []string {
	if module == "" {
		return lines
	}
	prefix := "oxsar/" + module + "/"
	out := make([]string, len(lines))
	for i, ln := range lines {
		out[i] = strings.ReplaceAll(ln, prefix, modulePathPlaceholder)
	}
	return out
}

func main() {
	fix := flag.Bool("fix", false, "перезаписать копии содержимым эталона (тело)")
	root := flag.String("root", "", "корень репо (по умолчанию — поиск вверх от cwd)")
	flag.Parse()

	repoRoot, err := resolveRoot(*root)
	if err != nil {
		fatalf("не удалось найти корень репо: %v", err)
	}

	files, err := scan(repoRoot)
	if err != nil {
		fatalf("scan: %v", err)
	}
	if len(files) == 0 {
		fmt.Println("DUPLICATE-файлов не найдено — нечего проверять")
		return
	}

	groups, warnings := groupByPaths(files)
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "WARN:", w)
	}

	hasDrift := false
	for _, g := range groups {
		drift, err := checkGroup(repoRoot, g, *fix)
		if err != nil {
			fatalf("group %v: %v", g[0].paths, err)
		}
		if drift {
			hasDrift = true
		}
	}

	if hasDrift && !*fix {
		fmt.Fprintf(os.Stderr, "\nDRIFT: расхождение между копиями. Запустите `make sync-duplicates` или `go run ./cmd/tools/check-duplicates -fix`.\n")
		os.Exit(1)
	}

	totalFiles := 0
	for _, g := range groups {
		totalFiles += len(g)
	}
	if *fix {
		fmt.Printf("OK (fix): %d групп, %d файлов синхронизированы по эталону\n", len(groups), totalFiles)
	} else if len(warnings) > 0 {
		fmt.Printf("OK с предупреждениями: %d групп, %d файлов; см. WARN выше\n", len(groups), totalFiles)
	} else {
		fmt.Printf("OK: %d групп, %d файлов синхронны\n", len(groups), totalFiles)
	}
}

// resolveRoot ищет корень репо: либо через -root, либо ищем вверх до файла CLAUDE.md.
func resolveRoot(explicit string) (string, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for d := cwd; ; {
		if _, err := os.Stat(filepath.Join(d, "CLAUDE.md")); err == nil {
			// Дополнительная проверка — есть директория projects/.
			if _, err := os.Stat(filepath.Join(d, "projects")); err == nil {
				return d, nil
			}
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", fmt.Errorf("не нашли CLAUDE.md, начиная с %s", cwd)
		}
		d = parent
	}
}

// scan обходит projects/*/backend/, ищет .go-файлы с DUPLICATE-маркером.
func scan(repoRoot string) ([]duplicateFile, error) {
	projectsDir := filepath.Join(repoRoot, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("read projects/: %w", err)
	}
	var out []duplicateFile
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		backend := filepath.Join(projectsDir, e.Name(), "backend")
		if _, err := os.Stat(backend); err != nil {
			continue
		}
		err := filepath.WalkDir(backend, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				name := d.Name()
				if name == "vendor" || name == "node_modules" || name == "dist" || name == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			df, ok, err := parseFile(repoRoot, path)
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			if ok {
				out = append(out, df)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].path < out[j].path })
	return out, nil
}

// parseFile читает файл, ищет DUPLICATE-маркер и парсит шапку.
// Возвращает (df, true, nil), если файл — DUPLICATE; (_, false, nil) иначе.
func parseFile(repoRoot, absPath string) (duplicateFile, bool, error) {
	rel, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return duplicateFile{}, false, err
	}
	rel = filepath.ToSlash(rel)

	// Извлекаем имя модуля: projects/<module>/backend/...
	module := ""
	parts := strings.Split(rel, "/")
	if len(parts) >= 3 && parts[0] == "projects" && parts[2] == "backend" {
		module = parts[1]
	}

	f, err := os.Open(absPath)
	if err != nil {
		return duplicateFile{}, false, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return duplicateFile{}, false, err
	}

	startIdx := -1
	for i, ln := range lines {
		if strings.HasPrefix(ln, markerStart) {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return duplicateFile{}, false, nil
	}

	// Парсим шапку: собираем пути и ищем строку "Причина дубля:".
	var paths []string
	reasonIdx := -1
	headerEnd := -1
	for i := startIdx; i < len(lines); i++ {
		ln := lines[i]
		trim := strings.TrimRight(ln, " \t")
		// Пустая строка — потенциальный конец шапки.
		if trim == "" {
			// Конец — если уже видели "Причина дубля:", или если есть пути и нет reason.
			if reasonIdx != -1 || len(paths) > 0 {
				headerEnd = i + 1 // тело начинается со следующей строки
				break
			}
			continue
		}
		// Строка должна быть комментарием — иначе шапка кончилась резко.
		if !strings.HasPrefix(trim, "//") {
			// Не пустая, не комментарий — шапка закончилась без явной пустой строки.
			headerEnd = i
			break
		}
		if strings.HasPrefix(ln, markerPath) {
			p := strings.TrimSpace(strings.TrimPrefix(ln, markerPath))
			if p != "" {
				paths = append(paths, p)
			}
		}
		if strings.HasPrefix(ln, markerReason) {
			reasonIdx = i
		}
	}
	if headerEnd == -1 {
		// Шапка тянется до EOF (странно, но обработаем).
		headerEnd = len(lines)
	}
	if len(paths) == 0 {
		return duplicateFile{}, false, fmt.Errorf("маркер DUPLICATE найден, но список путей пустой")
	}

	// Хеш нормализованного тела (с заменой per-module import-prefix).
	bodyLines := normalizeBody(lines[headerEnd:], module)
	body := strings.Join(bodyLines, "\n")
	sum := sha256.Sum256([]byte(body))

	return duplicateFile{
		path:      rel,
		module:    module,
		paths:     paths,
		headerEnd: headerEnd,
		lines:     lines,
		bodyHash:  hex.EncodeToString(sum[:]),
	}, true, nil
}

// groupByPaths группирует копии по их списку path'ов. Возвращает группы
// и warnings — список найденных нестыковок (файл объявляет себя в группе,
// но другая копия из той же группы объявляет другой набор путей).
func groupByPaths(files []duplicateFile) ([][]duplicateFile, []string) {
	// Ключ группы — отсортированный список путей, склеенный через '\n'.
	byKey := map[string][]duplicateFile{}
	for _, f := range files {
		key := groupKey(f.paths)
		byKey[key] = append(byKey[key], f)
	}

	// Дополнительная проверка: каждый файл должен фигурировать в своём же списке.
	var warnings []string
	for _, f := range files {
		found := false
		for _, p := range f.paths {
			if p == f.path {
				found = true
				break
			}
		}
		if !found {
			warnings = append(warnings, fmt.Sprintf("%s: файл не упомянут в собственном списке копий %v", f.path, f.paths))
		}
	}

	// Каждая группа должна содержать ровно столько копий, сколько указано в путях.
	var groups [][]duplicateFile
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		g := byKey[k]
		// Сортируем группу в порядке, указанном в paths[0] (эталон — первый).
		order := map[string]int{}
		for i, p := range g[0].paths {
			order[p] = i
		}
		sort.Slice(g, func(i, j int) bool {
			oi, ok1 := order[g[i].path]
			oj, ok2 := order[g[j].path]
			if !ok1 {
				oi = 1 << 30
			}
			if !ok2 {
				oj = 1 << 30
			}
			if oi != oj {
				return oi < oj
			}
			return g[i].path < g[j].path
		})
		expected := len(g[0].paths)
		if len(g) != expected {
			present := map[string]bool{}
			for _, f := range g {
				present[f.path] = true
			}
			var missing []string
			for _, p := range g[0].paths {
				if !present[p] {
					missing = append(missing, p)
				}
			}
			if len(missing) > 0 {
				warnings = append(warnings, fmt.Sprintf("группа %v: отсутствуют копии %v", g[0].paths, missing))
			}
		}
		groups = append(groups, g)
	}
	return groups, warnings
}

func groupKey(paths []string) string {
	c := make([]string, len(paths))
	copy(c, paths)
	sort.Strings(c)
	return strings.Join(c, "\n")
}

// checkGroup сравнивает хеши тел всех копий в группе. При расхождении —
// печатает diff. При -fix — тиражирует тело эталона.
// Возвращает (drift, error): drift=true означает расхождение (актуально
// только в режиме без -fix; с -fix drift всегда false).
func checkGroup(repoRoot string, g []duplicateFile, fix bool) (bool, error) {
	if len(g) < 2 {
		// Группа из одного файла — проверять нечего, но это подозрительно.
		fmt.Fprintf(os.Stderr, "WARN: группа %v содержит только %d файл(ов) из %d ожидаемых\n",
			g[0].paths, len(g), len(g[0].paths))
		return false, nil
	}
	ref := g[0]
	drift := false
	for _, c := range g[1:] {
		if c.bodyHash == ref.bodyHash {
			continue
		}
		drift = true
		if fix {
			if err := rewriteBody(repoRoot, c, ref); err != nil {
				return false, fmt.Errorf("rewrite %s: %w", c.path, err)
			}
			fmt.Printf("FIX: %s ← %s\n", c.path, ref.path)
		} else {
			fmt.Fprintf(os.Stderr, "\nDRIFT: %s\n  vs эталон %s\n", c.path, ref.path)
			// Сравниваем нормализованные тела, чтобы per-module import-paths
			// не загромождали diff. Реальный drift — всё остальное.
			refBody := normalizeBody(ref.lines[ref.headerEnd:], ref.module)
			cpBody := normalizeBody(c.lines[c.headerEnd:], c.module)
			printUnifiedDiff(ref.path, c.path, refBody, cpBody)
		}
	}
	if !drift && !fix {
		fmt.Printf("ok: %s (×%d)\n", g[0].paths[0], len(g))
	}
	return drift && !fix, nil
}

// rewriteBody перезаписывает файл copy: оставляет его шапку (lines[:headerEnd]),
// тело берёт из ref (lines[headerEnd:]) с заменой per-module import-prefix
// `oxsar/<refModule>/` → `oxsar/<copyModule>/`, чтобы код собирался в copy.
func rewriteBody(repoRoot string, c, ref duplicateFile) error {
	abs := filepath.Join(repoRoot, filepath.FromSlash(c.path))
	header := append([]string{}, c.lines[:c.headerEnd]...)
	body := append([]string{}, ref.lines[ref.headerEnd:]...)
	if ref.module != "" && c.module != "" && ref.module != c.module {
		refPrefix := "oxsar/" + ref.module + "/"
		copyPrefix := "oxsar/" + c.module + "/"
		for i, ln := range body {
			body[i] = strings.ReplaceAll(ln, refPrefix, copyPrefix)
		}
	}
	out := append(header, body...)
	content := strings.Join(out, "\n")
	// По соглашению у Go-файлов всегда есть финальный '\n'.
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(abs, []byte(content), 0o644)
}

// printUnifiedDiff — простой ручной unified-diff построчно.
// Не пытается схлопывать hunk'и — печатает все строки с маркерами,
// этого достаточно для CLI-вывода о drift'е (файлы маленькие).
func printUnifiedDiff(refPath, copyPath string, ref, cp []string) {
	fmt.Fprintf(os.Stderr, "--- %s\n+++ %s\n", refPath, copyPath)
	// LCS-based diff. n*m — но файлы небольшие.
	a, b := ref, cp
	n, m := len(a), len(b)
	// dp[i][j] = LCS длина для a[i:], b[j:]
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	i, j := 0, 0
	for i < n && j < m {
		if a[i] == b[j] {
			fmt.Fprintln(os.Stderr, " "+a[i])
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			fmt.Fprintln(os.Stderr, "-"+a[i])
			i++
		} else {
			fmt.Fprintln(os.Stderr, "+"+b[j])
			j++
		}
	}
	for ; i < n; i++ {
		fmt.Fprintln(os.Stderr, "-"+a[i])
	}
	for ; j < m; j++ {
		fmt.Fprintln(os.Stderr, "+"+b[j])
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(2)
}
