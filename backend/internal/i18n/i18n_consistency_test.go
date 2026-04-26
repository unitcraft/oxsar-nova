package i18n_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestI18nConsistency проверяет, что множества ключей в ru.yml и en.yml
// совпадают. Дрейф (ключ добавлен в одном файле, но забыт в другом)
// ловится в PR ещё до мержа.
func TestI18nConsistency(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "configs", "i18n")

	ruKeys := loadKeys(t, filepath.Join(dir, "ru.yml"))
	enKeys := loadKeys(t, filepath.Join(dir, "en.yml"))

	onlyInRu := diff(ruKeys, enKeys)
	onlyInEn := diff(enKeys, ruKeys)

	for _, k := range onlyInRu {
		t.Errorf("ключ %q есть в ru.yml, но отсутствует в en.yml", k)
	}
	for _, k := range onlyInEn {
		t.Errorf("ключ %q есть в en.yml, но отсутствует в ru.yml", k)
	}
}

// loadKeys читает YAML-файл и возвращает отсортированный список "group.key".
func loadKeys(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	var keys []string
	for group, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		for key := range m {
			if !strings.HasPrefix(key, "#") {
				keys = append(keys, group+"."+key)
			}
		}
	}
	sort.Strings(keys)
	return keys
}

func diff(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, k := range b {
		set[k] = struct{}{}
	}
	var out []string
	for _, k := range a {
		if _, ok := set[k]; !ok {
			out = append(out, k)
		}
	}
	return out
}
