// Package limits — лимит самозанятого (ФЗ-422 НПД) + soft-warning
// alerts. План 54.
//
// Модель:
//   - HARD_STOP_THRESHOLD_KOP (default 230_000_000 копеек = 2.3 млн ₽,
//     буфер 100k от лимита ФНС 2.4 млн).
//   - revenue_ytd_kop = SUM(amount_kop) FROM payment_orders WHERE
//     status='paid' AND paid_at >= start_of_year(LIMIT_CHECK_TIMEZONE).
//   - При revenue_ytd >= HARD_STOP → reconciler ставит
//     billing_system_state.payments_active = false (auto-disabled).
//   - BuildPayURL перед оплатой проверяет IsActive() — если false,
//     возвращает ErrLimitReached (HTTP 503) с нейтральным сообщением.
//   - Refunds не сбрасывают флаг автоматически (yo-yo prevention) —
//     только admin override.
//
// IsActive cached in-process на 30 секунд: это даёт бэкенду переживать
// всплески BuildPayURL без N запросов в БД, а 30 сек задержка между
// reconciler-обновлением и effect — приемлема (reconciler крутится
// каждые 15 минут).
package limits

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrLimitReached — лимит самозанятого превышен, пополнение временно
// отключено. Возвращается из BuildPayURL и связанных payment-flow.
// Сообщение для клиента — нейтральное (не раскрываем причину).
var ErrLimitReached = errors.New("payments temporarily disabled")

// PublicMessage — что показывать клиенту при ErrLimitReached.
const PublicMessage = "Пополнение временно недоступно. Попробуйте позже."

// State — снимок текущего состояния лимита.
type State struct {
	// Active — true если пополнение разрешено.
	Active bool
	// RevenueYTDKop — текущий доход за год (копейки).
	RevenueYTDKop int64
	// HardStopKop — порог auto-disable (копейки).
	HardStopKop int64
	// Percent — revenue_ytd / hard_stop * 100 (для UI/мониторинга).
	Percent float64
	// LastChangedAt — когда state менялся последний раз.
	LastChangedAt *time.Time
	// LastChangedBy — кто менял (NULL для system).
	LastChangedBy *uuid.UUID
	// LastChangeReason — комментарий из admin override.
	LastChangeReason string
	// AutoDisabledAt — non-nil если reconciler выключил автоматически.
	AutoDisabledAt *time.Time
}

// Config — настраивается через ENV в main.go.
type Config struct {
	HardStopKop  int64         // default 230_000_000 (2.3 млн ₽)
	Timezone     *time.Location // для year boundary; default Europe/Moscow
	CacheTTL     time.Duration // default 30s
	WarnAt80     bool
	WarnAt90     bool
	WarnAt95     bool
}

// DefaultConfig — production defaults для самозанятого по ФЗ-422.
func DefaultConfig() Config {
	moscow, _ := time.LoadLocation("Europe/Moscow")
	if moscow == nil {
		moscow = time.UTC
	}
	return Config{
		HardStopKop: 230_000_000,
		Timezone:    moscow,
		CacheTTL:    30 * time.Second,
		WarnAt80:    true,
		WarnAt90:    true,
		WarnAt95:    true,
	}
}

// Service — основное API limits-пакета.
type Service struct {
	pool *pgxpool.Pool
	cfg  Config

	mu         sync.RWMutex
	cached     bool
	cacheValue bool
	cacheAt    time.Time
}

// New создаёт Service с заданным pool и конфигом.
func New(pool *pgxpool.Pool, cfg Config) *Service {
	if cfg.HardStopKop <= 0 {
		cfg.HardStopKop = 230_000_000
	}
	if cfg.Timezone == nil {
		cfg.Timezone = time.UTC
	}
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 30 * time.Second
	}
	return &Service{pool: pool, cfg: cfg}
}

