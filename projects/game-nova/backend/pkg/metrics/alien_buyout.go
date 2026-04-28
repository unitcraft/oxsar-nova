// Метрики платного выкупа удержания пришельцами оксарами
// (план 66 Ф.5, R8).
//
// Отдельный файл, чтобы не попадать в шаблон metrics.go,
// который дублируется между сервисами oxsar-nova/identity/portal/billing.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	alienBuyoutOnce sync.Once

	// AlienBuyoutTotal — счётчик попыток buyout по статусу.
	// Labels:
	//   status — ok|insufficient|conflict|not_found|billing_unavailable|error.
	AlienBuyoutTotal *prometheus.CounterVec

	// AlienBuyoutOxsars — суммарно списанные оксары (только ok-ветка).
	// Counter, не Gauge: значение монотонно растёт. Sum() даёт итог
	// рублёвой выручки по фиче.
	AlienBuyoutOxsars prometheus.Counter
)

// RegisterAlienBuyout регистрирует метрики buyout в default-Registerer.
// Идемпотентно. Вызывается автоматически из metrics.Register, но может быть
// вызван явно из тестов / cmd/tools.
func RegisterAlienBuyout() {
	alienBuyoutOnce.Do(func() {
		AlienBuyoutTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "alien_buyout",
			Name:      "total",
			Help:      "Alien holding buyout attempts by status (ok|insufficient|conflict|not_found|billing_unavailable|error).",
		}, []string{"status"})

		AlienBuyoutOxsars = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "alien_buyout",
			Name:      "oxsars_total",
			Help:      "Total oxsars spent on alien holding buyout (ok branch only).",
		})

		prometheus.MustRegister(AlienBuyoutTotal, AlienBuyoutOxsars)
	})
}
