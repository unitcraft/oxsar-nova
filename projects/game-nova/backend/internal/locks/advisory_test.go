package locks

import (
	"context"
	"testing"
)

func TestHashLockName_Stable(t *testing.T) {
	if hashLockName("alien_spawn") != hashLockName("alien_spawn") {
		t.Error("hash should be deterministic")
	}
}

func TestHashLockName_Distinct(t *testing.T) {
	a := hashLockName("alien_spawn")
	b := hashLockName("inactivity_reminders")
	if a == b {
		t.Errorf("different names should hash to different keys: %d == %d", a, b)
	}
}

func TestTryRun_NilPool(t *testing.T) {
	acquired, err := TryRun(context.Background(), nil, "x", func(ctx context.Context) error { return nil })
	if acquired {
		t.Error("expected acquired=false")
	}
	if err == nil {
		t.Error("expected error for nil pool")
	}
}

func TestTryRun_EmptyName(t *testing.T) {
	// Pool можно не передавать — проверка имени идёт первой.
	acquired, err := TryRun(context.Background(), nil, "", func(ctx context.Context) error { return nil })
	if acquired {
		t.Error("expected acquired=false")
	}
	if err == nil {
		t.Error("expected error for empty name")
	}
}
