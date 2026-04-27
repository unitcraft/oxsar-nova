package billing

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics для billing-service. План 38 §9 Observability.
//
// Конвенция оксаровская: <service>_<unit>_<суффикс>.
// Подключаются автоматически через promauto в default registry.
var (
	// TransactionsTotal — счётчик всех транзакций по reason+type.
	// Лейблы:
	//   reason — 'top_up' | 'feedback_vote' | 'shop_purchase' | 'refund' | ...
	//   type   — 'spend' | 'credit'
	TransactionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_transactions_total",
		Help: "Total number of wallet transactions by reason and type.",
	}, []string{"reason", "type"})

	// SpendErrorsTotal — счётчик ошибок Spend (insufficient, frozen, internal).
	SpendErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_spend_errors_total",
		Help: "Total number of failed Spend attempts by reason.",
	}, []string{"error"})

	// WebhooksTotal — счётчик webhook'ов от платёжных шлюзов.
	// Лейблы: provider, status — 'ok' | 'invalid_signature' | 'expired' | 'replay' | 'error'.
	WebhooksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_webhooks_total",
		Help: "Total number of webhook calls by provider and status.",
	}, []string{"provider", "status"})

	// ReconcileChecks — gauge: сколько кошельков проверено в последнем прогоне.
	ReconcileChecks = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "billing_reconcile_checks_last",
		Help: "Wallets checked in last reconcile run.",
	})

	// ReconcileMismatches — gauge: сколько расхождений в последнем прогоне.
	// Алерт: > 0 в течение N минут — paging on-call.
	ReconcileMismatches = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "billing_reconcile_mismatches_last",
		Help: "Wallet balance mismatches in last reconcile run (target=0).",
	})

	// ReconcileErrorsTotal — счётчик ошибок reconcile (DB-проблемы и т.п.).
	ReconcileErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "billing_reconcile_errors_total",
		Help: "Total reconcile errors (DB issues, etc).",
	})

	// WalletsFrozenTotal — счётчик заморозок (cumulative). Алерт на rate > 0.
	WalletsFrozenTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "billing_wallets_frozen_total",
		Help: "Total number of wallets frozen by reconcile (cumulative).",
	})
)
