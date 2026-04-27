// Package metrics — Prometheus-метрики oxsar-nova.
//
// Экспорт на /metrics. Worker и server используют общий набор
// метрик через вызовы EventProcessed/EventDuration/... Метрики
// регистрируются лениво при первом вызове пакетной функции Register.
package metrics

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth и oxsar/portal. При любом изменении синхронизируйте КОПИИ:
//   - projects/game-nova/backend/pkg/metrics/metrics.go
//   - projects/auth/backend/pkg/metrics/metrics.go
//   - projects/portal/backend/pkg/metrics/metrics.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

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

	// Scheduler-метрики (план 32). Job — символическое имя в schedule.yaml
	// (alien_spawn, score_recalc_all, …).
	SchedulerJobRuns     *prometheus.CounterVec   // labels: job, status (ok|error|skip)
	SchedulerJobDuration *prometheus.HistogramVec // labels: job
	SchedulerJobLastRun  *prometheus.GaugeVec     // labels: job — unix-timestamp последнего запуска (любой status)
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

		SchedulerJobRuns = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "scheduler",
			Name:      "job_runs_total",
			Help:      "Scheduler job runs by name and status (ok|error|skip).",
		}, []string{"job", "status"})

		SchedulerJobDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "oxsar",
			Subsystem: "scheduler",
			Name:      "job_duration_seconds",
			Help:      "Scheduler job execution duration by name (only for non-skip runs).",
			Buckets:   []float64{0.01, 0.1, 1, 5, 30, 60, 300, 1800},
		}, []string{"job"})

		SchedulerJobLastRun = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "oxsar",
			Subsystem: "scheduler",
			Name:      "job_last_run_timestamp",
			Help:      "Unix timestamp of the last scheduler job tick (regardless of status).",
		}, []string{"job"})

		prometheus.MustRegister(EventsProcessed, EventHandlerSec, EventsQueue, EventsLagSec,
			SchedulerJobRuns, SchedulerJobDuration, SchedulerJobLastRun)
	})
	return promhttp.Handler()
}
