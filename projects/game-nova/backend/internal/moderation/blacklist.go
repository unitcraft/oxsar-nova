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
// массив строк.
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

// NewBlacklist — конструктор для тестов.
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

// IsForbidden — true, если input содержит запрещённый корень.
// Защита от обхода: нормализуем перед поиском (lowercase, убираем
// пробелы, цифры и пунктуацию).
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

// MaskForbidden заменяет найденные корни на звёздочки в исходном тексте,
// сохраняя длину. Используется для чата (план 46 Ф.4).
//
// Не идеально: совпадение ищется в нормализованном виде, а маскирование
// — в исходном по совпадающим позициям букв. Для запуска достаточно;
// для серьёзной модерации — сторонний сервис.
func (b *Blacklist) MaskForbidden(input string) string {
	if found, _ := b.IsForbidden(input); !found {
		return input
	}
	// Простая стратегия: если найдено хоть одно запрещённое — режем
	// все буквенные последовательности длиной >=3 на звёздочки.
	// Это «крупно», но безопасно: лучше перерезать, чем пропустить.
	runes := []rune(input)
	out := make([]rune, len(runes))
	for i, r := range runes {
		if isLetter(r) {
			out[i] = '*'
		} else {
			out[i] = r
		}
	}
	return string(out)
}

// Size — количество корней (для логов при старте).
func (b *Blacklist) Size() int {
	if b == nil {
		return 0
	}
	return len(b.roots)
}

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
