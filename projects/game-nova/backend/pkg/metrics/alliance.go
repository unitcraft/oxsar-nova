// Метрики alliance-actions (план 67 Ф.2, R8).
//
// Отдельный файл, чтобы не попадать в шаблон metrics.go,
// который дублируется между сервисами oxsar-nova/identity/portal/billing.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	allianceOnce sync.Once

	// AllianceActions — счётчик событий, влияющих на состояние альянса.
	// Labels:
	//   action — символическое имя (alliance_created, member_kicked,
	//            description_changed, rank_created, relation_proposed, …),
	//            совпадает с alliance.Action* константами;
	//   status — ok|forbidden|error.
	AllianceActions *prometheus.CounterVec
)

// RegisterAlliance регистрирует alliance-метрики в default-Registerer.
// Идемпотентно. Вызывается автоматически из metrics.Register, но может
// быть вызван явно из тестов / cmd/tools.
func RegisterAlliance() {
	allianceOnce.Do(func() {
		AllianceActions = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "oxsar",
			Subsystem: "alliance",
			Name:      "actions_total",
			Help:      "Alliance-affecting actions by action name and status (ok|forbidden|error).",
		}, []string{"action", "status"})

		prometheus.MustRegister(AllianceActions)
	})
}
