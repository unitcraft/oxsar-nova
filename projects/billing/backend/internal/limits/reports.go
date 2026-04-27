package limits

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"time"
)

// RevenueBucket — одна точка revenue-отчёта.
type RevenueBucket struct {
	Period   string  // YYYY-MM-DD (day) или YYYY-MM (month)
	TotalKop int64
	Count    int64
	AvgRub   float64 // средний чек в рублях, округление до копейки
}

// QueryRevenueBuckets — агрегация payment_orders по периодам.
//
// granularity: "day" → DATE(paid_at), "month" → to_char('YYYY-MM').
// from/to — uniform полуоткрытый интервал [from, to).
//
// Только status='paid'. Возвращает упорядоченный по periоду слайс.
func (s *Service) QueryRevenueBuckets(ctx context.Context, from, to time.Time, granularity string) ([]RevenueBucket, error) {
	var dateExpr string
	switch granularity {
	case "day":
		dateExpr = "to_char(paid_at AT TIME ZONE $3, 'YYYY-MM-DD')"
	case "month":
		dateExpr = "to_char(paid_at AT TIME ZONE $3, 'YYYY-MM')"
	default:
		return nil, fmt.Errorf("invalid granularity %q", granularity)
	}
	tzName := s.cfg.Timezone.String()
	q := fmt.Sprintf(`
		SELECT %s AS period,
		       SUM(amount_kop) AS total_kop,
		       COUNT(*) AS cnt
		FROM payment_orders
		WHERE status = 'paid' AND paid_at >= $1 AND paid_at < $2
		GROUP BY period
		ORDER BY period
	`, dateExpr)
	rows, err := s.pool.Query(ctx, q, from, to, tzName)
	if err != nil {
		return nil, fmt.Errorf("query revenue: %w", err)
	}
	defer rows.Close()
	var out []RevenueBucket
	for rows.Next() {
		var b RevenueBucket
		if err := rows.Scan(&b.Period, &b.TotalKop, &b.Count); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		if b.Count > 0 {
			b.AvgRub = float64(b.TotalKop) / float64(b.Count) / 100.0
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// StreamPaymentsCSV — выгружает все paid-платежи за период в CSV-формате
// (RFC 4180-ish: кавычки для строк с запятыми/кавычками/переносами).
//
// Колонки: payment_id, user_id, amount_rub, paid_at_iso, provider, package_id.
//
// Streaming: читаем pgx.Rows по одной записи и пишем в io.Writer без
// промежуточного буфера. Подходит для больших выгрузок (бухгалтерия).
func (s *Service) StreamPaymentsCSV(ctx context.Context, w io.Writer, from, to time.Time) error {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, amount_kop, paid_at, provider, package_id
		FROM payment_orders
		WHERE status = 'paid' AND paid_at >= $1 AND paid_at < $2
		ORDER BY paid_at
	`, from, to)
	if err != nil {
		return fmt.Errorf("query payments: %w", err)
	}
	defer rows.Close()
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{
		"payment_id", "user_id", "amount_rub", "paid_at", "provider", "package_id",
	}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	count := 0
	for rows.Next() {
		var id, userID, provider, packageID string
		var amountKop int64
		var paidAt time.Time
		if err := rows.Scan(&id, &userID, &amountKop, &paidAt, &provider, &packageID); err != nil {
			slog.ErrorContext(ctx, "csv scan failed", slog.String("err", err.Error()))
			return fmt.Errorf("scan: %w", err)
		}
		amountRub := strconv.FormatFloat(float64(amountKop)/100.0, 'f', 2, 64)
		if err := cw.Write([]string{
			id, userID, amountRub, paidAt.UTC().Format(time.RFC3339), provider, packageID,
		}); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
		count++
		if count%1000 == 0 {
			cw.Flush()
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows err: %w", err)
	}
	return nil
}
