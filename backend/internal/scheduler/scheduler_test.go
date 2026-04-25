package scheduler

import (
	"context"
	"errors"
	"testing"

	"github.com/oxsar/nova/backend/pkg/metrics"
)

func newTestConfig(t *testing.T) *Config {
	t.Helper()
	cfg, err := ParseConfig([]byte(`
jobs:
  ok_job:
    schedule: "0 9 * * *"
    enabled: true
    description: "test"
  off_job:
    schedule: "0 9 * * *"
    enabled: false
    description: "disabled"
`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return cfg
}

func TestRegister_UnknownJob(t *testing.T) {
	cfg := newTestConfig(t)
	s := New(cfg, nil, nil)
	err := s.Register("missing_in_config", func(ctx context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected error for job not in config")
	}
}

func TestRegister_DisabledJob(t *testing.T) {
	cfg := newTestConfig(t)
	s := New(cfg, nil, nil)
	if err := s.Register("off_job", func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("disabled job register: %v", err)
	}
	if names := s.JobNames(); len(names) != 0 {
		t.Errorf("disabled job must not be in JobNames: %v", names)
	}
}

func TestRegister_NilFn(t *testing.T) {
	cfg := newTestConfig(t)
	s := New(cfg, nil, nil)
	if err := s.Register("ok_job", nil); err == nil {
		t.Fatal("expected error for nil fn")
	}
}

func TestRegister_EmptyName(t *testing.T) {
	cfg := newTestConfig(t)
	s := New(cfg, nil, nil)
	if err := s.Register("", func(ctx context.Context) error { return nil }); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegister_OkPath(t *testing.T) {
	cfg := newTestConfig(t)
	s := New(cfg, nil, nil)
	if err := s.Register("ok_job", func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("register: %v", err)
	}
	names := s.JobNames()
	if len(names) != 1 || names[0] != "ok_job" {
		t.Errorf("JobNames = %v, want [ok_job]", names)
	}
}

func TestRunJob_NilPoolIsTreatedAsError(t *testing.T) {
	// Без БД locks.TryRun вернёт error+acquired=false. Проверяем, что
	// runJob не падает и идёт в путь "pre-lock error" (не "skip").
	metrics.Register() // регистрируем метрики, иначе Inc() не словит nil-pointer
	cfg := newTestConfig(t)
	s := New(cfg, nil, nil)

	called := false
	s.runJob(context.Background(), "ok_job", func(ctx context.Context) error {
		called = true
		return nil
	})
	if called {
		t.Error("fn must not be called when lock acquire fails")
	}
}

func TestStop_NoJobs(t *testing.T) {
	cfg := newTestConfig(t)
	s := New(cfg, nil, nil)
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Errorf("stop: %v", err)
	}
}

func TestDisableAll(t *testing.T) {
	t.Setenv("SCHEDULER_DISABLE_ALL", "true")
	cfg := newTestConfig(t)
	s := New(cfg, nil, nil)
	if err := s.Register("ok_job", func(ctx context.Context) error { return errors.New("boom") }); err != nil {
		t.Fatalf("register on disabled scheduler must be no-op, got: %v", err)
	}
	if names := s.JobNames(); len(names) != 0 {
		t.Errorf("disabled scheduler must have 0 jobs, got %v", names)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Errorf("start on disabled: %v", err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Errorf("stop on disabled: %v", err)
	}
}
