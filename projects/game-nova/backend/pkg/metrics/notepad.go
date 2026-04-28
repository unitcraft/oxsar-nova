// Метрики notepad endpoint'ов (план 69 Ф.6 пост-фикс, R8).
//
// Отдельный файл, чтобы не попадать в шаблон metrics.go,
// который дублируется между сервисами oxsar-nova/identity/portal/billing.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	notepadOnce sync.Once

	// NotepadActions — счётчик действий с notepad. Labels:
	//   action — get|save
	//   status — ok|unauthorized|bad_request|too_long|error
	NotepadActions *prometheus.CounterVec

	// NotepadDuration — длительность handler'а в секундах.
	NotepadDuration *prometheus.HistogramVec
)

// RegisterNotepad регистрирует notepad-метрики в default-Registerer.
// Идемпотентно. Вызывается из metrics.Register.
func RegisterNotepad() {
	notepadOnce.Do(func() {
		NotepadActions = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "notepad",
			Name:      "actions_total",
			Help:      "Notepad endpoint actions by action (get|save) and status.",
		}, []string{"action", "status"})

		NotepadDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "oxsar",
			Subsystem: "notepad",
			Name:      "duration_seconds",
			Help:      "Notepad endpoint handler duration.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		}, []string{"action"})

		prometheus.MustRegister(NotepadActions, NotepadDuration)
	})
}
