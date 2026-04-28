// Метрики billing-client (план 77, R8).
//
// Отдельный файл, чтобы не попадать в шаблон metrics.go,
// который дублируется между сервисами oxsar-nova/identity/portal/billing.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	billingOnce sync.Once

	// BillingClientSpend — счётчик вызовов billing-client.
	// Labels:
	//   operation — spend|refund.
	//   status    — ok|insufficient|conflict|frozen|unavailable|error.
	BillingClientSpend *prometheus.CounterVec

	// BillingClientDuration — длительность вызова billing-client (включая
	// retry на транзиентных ошибках).
	// Labels:
	//   operation — spend|refund.
	BillingClientDuration *prometheus.HistogramVec
)

// RegisterBilling регистрирует billing-client-метрики в default-Registerer.
// Идемпотентно. Вызывается автоматически из metrics.Register, но может быть
// вызван явно из тестов / cmd/tools.
func RegisterBilling() {
	billingOnce.Do(func() {
		BillingClientSpend = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "billing_client",
			Name:      "spend_total",
			Help:      "Billing client spend/refund calls by operation and status (ok|insufficient|conflict|frozen|unavailable|error).",
		}, []string{"operation", "status"})

		BillingClientDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "oxsar",
			Subsystem: "billing_client",
			Name:      "duration_seconds",
			Help:      "Billing client call duration (including retries) by operation.",
			Buckets:   []float64{0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
		}, []string{"operation"})

		prometheus.MustRegister(BillingClientSpend, BillingClientDuration)
	})
}
