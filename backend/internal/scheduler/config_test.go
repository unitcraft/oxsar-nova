package scheduler

import (
	"strings"
	"testing"
)

func TestParseConfig_Basic(t *testing.T) {
	yaml := `
jobs:
  alien_spawn:
    schedule: "0 */6 * * *"
    enabled: true
    description: "spawn"
  inactivity_reminders:
    schedule: "0 9 * * *"
    enabled: false
    description: "off"
`
	cfg, err := ParseConfig([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := cfg.Jobs["alien_spawn"].Schedule; got != "0 */6 * * *" {
		t.Errorf("alien_spawn schedule = %q", got)
	}
	if !cfg.Jobs["alien_spawn"].Enabled {
		t.Error("alien_spawn must be enabled")
	}
	if cfg.Jobs["inactivity_reminders"].Enabled {
		t.Error("inactivity_reminders must be disabled")
	}
}

func TestParseConfig_InvalidCron(t *testing.T) {
	yaml := `
jobs:
  bad:
    schedule: "not a cron"
    enabled: true
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid cron")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("error must mention job name: %v", err)
	}
}

func TestParseConfig_EmptyScheduleEnabled(t *testing.T) {
	yaml := `
jobs:
  bad:
    enabled: true
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for empty schedule on enabled job")
	}
}

func TestParseConfig_DisabledJobSkipsValidation(t *testing.T) {
	// Disabled job не валидируется — это позволяет «закомментировать»
	// сломанную job через enabled:false без удаления.
	yaml := `
jobs:
  bad:
    schedule: "garbage"
    enabled: false
`
	if _, err := ParseConfig([]byte(yaml)); err != nil {
		t.Fatalf("disabled job must skip validation, got: %v", err)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := &Config{Jobs: map[string]JobConfig{
		"alien_spawn":          {Schedule: "0 */6 * * *", Enabled: true},
		"inactivity_reminders": {Schedule: "0 9 * * *", Enabled: true},
	}}
	applyEnvOverrides(cfg, []string{
		"SCHEDULER_ALIEN_SPAWN_CRON=0 */3 * * *",
		"SCHEDULER_INACTIVITY_REMINDERS_ENABLED=false",
		"SCHEDULER_UNKNOWN_CRON=ignored",
		"NOT_SCHEDULER=skip",
	})
	if got := cfg.Jobs["alien_spawn"].Schedule; got != "0 */3 * * *" {
		t.Errorf("alien_spawn schedule = %q", got)
	}
	if cfg.Jobs["inactivity_reminders"].Enabled {
		t.Error("inactivity_reminders must be disabled by env")
	}
	if _, ok := cfg.Jobs["unknown"]; ok {
		t.Error("unknown job must not be created from env")
	}
}

func TestApplyEnvOverrides_EnabledTrueVariants(t *testing.T) {
	cfg := &Config{Jobs: map[string]JobConfig{
		"a": {Enabled: false},
		"b": {Enabled: false},
	}}
	applyEnvOverrides(cfg, []string{
		"SCHEDULER_A_ENABLED=true",
		"SCHEDULER_B_ENABLED=1",
	})
	if !cfg.Jobs["a"].Enabled || !cfg.Jobs["b"].Enabled {
		t.Error("enabled must accept true and 1")
	}
}

func TestParseConfig_EnvOverrideAfterValidate(t *testing.T) {
	// ENV-override применяется ДО валидации — если в YAML schedule
	// валиден, а ENV ломает его, должна вернуться ошибка.
	t.Setenv("SCHEDULER_X_CRON", "garbage")
	yaml := `
jobs:
  x:
    schedule: "0 9 * * *"
    enabled: true
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected validation to fail after env-override breaks schedule")
	}
}
