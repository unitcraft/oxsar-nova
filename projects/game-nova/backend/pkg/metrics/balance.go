// Метрики per-universe balance loader (план 64, R8).
//
// Отдельный файл, чтобы не попадать в шаблон metrics.go,
// который дублируется между сервисами oxsar-nova/identity/portal/billing.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	balanceOnce sync.Once

	// BalanceLoadTotal — счётчик загрузок bundle. Labels:
	//   universe — universe.id (или "" для default)
	//   status   — ok|error
	//   override — true|false (был ли применён override-файл)
	BalanceLoadTotal *prometheus.CounterVec

	// BalanceLoadDuration — длительность LoadFor (с учётом первого
	// LoadDefaults и парсинга override).
	BalanceLoadDuration *prometheus.HistogramVec
)

// RegisterBalance регистрирует balance-метрики в default-Registerer.
// Идемпотентно. Вызывается автоматически из metrics.Register, но может
// быть вызван явно из тестов / cmd/tools.
func RegisterBalance() {
	balanceOnce.Do(func() {
		BalanceLoadTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "balance",
			Name:      "load_total",
			Help:      "Balance bundle loads by universe, status (ok|error) and override applied.",
		}, []string{"universe", "status", "override"})

		BalanceLoadDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "oxsar",
			Subsystem: "balance",
			Name:      "load_duration_seconds",
			Help:      "Balance bundle load duration (catalog read + override merge).",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		}, []string{"universe"})

		prometheus.MustRegister(BalanceLoadTotal, BalanceLoadDuration)
	})
}
