// Package moderation — UGC-модерация для никнеймов, чата, описаний.
// План 46 (149-ФЗ): blacklist запрещённых слов, проверка пользовательского
// ввода. Источник истины — configs/moderation/blacklist.yaml (корень репо).
package moderation

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrForbidden — введённое значение содержит запрещённое слово.
var ErrForbidden = errors.New("moderation: forbidden word")

// Blacklist — набор запрещённых корней. Сравнение по подстроке после
// нормализации (lowercase + удаление пробелов и неалфавитных символов).
type Blacklist struct {
	roots []string
}

// LoadBlacklist читает YAML и собирает плоский список корней.
// Формат YAML — map с произвольными группами (profanity_ru, drugs, …) →
// массив строк. Группы носят информационный характер.
func LoadBlacklist(path string) (*Blacklist, error) {
	data, err := os.ReadFile(path) //nolint:gosec // путь конфигурируется при старте
	if err != nil {
		return nil, fmt.Errorf("moderation: read %s: %w", path, err)
	}
	var raw map[string][]string
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("moderation: parse %s: %w", path, err)
	}
	bl := &Blacklist{}
	for _, words := range raw {
		for _, w := range words {
			n := normalize(w)
			if n != "" {
				bl.roots = append(bl.roots, n)
			}
		}
	}
	return bl, nil
}

// NewBlacklist собирает Blacklist из готового списка корней (для тестов
// и форсированной инъекции через env).
func NewBlacklist(roots []string) *Blacklist {
	bl := &Blacklist{}
	for _, w := range roots {
		n := normalize(w)
		if n != "" {
			bl.roots = append(bl.roots, n)
		}
	}
	return bl
}

// IsForbidden проверяет, содержит ли input запрещённый корень.
// Возвращает (true, root) если найден; (false, "") иначе.
//
// Защита от обхода: нормализуем (нижний регистр, убираем пробелы и
// неалфавитные символы) перед поиском. Не идеально (l33t-speak обходит),
// но достаточно для запуска. Для серьёзной модерации — план +N с
// сторонним сервисом.
func (b *Blacklist) IsForbidden(input string) (bool, string) {
	if b == nil || len(b.roots) == 0 {
		return false, ""
	}
	n := normalize(input)
	if n == "" {
		return false, ""
	}
	for _, root := range b.roots {
		if strings.Contains(n, root) {
			return true, root
		}
	}
	return false, ""
}

// Size — количество корней в blacklist (для логирования при старте).
func (b *Blacklist) Size() int {
	if b == nil {
		return 0
	}
	return len(b.roots)
}

// normalize приводит строку к виду для сравнения: нижний регистр,
// удалены пробелы, цифры, пунктуация — остаются только буквы (любые
// алфавиты, включая кириллицу). Этим режется простой обходной паттерн
// "h e r o i n" → "heroin".
func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isLetter(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'а' && r <= 'я') || r == 'ё' ||
		(r >= 'A' && r <= 'Z') ||
		(r >= 'А' && r <= 'Я') || r == 'Ё'
}
