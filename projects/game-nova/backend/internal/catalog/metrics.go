package catalog

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsOnce sync.Once
	requestsTotal *prometheus.CounterVec
)

// initMetrics регистрирует Prometheus метрики catalog-домена.
// Идемпотентно (sync.Once). При проде регистрация делается через
// импортирующий пакет (cmd/server/main.go вызывает metrics.Register()
// который пробрасывает /metrics handler). Здесь лишь добавляем
// собственные коллекторы в DefaultRegisterer.
func initMetrics() {
	metricsOnce.Do(func() {
		requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "catalog",
			Name:      "requests_total",
			Help:      "Total catalog endpoint requests by kind (building|ship|defense|research|artefact|unit) and status (ok|not_found).",
		}, []string{"kind", "status"})
		prometheus.MustRegister(requestsTotal)
	})
}

// incCatalogReq увеличивает счётчик catalog запросов. Лениво
// инициализирует метрики при первом вызове.
func incCatalogReq(kind, status string) {
	initMetrics()
	requestsTotal.WithLabelValues(kind, status).Inc()
}
