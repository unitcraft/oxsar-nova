package auth

import (
	"testing"
	"time"
)

func TestVacationMinInterval(t *testing.T) {
	t.Parallel()
	if VacationMinInterval != 20*24*time.Hour {
		t.Fatalf("VacationMinInterval = %v, want 20 days", VacationMinInterval)
	}
}

func TestVacationErrors_NotNil(t *testing.T) {
	t.Parallel()
	if ErrVacationAlreadyActive == nil {
		t.Fatal("ErrVacationAlreadyActive must not be nil")
	}
	if ErrVacationNotActive == nil {
		t.Fatal("ErrVacationNotActive must not be nil")
	}
	if ErrVacationIntervalNotMet == nil {
		t.Fatal("ErrVacationIntervalNotMet must not be nil")
	}
}
