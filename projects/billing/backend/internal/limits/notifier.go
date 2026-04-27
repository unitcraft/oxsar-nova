package limits

import (
	"context"
	"log/slog"
)

// SlogNotifier — реализация Notifier, которая пишет structured warn в
// slog. По плану 54 это MVP: когда план 57 (mail-service) будет готов,
// добавится EmailNotifier и/или composite. Также Prometheus alert rules
// в deploy/prometheus/alerts.yml ловят metrics из MetricsHook независимо
// от Notifier'а.
type SlogNotifier struct {
	log *slog.Logger
}

// NewSlogNotifier — log==nil допустимо, тогда используется slog.Default().
func NewSlogNotifier(log *slog.Logger) *SlogNotifier {
	if log == nil {
		log = slog.Default()
	}
	return &SlogNotifier{log: log}
}

func (n *SlogNotifier) NotifyThresholdReached(ctx context.Context, t Threshold, revenueKop, hardStopKop int64) {
	n.log.WarnContext(ctx, "billing limit threshold reached",
		slog.String("threshold", thresholdName(t)),
		slog.Int64("revenue_ytd_kop", revenueKop),
		slog.Int64("hard_stop_kop", hardStopKop),
		slog.Int64("revenue_ytd_rub", revenueKop/100),
		slog.Int64("hard_stop_rub", hardStopKop/100),
		slog.String("event", "billing_threshold_reached"),
	)
}

func (n *SlogNotifier) NotifyAutoDisabled(ctx context.Context, revenueKop, hardStopKop int64) {
	n.log.ErrorContext(ctx, "billing payments AUTO-DISABLED (hard stop)",
		slog.Int64("revenue_ytd_kop", revenueKop),
		slog.Int64("hard_stop_kop", hardStopKop),
		slog.Int64("revenue_ytd_rub", revenueKop/100),
		slog.Int64("hard_stop_rub", hardStopKop/100),
		slog.String("event", "billing_auto_disabled"),
		slog.String("action_required", "review revenue and decide whether to enable manually"),
	)
}
