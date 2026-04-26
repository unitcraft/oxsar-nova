// Package i18n — locale loader и Tr() для серверных сообщений.
//
// Источник истины — configs/i18n/ru.yml и en.yml (см. §10.3 ТЗ).
//
// Принципы:
//   - immutable: загружается один раз на старте, в рантайме read-only;
//   - fallback: если ключа нет в запрошенном языке, падаем в ru;
//     если и в ru нет — возвращаем маркер "[group.key]";
//   - плейсхолдеры — только именованные {{name}}, подставляются через vars.
package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Lang — ISO-like код языка.
type Lang string

const (
	LangRu Lang = "ru"
	LangEn Lang = "en"
)

// Dict — загруженный словарь одной локали.
// Внешняя карта: группа → (ключ → текст).
type Dict map[string]map[string]string

// Bundle — все загруженные локали.
type Bundle struct {
	fallback Lang
	locales  map[Lang]Dict
}

// Load читает все *.yml в dir как локали (имя файла до .yml = Lang).
// fallback — язык, на который падаем при отсутствии ключа. Обычно ru.
// Если fallback-локали нет в dir, Load возвращает ошибку.
func Load(dir string, fallback Lang) (*Bundle, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("i18n: read dir: %w", err)
	}
	b := &Bundle{fallback: fallback, locales: map[Lang]Dict{}}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yml") {
			continue
		}
		lang := Lang(strings.TrimSuffix(e.Name(), ".yml"))
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("i18n: read %s: %w", e.Name(), err)
		}
		var d Dict
		if err := yaml.Unmarshal(data, &d); err != nil {
			return nil, fmt.Errorf("i18n: parse %s: %w", e.Name(), err)
		}
		b.locales[lang] = d
	}
	if _, ok := b.locales[fallback]; !ok {
		return nil, fmt.Errorf("i18n: fallback locale %q not loaded", fallback)
	}
	return b, nil
}

// Tr возвращает переведённую строку.
//
// Порядок поиска:
//  1. locales[lang][group][key]
//  2. locales[fallback][group][key]
//  3. "[group.key]"
//
// vars — именованные плейсхолдеры {{name}} → значение. Nil допустим.
func (b *Bundle) Tr(lang Lang, group, key string, vars map[string]string) string {
	if s, ok := b.lookup(lang, group, key); ok {
		return interpolate(s, vars)
	}
	if lang != b.fallback {
		if s, ok := b.lookup(b.fallback, group, key); ok {
			return interpolate(s, vars)
		}
	}
	return "[" + group + "." + key + "]"
}

// Has проверяет, есть ли ключ именно в этой локали (без fallback'а).
// Пригодится в тестах и CI («все новые ключи добавлены в ru.yml»).
func (b *Bundle) Has(lang Lang, group, key string) bool {
	_, ok := b.lookup(lang, group, key)
	return ok
}

// Languages возвращает список загруженных локалей (для /api/i18n).
func (b *Bundle) Languages() []Lang {
	out := make([]Lang, 0, len(b.locales))
	for l := range b.locales {
		out = append(out, l)
	}
	return out
}

// Locale возвращает полный словарь локали (для фронта через /api/i18n/{lang}).
func (b *Bundle) Locale(lang Lang) Dict {
	if d, ok := b.locales[lang]; ok {
		return d
	}
	return b.locales[b.fallback]
}

func (b *Bundle) lookup(lang Lang, group, key string) (string, bool) {
	d, ok := b.locales[lang]
	if !ok {
		return "", false
	}
	g, ok := d[group]
	if !ok {
		return "", false
	}
	v, ok := g[key]
	return v, ok
}

var rePlaceholder = regexp.MustCompile(`\{\{(\w+)\}\}`)

func interpolate(tmpl string, vars map[string]string) string {
	if len(vars) == 0 {
		return tmpl
	}
	return rePlaceholder.ReplaceAllStringFunc(tmpl, func(m string) string {
		name := m[2 : len(m)-2]
		if v, ok := vars[name]; ok {
			return v
		}
		return m
	})
}
