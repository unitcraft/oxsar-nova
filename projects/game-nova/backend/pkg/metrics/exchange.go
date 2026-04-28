// Метрики биржи артефактов (план 68, R8).
//
// Отдельный файл, чтобы не попадать в шаблон metrics.go,
// который дублируется между сервисами oxsar-nova/identity/portal/billing.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	exchangeOnce sync.Once

	// ExchangeLotsTotal — счётчик действий по лотам.
	// Labels:
	//   action — create|buy|cancel.
	//   status — ok|insufficient|conflict|forbidden|not_found|error.
	ExchangeLotsTotal *prometheus.CounterVec

	// ExchangeOxsaritsVolume — суммарно перевод оксаритов через биржу
	// (только успешные buy). Counter, монотонно растёт.
	ExchangeOxsaritsVolume prometheus.Counter

	// ExchangeActionDuration — длительность service-методов.
	// Labels:
	//   action — create|buy|cancel|list|get|stats.
	ExchangeActionDuration *prometheus.HistogramVec

	// ExchangeEventTotal — счётчик event-handler'ов биржи.
	// Labels:
	//   kind   — expire|ban.
	//   status — ok|noop|error.
	ExchangeEventTotal *prometheus.CounterVec
)

// RegisterExchange регистрирует метрики биржи в default-Registerer.
// Идемпотентно. Вызывается из metrics.Register, но безопасно из тестов.
func RegisterExchange() {
	exchangeOnce.Do(func() {
		ExchangeLotsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "exchange",
			Name:      "lots_total",
			Help:      "Exchange lot actions by action (create|buy|cancel) and status (ok|insufficient|conflict|forbidden|not_found|error).",
		}, []string{"action", "status"})

		ExchangeOxsaritsVolume = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "exchange",
			Name:      "oxsarits_volume_total",
			Help:      "Total oxsarits transferred through successful exchange buys.",
		})

		ExchangeActionDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "oxsar",
			Subsystem: "exchange",
			Name:      "action_duration_seconds",
			Help:      "Exchange service method duration by action (create|buy|cancel|list|get|stats).",
			Buckets:   []float64{0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		}, []string{"action"})

		ExchangeEventTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "exchange",
			Name:      "event_total",
			Help:      "Exchange event-handler runs by kind (expire|ban) and status (ok|noop|error).",
		}, []string{"kind", "status"})

		prometheus.MustRegister(ExchangeLotsTotal, ExchangeOxsaritsVolume,
			ExchangeActionDuration, ExchangeEventTotal)
	})
}
