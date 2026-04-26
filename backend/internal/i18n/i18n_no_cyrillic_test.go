package i18n_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestNoCyrillicLiterals проверяет, что в прод-коде backend нет
// хардкод-кириллицы в строковых литералах. Все user-facing строки
// должны идти через Bundle.Tr (configs/i18n/*.yml).
//
// Исключения:
//   - *_test.go, *_mock.go — тестовые данные разрешены
//   - строки в комментариях (// и /* */) — разрешены
//   - migrations/, cmd/tools/ — SQL-фикстуры и одноразовые утилиты
//   - slog.*/fmt.Errorf с %w — внутренние ошибки не идут к пользователю
func TestNoCyrillicLiterals(t *testing.T) {
	reCyrillic := regexp.MustCompile(`[А-Яа-яЁё]`)

	// Паттерн строкового литерала с кириллицей:
	// захватывает содержимое между кавычками/бэктиками
	reStringLit := regexp.MustCompile("(?:\"[^\"]*[А-Яа-яЁё][^\"]*\"|`[^`]*[А-Яа-яЁё][^`]*`)")

	// Паттерны строк, которые разрешены (внутренние, не user-facing)
	allowedPatterns := []*regexp.Regexp{
		// fmt.Errorf / errors.New — внутренние ошибки разработчика
		regexp.MustCompile(`fmt\.Errorf\(`),
		regexp.MustCompile(`errors\.New\(`),
		// slog.* логирование
		regexp.MustCompile(`slog\.(Info|Warn|Error|Debug)\(`),
		// panic — только bootstrap
		regexp.MustCompile(`panic\(`),
	}

	// Каталоги для обхода
	root := filepath.Join("..", "..", "..")
	searchDirs := []string{
		filepath.Join(root, "backend", "internal"),
		filepath.Join(root, "backend", "cmd", "server"),
		filepath.Join(root, "backend", "cmd", "worker"),
	}

	var failures []string

	for _, dir := range searchDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			// Исключить тестовые и mock файлы
			base := filepath.Base(path)
			if strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, "_mock.go") {
				return nil
			}
			// Исключить файлы с нелитеральными строками: AI-промпты (aiadvisor)
			// и конфигурацию платёжного шлюза (payment/packages) — не user-facing.
			if strings.Contains(path, "aiadvisor") || strings.Contains(path, "payment"+string(os.PathSeparator)+"packages") {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			lineNum := 0
			inBlockComment := false

			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				trimmed := strings.TrimSpace(line)

				// Блочные комментарии /* ... */
				if inBlockComment {
					if strings.Contains(line, "*/") {
						inBlockComment = false
					}
					continue
				}
				if strings.Contains(line, "/*") && !strings.Contains(line, "*/") {
					inBlockComment = true
					continue
				}

				// Строчные комментарии — убрать хвост
				if idx := strings.Index(line, "//"); idx >= 0 {
					line = line[:idx]
				}

				// Нет кириллицы в остатке — пропустить
				if !reCyrillic.MatchString(line) {
					continue
				}

				// Нет строковых литералов с кириллицей — пропустить
				if !reStringLit.MatchString(line) {
					continue
				}

				// Проверить allowed patterns (весь trimmed, не обрезанный)
				allowed := false
				for _, pat := range allowedPatterns {
					if pat.MatchString(trimmed) {
						allowed = true
						break
					}
				}
				if allowed {
					continue
				}

				rel, _ := filepath.Rel(root, path)
				failures = append(failures, rel+":"+strconv.Itoa(lineNum)+": "+strings.TrimSpace(trimmed))
			}
			return scanner.Err()
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}

	for _, f := range failures {
		t.Errorf("кириллица в строковом литерале (используй Bundle.Tr): %s", f)
	}
}
