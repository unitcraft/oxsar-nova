package goal

import (
	"testing"
	"time"
)

func TestPeriodKey(t *testing.T) {
	t1 := time.Date(2026, 4, 26, 14, 30, 0, 0, time.UTC)
	tests := []struct {
		lc   Lifecycle
		want string
	}{
		{LifecyclePermanent, ""},
		{LifecycleOneTime, ""},
		{LifecycleSeasonal, ""},
		{LifecycleRepeatable, ""},
		{LifecycleDaily, "2026-04-26"},
		{LifecycleWeekly, "2026-W17"}, // ISO week 17 of 2026
	}
	for _, tt := range tests {
		got := PeriodKey(tt.lc, t1)
		if got != tt.want {
			t.Errorf("PeriodKey(%s) = %q, want %q", tt.lc, got, tt.want)
		}
	}
}

func TestPeriodKey_UTCNormalization(t *testing.T) {
	// Local time в +12 часовой зоне: одна и та же UTC дата.
	zone := time.FixedZone("UTC+12", 12*3600)
	t1 := time.Date(2026, 4, 27, 1, 0, 0, 0, zone) // в UTC = 2026-04-26 13:00
	got := PeriodKey(LifecycleDaily, t1)
	if got != "2026-04-26" {
		t.Errorf("expected UTC normalisation, got %q", got)
	}
}
