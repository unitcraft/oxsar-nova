// Метрики KindTeleportPlanet (план 65 Ф.6, R8).
//
// Отдельный файл, чтобы не попадать в шаблон metrics.go,
// который дублируется между сервисами oxsar-nova/identity/portal/billing.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	teleportOnce sync.Once

	// PlanetTeleportTotal — счётчик попыток телепорта планеты.
	// Labels:
	//   status — ok|invalid_coords|cooldown|occupied|insufficient|conflict|billing_unavailable|error.
	PlanetTeleportTotal *prometheus.CounterVec

	// PlanetTeleportDuration — длительность HTTP-handler'а POST /api/planets/{id}/teleport
	// (включая вызов billing-client'а и запись event).
	PlanetTeleportDuration prometheus.Histogram
)

// RegisterTeleport регистрирует метрики KindTeleportPlanet в default-Registerer.
// Идемпотентно. Вызывается автоматически из metrics.Register, но может быть
// вызван явно из тестов / cmd/tools.
func RegisterTeleport() {
	teleportOnce.Do(func() {
		PlanetTeleportTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "planet",
			Name:      "teleport_total",
			Help:      "Planet teleport attempts by status (ok|invalid_coords|cooldown|occupied|insufficient|conflict|billing_unavailable|error).",
		}, []string{"status"})

		PlanetTeleportDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "oxsar",
			Subsystem: "planet",
			Name:      "teleport_duration_seconds",
			Help:      "HTTP handler duration for POST /api/planets/{id}/teleport.",
			Buckets:   []float64{0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
		})

		prometheus.MustRegister(PlanetTeleportTotal, PlanetTeleportDuration)
	})
}
