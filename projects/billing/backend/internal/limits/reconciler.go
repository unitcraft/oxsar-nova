package limits

import (
	"context"
	"log/slog"
	"time"
)

// Notifier — куда отправляются soft-warning alerts. На MVP реализация
// пишет в slog (структурный warn). Когда план 57 (mail-service) будет
// готов — добавится email-notifier; интерфейс не меняется.
type Notifier interface {
	NotifyThresholdReached(ctx context.Context, t Threshold, revenueKop, hardStopKop int64)
	NotifyAutoDisabled(ctx context.Context, revenueKop, hardStopKop int64)
}

// MetricsHook — Prometheus gauges, обновляются reconciler'ом.
// Реализуется в pkg/metrics. Nil-friendly (если nil — no-op).
type MetricsHook interface {
	SetRevenueYTDKop(kop int64)
	SetPaymentsDisabled(disabled bool)
	SetHardStopKop(kop int64)
}

// Reconciler — 15-минутный loop пересчитывает revenue_ytd и:
//   1. Эмитит метрики.
//   2. При revenue >= HARD_STOP — AutoDisable (если ещё не выключено).
//   3. При пересечении 80/90/95% — отправляет alert через Notifier
//      (один раз per year via MarkAlerted).
//
// Никаких advisory-locks: для MVP billing работает в одном инстансе.
// Multi-instance поддержка — отдельная задача (PG advisory lock как в
// game-nova/scheduler).
type Reconciler struct {
	svc      *Service
	notifier Notifier
	metrics  MetricsHook
	interval time.Duration
}

// NewReconciler — interval=15m по плану 54.
func NewReconciler(svc *Service, notifier Notifier, metrics MetricsHook, interval time.Duration) *Reconciler {
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	return &Reconciler{svc: svc, notifier: notifier, metrics: metrics, interval: interval}
}

// Run запускает loop до отмены ctx. Должен вызываться через `go`.
func (r *Reconciler) Run(ctx context.Context) {
	t := time.NewTicker(r.interval)
	defer t.Stop()
	// Первый прогон через 5s после старта (чтобы дать БД прогреться
	// и обновить метрики при старте сервиса).
	first := time.NewTimer(5 * time.Second)
	defer first.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-first.C:
			r.runOnce(ctx)
		case <-t.C:
			r.runOnce(ctx)
		}
	}
}

// runOnce — один проход reconciliation. Используется в тестах и из Run.
//
// Логика:
//   1. Считает revenue_ytd.
//   2. Эмитит metrics (revenue, hard_stop, disabled).
//   3. Определяет highest passed threshold (80/90/95/hard).
//   4. Если ThresholdHard — AutoDisable + NotifyAutoDisabled (если не
//      было раньше в этом году).
//   5. Иначе — для прошедшего порога вызывает MarkAlerted + Notify
//      (если возвращает true → ещё не было в этом году).
func (r *Reconciler) runOnce(ctx context.Context) {
	revenue, err := r.svc.GetRevenueYTD(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "reconciler revenue query failed",
			slog.String("err", err.Error()))
		return
	}
	hardStop := r.svc.cfg.HardStopKop

	if r.metrics != nil {
		r.metrics.SetRevenueYTDKop(revenue)
		r.metrics.SetHardStopKop(hardStop)
	}

	threshold := HighestPassed(revenue, hardStop)
	year := time.Now().In(r.svc.cfg.Timezone).Year()

	if threshold == ThresholdHard {
		// 1) Auto-disable (idempotent).
		if err := r.svc.AutoDisable(ctx, "auto: revenue ytd >= hard stop threshold"); err != nil {
			slog.ErrorContext(ctx, "auto-disable failed",
				slog.String("err", err.Error()))
		} else if r.metrics != nil {
			r.metrics.SetPaymentsDisabled(true)
		}
		// 2) Alert один раз per year.
		fresh, err := r.svc.MarkAlerted(ctx, year, ThresholdHard)
		if err != nil {
			slog.ErrorContext(ctx, "mark alerted hard failed",
				slog.String("err", err.Error()))
		} else if fresh && r.notifier != nil {
			r.notifier.NotifyAutoDisabled(ctx, revenue, hardStop)
		}
		return
	}

	// Soft-warning: 80/90/95. Если порог пройден — отметить и оповестить
	// (Notifier == nil → пропустить уведомление, но в audit/state всё равно
	// пишем).
	if threshold == ThresholdNone {
		// Никакого порога не пройдено: только метрики обновили.
		if r.metrics != nil {
			// Текущий disabled-флаг считаем из state, чтобы метрика была
			// корректной даже после admin override.
			active, _ := r.svc.IsActive(ctx)
			r.metrics.SetPaymentsDisabled(!active)
		}
		return
	}

	// Если соответствующий ENV-флаг отключён — пропускаем уведомление,
	// но MarkAlerted всё равно вызываем, чтобы при включении флага не
	// шлались задним числом.
	enabled := warnEnabled(r.svc.cfg, threshold)
	fresh, err := r.svc.MarkAlerted(ctx, year, threshold)
	if err != nil {
		slog.ErrorContext(ctx, "mark alerted failed",
			slog.String("threshold", thresholdName(threshold)),
			slog.String("err", err.Error()))
		return
	}
	if fresh && enabled && r.notifier != nil {
		r.notifier.NotifyThresholdReached(ctx, threshold, revenue, hardStop)
	}

	if r.metrics != nil {
		active, _ := r.svc.IsActive(ctx)
		r.metrics.SetPaymentsDisabled(!active)
	}
}

func warnEnabled(cfg Config, t Threshold) bool {
	switch t {
	case Threshold80:
		return cfg.WarnAt80
	case Threshold90:
		return cfg.WarnAt90
	case Threshold95:
		return cfg.WarnAt95
	}
	return true
}

func thresholdName(t Threshold) string {
	switch t {
	case Threshold80:
		return "80%"
	case Threshold90:
		return "90%"
	case Threshold95:
		return "95%"
	case ThresholdHard:
		return "hard_stop"
	}
	return "none"
}
