// Package features — feature flags для безопасной выкатки рефакторингов.
//
// План 31 Ф.2. Используется парами с health/draining (Ф.1):
// рефакторинг катится за флагом (новый код мёртв при flag=false), при
// проблеме — переключаем flag и restart backend (graceful + drain).
//
// Источник истины — YAML-файл (configs/features.yaml). Изменения
// требуют restart процесса (hot-reload не реализован сознательно:
// restart дёшев + предсказуем).
//
// Использование (server/worker):
//
//	flags, err := features.Load("configs/features.yaml")
//	if features.Enabled(flags, "goal_engine") {
//	    return goalEngine.Handle(...)
//	}
//	return oldAchievementSvc.Handle(...)
//
// Endpoint /api/features (GET, без auth) возвращает список включенных
// флагов для UI — позволяет фронтенду рисовать UI conditionally.
package features

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Set — immutable набор флагов, прочитанный из YAML на старте.
//
// Безопасен для concurrent-чтения (значения только читаются после Load).
type Set struct {
	flags map[string]Flag
	mu    sync.RWMutex // для тестов / hot-reload в будущем
}

// Flag — метаданные флага: enabled + описание (для документации и /api/features).
type Flag struct {
	Enabled     bool   `yaml:"enabled"`
	Description string `yaml:"description"`
}

// fileSchema — корневая структура features.yaml.
type fileSchema struct {
	Features map[string]Flag `yaml:"features"`
}

// Load читает YAML-файл и возвращает Set. Если файла нет — возвращает
// пустой Set (все флаги = false), не ошибку — это валидный сценарий
// для dev/test.
func Load(path string) (*Set, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Set{flags: map[string]Flag{}}, nil
		}
		return nil, fmt.Errorf("features: read %s: %w", path, err)
	}
	return ParseBytes(data)
}

// ParseBytes — парсит YAML-байты. Используется в Load и тестах.
func ParseBytes(data []byte) (*Set, error) {
	var fs fileSchema
	if err := yaml.Unmarshal(data, &fs); err != nil {
		return nil, fmt.Errorf("features: parse: %w", err)
	}
	if fs.Features == nil {
		fs.Features = map[string]Flag{}
	}
	return &Set{flags: fs.Features}, nil
}

// Enabled возвращает true, если флаг key установлен в enabled=true.
// Неизвестные ключи возвращают false (fail-closed: новый код за флагом
// не активируется случайно).
//
// Принимает *Set, чтобы вызовы вида `features.Enabled(s, "goal_engine")`
// читались как «спросить у Set: enabled?». При s=nil возвращает false.
func Enabled(s *Set, key string) bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[key]
	return ok && f.Enabled
}

// All возвращает копию map[name]Flag для UI / документации /api/features.
func All(s *Set) map[string]Flag {
	if s == nil {
		return map[string]Flag{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Flag, len(s.flags))
	for k, v := range s.flags {
		out[k] = v
	}
	return out
}

// EnabledKeys возвращает отсортированный список ключей, где enabled=true.
// Удобно для UI — короче, чем All().
func EnabledKeys(s *Set) []string {
	all := All(s)
	out := make([]string, 0, len(all))
	for k, v := range all {
		if v.Enabled {
			out = append(out, k)
		}
	}
	// Сортируем для стабильного output (тесты, /api/features).
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