// IsActive — быстрая проверка для BuildPayURL.
//
// Использует in-process cache на CacheTTL (default 30s). Если cache
// устарел или ещё не заполнен — читает billing_system_state.
//
// Никогда не возвращает error: при сбое БД политика fail-closed —
// возвращаем (false, error). Вызывающий код обрабатывает как «лимит
// сейчас недоступен» — это безопаснее чем разрешить платежи без
// проверки.
func (s *Service) IsActive(ctx context.Context) (bool, error) {
	s.mu.RLock()
	if s.cached && time.Since(s.cacheAt) < s.cfg.CacheTTL {
		v := s.cacheValue
		s.mu.RUnlock()
		return v, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	// Двойная проверка под write-lock'ом (другая goroutine могла обновить).
	if s.cached && time.Since(s.cacheAt) < s.cfg.CacheTTL {
		return s.cacheValue, nil
	}

	var active bool
	err := s.pool.QueryRow(ctx,
		`SELECT payments_active FROM billing_system_state WHERE id = 1`,
	).Scan(&active)
	if err != nil {
		// Fail-closed: при ошибке БД считаем не-активным.
		return false, fmt.Errorf("read billing_system_state: %w", err)
	}
	s.cached = true
	s.cacheValue = active
	s.cacheAt = time.Now()
	return active, nil
}

// invalidateCache — сбрасывает кеш (после SetActive или другого write).
func (s *Service) invalidateCache() {
	s.mu.Lock()
	s.cached = false
	s.mu.Unlock()
}

// GetState — полное состояние лимита для admin UI.
func (s *Service) GetState(ctx context.Context) (State, error) {
	st := State{HardStopKop: s.cfg.HardStopKop}
	err := s.pool.QueryRow(ctx, `
		SELECT payments_active, last_changed_by, last_changed_at,
		       COALESCE(last_change_reason, ''), auto_disabled_at
		FROM billing_system_state WHERE id = 1
	`).Scan(&st.Active, &st.LastChangedBy, &st.LastChangedAt,
		&st.LastChangeReason, &st.AutoDisabledAt)
	if err != nil {
		return st, fmt.Errorf("read state: %w", err)
	}
	revenue, err := s.GetRevenueYTD(ctx)
	if err != nil {
		return st, err
	}
	st.RevenueYTDKop = revenue
	if st.HardStopKop > 0 {
		st.Percent = float64(revenue) / float64(st.HardStopKop) * 100.0
	}
	return st, nil
}

// GetRevenueYTD — сумма успешных платежей с начала года в timezone из cfg.
//
// Источник: payment_orders.amount_kop где status='paid' AND paid_at в
// текущем году (по cfg.Timezone). Refunds в текущей схеме отдельной
// сущностью не отслеживаются (план 38 не выделяет refunds в payment_orders),
// поэтому revenue_ytd считается без вычитания возвратов. Это
// консервативно: фактический доход самозанятого по ФЗ-422 будет
// ≤ revenue_ytd, что добавляет ещё буфер сверх HARD_STOP.
//
// Когда план 38 выделит refunds — здесь добавится `- SUM(refunds)`.
func (s *Service) GetRevenueYTD(ctx context.Context) (int64, error) {
	startOfYear := startOfYear(time.Now().In(s.cfg.Timezone), s.cfg.Timezone)
	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_kop), 0)
		FROM payment_orders
		WHERE status = 'paid' AND paid_at >= $1
	`, startOfYear).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("revenue ytd query: %w", err)
	}
	return total, nil
}

// SetActive — admin override (включить/выключить пополнение вручную).
// Пишет audit-запись с reason. Сбрасывает кеш.
//
// actorID — UUID админа (из JWT). reason — обязателен.
func (s *Service) SetActive(
	ctx context.Context,
	active bool,
	actorID uuid.UUID,
	reason string,
	ip *string,
	userAgent string,
) error {
	if reason == "" {
		return errors.New("reason is required")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Ручной override снимает auto_disabled_at (флаг ставился reconciler'ом).
	_, err = tx.Exec(ctx, `
		UPDATE billing_system_state SET
			payments_active = $1,
			last_changed_by = $2,
			last_changed_at = now(),
			last_change_reason = $3,
			auto_disabled_at = CASE WHEN $1 = true THEN NULL ELSE auto_disabled_at END
		WHERE id = 1
	`, active, actorID, reason)
	if err != nil {
		return fmt.Errorf("update state: %w", err)
	}

	action := "limit:enable"
	if !active {
		action = "limit:disable"
	}
	payload := fmt.Sprintf(`{"active":%t}`, active)
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_audit_log
			(actor_id, action, target_type, target_id, payload, reason,
			 ip_address, user_agent)
		VALUES ($1, $2, 'system', 'limit', $3::jsonb, $4, $5::inet, $6)
	`, actorID, action, payload, reason, ip, userAgent)
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	s.invalidateCache()
	return nil
}

