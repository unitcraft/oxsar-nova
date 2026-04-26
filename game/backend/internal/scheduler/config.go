// Package scheduler — обёртка над robfig/cron/v3 для периодических задач
// worker'а с distributed lock через Postgres advisory locks (план 32).
//
// Каждая job регистрируется парой (name, fn). Расписание читается из
// configs/schedule.yaml (или ENV-override'ов). На каждом cron-tick вызов
// fn оборачивается в locks.TryRun — при N≥2 worker'ах ровно один
// инстанс выполняет, остальные тихо инкрементят skip-метрику.
//
// При acquired=false err остаётся nil, поэтому cron не ретраит. Ошибка
// fn логируется и идёт в counter status="error", но cron продолжит
// работать по расписанию.
package scheduler

import (
	"fmt"
	"os"
	"strings"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

// JobConfig — расписание одной job'ы.
type JobConfig struct {
	// Schedule — стандартное cron-выражение (5 полей) или "@every <duration>".
	Schedule string `yaml:"schedule"`

	// Enabled — если false, job не регистрируется. Удобно отключать
	// конкретную задачу через ENV без правки YAML.
	Enabled bool `yaml:"enabled"`

	// Description — справочное поле, пишется в лог при регистрации.
	Description string `yaml:"description"`
}

// Config — корневой YAML, отображается на configs/schedule.yaml.
type Config struct {
	Jobs map[string]JobConfig `yaml:"jobs"`
}

// LoadConfig читает YAML из файла и применяет ENV-override'ы.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("scheduler: read %q: %w", path, err)
	}
	cfg, err := ParseConfig(data)
	if err != nil {
		return nil, fmt.Errorf("scheduler: parse %q: %w", path, err)
	}
	return cfg, nil
}

// ParseConfig парсит YAML и применяет ENV-override'ы. Выделено отдельно
// от LoadConfig для тестов, не зависящих от файловой системы.
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("yaml: %w", err)
	}
	if cfg.Jobs == nil {
		cfg.Jobs = map[string]JobConfig{}
	}
	applyEnvOverrides(&cfg, os.Environ())
	if err := validateSchedules(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// applyEnvOverrides применяет переопределения вида:
//
//	SCHEDULER_<JOB>_CRON     — заменяет schedule
//	SCHEDULER_<JOB>_ENABLED  — true|false для enabled
//
// JOB переводится в lower_case и матчится по имени из YAML. Если в YAML
// такой job нет — override игнорируется (с silent-skip; добавление
// jobs только из ENV не поддерживается, чтобы не было задач без описания).
func applyEnvOverrides(cfg *Config, environ []string) {
	for _, kv := range environ {
		i := strings.IndexByte(kv, '=')
		if i < 0 {
			continue
		}
		key, val := kv[:i], kv[i+1:]
		if !strings.HasPrefix(key, "SCHEDULER_") {
			continue
		}
		rest := strings.TrimPrefix(key, "SCHEDULER_")

		switch {
		case strings.HasSuffix(rest, "_CRON"):
			job := strings.ToLower(strings.TrimSuffix(rest, "_CRON"))
			if jc, ok := cfg.Jobs[job]; ok {
				jc.Schedule = val
				cfg.Jobs[job] = jc
			}
		case strings.HasSuffix(rest, "_ENABLED"):
			job := strings.ToLower(strings.TrimSuffix(rest, "_ENABLED"))
			if jc, ok := cfg.Jobs[job]; ok {
				jc.Enabled = strings.EqualFold(val, "true") || val == "1"
				cfg.Jobs[job] = jc
			}
		}
	}
}

// validateSchedules — fail-fast при invalid cron-выражении. Лучше
// упасть при старте, чем молча не запускать job.
func validateSchedules(cfg *Config) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	for name, jc := range cfg.Jobs {
		if !jc.Enabled {
			continue
		}
		if jc.Schedule == "" {
			return fmt.Errorf("scheduler: job %q: empty schedule", name)
		}
		if _, err := parser.Parse(jc.Schedule); err != nil {
			return fmt.Errorf("scheduler: job %q: invalid schedule %q: %w", name, jc.Schedule, err)
		}
	}
	return nil
}

// DisabledByEnv возвращает true, если SCHEDULER_DISABLE_ALL=true
// (kill-switch для всего scheduler'а).
func DisabledByEnv() bool {
	v := os.Getenv("SCHEDULER_DISABLE_ALL")
	return strings.EqualFold(v, "true") || v == "1"
}
