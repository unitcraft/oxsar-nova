package billing

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Reconciler сверяет сумму транзакций с балансом каждого кошелька.
// План 38 §Reconciliation.
//
// Идея: wallet.balance — материализованная сумма SUM(transactions.delta).
// Если они расходятся — что-то пошло не так (баг, ручной UPDATE через psql,
// hardware corruption). В этом случае:
//   1. Логируем ошибку (структурно).
//   2. Замораживаем кошелёк (frozen=true) — все списания/пополнения
//      возвращают 423 Locked.
//   3. Алерт в Prometheus → Grafana → on-call.
//
// Запускается как background-goroutine с заданным интервалом (по умолчанию
// каждый час). Проверка идёт батчами — не блокирует обычные операции.
type Reconciler struct {
	svc      *Service
	interval time.Duration
}

func NewReconciler(svc *Service, interval time.Duration) *Reconciler {
	if interval <= 0 {
		interval = time.Hour
	}
	return &Reconciler{svc: svc, interval: interval}
}

// Run запускает цикл reconcile до отмены ctx.
// Должен вызываться через `go reconciler.Run(ctx)`.
func (r *Reconciler) Run(ctx context.Context) {
	t := time.NewTicker(r.interval)
	defer t.Stop()
	// Первый прогон сразу на старте (после 5s, чтобы дать БД прогреться).
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

// ReconcileResult — результат одного прогона.
type ReconcileResult struct {
	Checked    int   // сколько кошельков проверили
	Mismatched int   // у скольких баланс не совпал с SUM(transactions.delta)
	Frozen     int   // скольких заморозили
}

// runOnce выполняет один проход reconcile. Можно вызывать вручную из тестов.
func (r *Reconciler) runOnce(ctx context.Context) ReconcileResult {
	res := ReconcileResult{}
	rows, err := r.svc.db.Pool().Query(ctx, `
		SELECT w.id, w.user_id, w.balance, w.frozen,
		       COALESCE(SUM(t.delta), 0) AS expected
		FROM wallets w
		LEFT JOIN transactions t ON t.wallet_id = w.id
		GROUP BY w.id
	`)
	if err != nil {
		slog.ErrorContext(ctx, "reconcile query failed",
			slog.String("err", err.Error()))
		ReconcileErrorsTotal.Inc()
		return res
	}
	defer rows.Close()

	type mismatch struct {
		walletID string
		userID   string
		actual   int64
		expected int64
	}
	var bad []mismatch
	for rows.Next() {
		var walletID, userID string
		var balance, expected int64
		var frozen bool
		if err := rows.Scan(&walletID, &userID, &balance, &frozen, &expected); err != nil {
			slog.ErrorContext(ctx, "reconcile scan failed",
				slog.String("err", err.Error()))
			continue
		}
		res.Checked++
		if balance != expected && !frozen {
			bad = append(bad, mismatch{walletID, userID, balance, expected})
		}
	}
	res.Mismatched = len(bad)

	// Замораживаем найденные. По одной транзакции на кошелёк, чтобы один
	// плохой не блокировал других.
	for _, m := range bad {
		reason := fmt.Sprintf(
			"reconcile mismatch: balance=%d, expected SUM(delta)=%d",
			m.actual, m.expected)
		if _, err := r.svc.db.Pool().Exec(ctx, `
			UPDATE wallets SET frozen = true, frozen_reason = $1, updated_at = now()
			WHERE id = $2
		`, reason, m.walletID); err != nil {
			slog.ErrorContext(ctx, "reconcile freeze failed",
				slog.String("wallet_id", m.walletID),
				slog.String("err", err.Error()))
			ReconcileErrorsTotal.Inc()
			continue
		}
		res.Frozen++
		WalletsFrozenTotal.Inc()
		slog.WarnContext(ctx, "wallet frozen due to reconcile mismatch",
			slog.String("wallet_id", m.walletID),
			slog.String("user_id", m.userID),
			slog.Int64("balance", m.actual),
			slog.Int64("expected", m.expected))
	}
	ReconcileChecks.Set(float64(res.Checked))
	ReconcileMismatches.Set(float64(res.Mismatched))
	slog.InfoContext(ctx, "reconcile done",
		slog.Int("checked", res.Checked),
		slog.Int("mismatched", res.Mismatched),
		slog.Int("frozen", res.Frozen))
	return res
}