// AutoDisable — внутренний вызов из reconciler. Ставит payments_active=false
// и auto_disabled_at=now(). actor — system UUID (zero), пишется в audit.
//
// Идемпотентен: если уже выключено auto, не дублирует запись.
func (s *Service) AutoDisable(ctx context.Context, reason string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var alreadyOff bool
	err = tx.QueryRow(ctx, `
		SELECT NOT payments_active FROM billing_system_state WHERE id = 1
	`).Scan(&alreadyOff)
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}
	if alreadyOff {
		return nil
	}

	_, err = tx.Exec(ctx, `
		UPDATE billing_system_state SET
			payments_active = false,
			last_changed_by = NULL,
			last_changed_at = now(),
			last_change_reason = $1,
			auto_disabled_at = now()
		WHERE id = 1
	`, reason)
	if err != nil {
		return fmt.Errorf("update state: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_audit_log
			(actor_id, action, target_type, target_id, payload, reason)
		VALUES ($1, 'limit:auto_disable', 'system', 'limit', '{}'::jsonb, $2)
	`, uuid.Nil, reason)
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	s.invalidateCache()
	return nil
}

// Threshold — пороги soft-warning.
type Threshold int

const (
	ThresholdNone Threshold = 0
	Threshold80   Threshold = 80
	Threshold90   Threshold = 90
	Threshold95   Threshold = 95
	ThresholdHard Threshold = 100
)

// HighestPassed — какой максимальный порог пройден текущим revenue.
// Возвращает ThresholdNone если revenue < 80%.
func HighestPassed(revenueKop, hardStopKop int64) Threshold {
	if hardStopKop <= 0 {
		return ThresholdNone
	}
	pct := float64(revenueKop) / float64(hardStopKop) * 100.0
	switch {
	case pct >= 100.0:
		return ThresholdHard
	case pct >= 95.0:
		return Threshold95
	case pct >= 90.0:
		return Threshold90
	case pct >= 80.0:
		return Threshold80
	}
	return ThresholdNone
}

// MarkAlerted — фиксирует, что alert для (year, threshold) отправлен.
// INSERT row если для года ещё нет; иначе UPDATE соответствующей колонки
// (если она NULL — это первая отправка).
//
// Returns true если запись свежая (нужно слать), false если уже было.
func (s *Service) MarkAlerted(ctx context.Context, year int, t Threshold) (bool, error) {
	col := alertColumn(t)
	if col == "" {
		return false, fmt.Errorf("invalid threshold %d", t)
	}
	// 1. Гарантируем что строка для year существует.
	_, err := s.pool.Exec(ctx, `
		INSERT INTO billing_alert_state (year) VALUES ($1) ON CONFLICT DO NOTHING
	`, year)
	if err != nil {
		return false, fmt.Errorf("ensure year row: %w", err)
	}
	// 2. UPDATE column = now() WHERE column IS NULL — атомарно.
	tag, err := s.pool.Exec(ctx, fmt.Sprintf(`
		UPDATE billing_alert_state SET %s = now()
		WHERE year = $1 AND %s IS NULL
	`, col, col), year)
	if err != nil {
		return false, fmt.Errorf("mark alerted: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// AlertedAt — когда конкретный порог уже был отправлен в указанном году
// (NULL если ещё нет).
func (s *Service) AlertedAt(ctx context.Context, year int, t Threshold) (*time.Time, error) {
	col := alertColumn(t)
	if col == "" {
		return nil, fmt.Errorf("invalid threshold %d", t)
	}
	var at *time.Time
	err := s.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM billing_alert_state WHERE year = $1
	`, col), year).Scan(&at)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return at, nil
}

func alertColumn(t Threshold) string {
	switch t {
	case Threshold80:
		return "threshold_80_sent"
	case Threshold90:
		return "threshold_90_sent"
	case Threshold95:
		return "threshold_95_sent"
	case ThresholdHard:
		return "threshold_hard_sent"
	}
	return ""
}

// startOfYear — 1 января 00:00 в указанной timezone.
func startOfYear(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, loc)
}
