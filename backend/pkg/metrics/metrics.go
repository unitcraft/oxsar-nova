// Package metrics — Prometheus-метрики oxsar-nova.
//
// Экспорт на /metrics. Worker и server используют общий набор
// метрик через вызовы EventProcessed/EventDuration/... Метрики
// регистрируются лениво при первом вызове пакетной функции Register.
package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once sync.Once

	EventsProcessed *prometheus.CounterVec
	EventHandlerSec *prometheus.HistogramVec
	EventsQueue     *prometheus.GaugeVec
	EventsLagSec    prometheus.Gauge
)

// Register инициализирует все метрики и возвращает http.Handler для
// экспорта на /metrics. Идемпотентно (sync.Once).
func Register() http.Handler {
	once.Do(func() {
		EventsProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "events",
			Name:      "processed_total",
			Help:      "Total events processed by state (ok|error|skip|retry) and kind.",
		}, []string{"kind", "state"})

		EventHandlerSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "oxsar",
			Subsystem: "events",
			Name:      "handler_duration_seconds",
			Help:      "Event handler duration by kind.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		}, []string{"kind"})

		EventsQueue = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "oxsar",
			Subsystem: "events",
			Name:      "queue_depth",
			Help:      "Current event queue depth by state.",
		}, []string{"state"})

		EventsLagSec = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "oxsar",
			Subsystem: "events",
			Name:      "lag_seconds",
			Help:      "Age of the oldest wait event with fire_at<=now.",
		})

		prometheus.MustRegister(EventsProcessed, EventHandlerSec, EventsQueue, EventsLagSec)
	})
	return promhttp.Handler()
}
