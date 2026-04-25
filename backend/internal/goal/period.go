package goal

import (
	"fmt"
	"time"
)

// PeriodKey возвращает period_key для goal на момент now.
//
//   permanent / one-time / seasonal — ''
//   daily                            — 'YYYY-MM-DD' (UTC)
//   weekly                           — 'YYYY-Www' (ISO week, UTC)
//   repeatable                       — '' (зарезервировано)
//
// Используется при INSERT/UPDATE goal_progress, чтобы daily/weekly
// автоматически создавал новую строку при смене дня/недели.
func PeriodKey(lc Lifecycle, now time.Time) string {
	now = now.UTC()
	switch lc {
	case LifecycleDaily:
		return now.Format("2006-01-02")
	case LifecycleWeekly:
		year, week := now.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	default:
		return ""
	}
}
