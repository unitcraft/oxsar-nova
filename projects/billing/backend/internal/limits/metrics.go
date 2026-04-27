package limits

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics — реализация MetricsHook поверх client_golang.
//
// Метрики (план 54 §Prometheus):
//   billing_revenue_ytd_kop      Gauge — текущий доход год-к-дате (копейки).
//   billing_hard_stop_kop        Gauge — порог auto-disable (копейки).
//   billing_payments_disabled    Gauge — 0/1 если пополнение отключено.
//
// Альерты в deploy/prometheus (план 54 Ф.6, отдельной задачей):
//   billing_revenue_ytd_kop / billing_hard_stop_kop > 0.8 → 80%
//   ... > 0.9 → 90%, > 0.95 → 95%
//   billing_payments_disabled == 1 → hard stop
type PrometheusMetrics struct {
	revenueYTD     prometheus.Gauge
	hardStop       prometheus.Gauge
	paymentsDisabled prometheus.Gauge
}

// NewPrometheusMetrics регистрирует метрики в default-registry. Вызывать
// один раз при старте сервиса (повторный вызов вызовет panic из promauto).
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		revenueYTD: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "billing_revenue_ytd_kop",
			Help: "Current year-to-date revenue in kopeks (НПД ФЗ-422 limit tracking).",
		}),
		hardStop: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "billing_hard_stop_kop",
			Help: "Hard-stop threshold in kopeks (auto-disable triggers when revenue >= this).",
		}),
		paymentsDisabled: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "billing_payments_disabled",
			Help: "1 if payments are disabled (auto-stop or admin override), 0 otherwise.",
		}),
	}
}

func (m *PrometheusMetrics) SetRevenueYTDKop(kop int64) {
	m.revenueYTD.Set(float64(kop))
}

func (m *PrometheusMetrics) SetHardStopKop(kop int64) {
	m.hardStop.Set(float64(kop))
}

func (m *PrometheusMetrics) SetPaymentsDisabled(disabled bool) {
	if disabled {
		m.paymentsDisabled.Set(1)
	} else {
		m.paymentsDisabled.Set(0)
	}
}
