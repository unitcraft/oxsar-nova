package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"github.com/oxsar/nova/backend/internal/locks"
	"github.com/oxsar/nova/backend/pkg/metrics"
)

// JobFunc — функция, исполняемая по расписанию. Вызов уже произошёл
// внутри advisory lock'а (см. README пакета). Возвращаемая ошибка
// логируется и идёт в counter, cron не делает ретрая.
type JobFunc func(ctx context.Context) error

// Scheduler — обёртка над cron.Cron с lock-wrap'ом и метриками.
//
// Использование:
//
//	sch := scheduler.New(cfg, pool, log)
//	sch.Register("alien_spawn", alienSvc.Spawn)
//	sch.Register("score_recalc_all", scoreSvc.RecalcAll)
//	if err := sch.Start(ctx); err != nil { ... }
//	defer sch.Stop()
type Scheduler struct {
	cfg  *Config
	pool *pgxpool.Pool
	log  *slog.Logger

	cron *cron.Cron

	mu       sync.Mutex
	jobs     map[string]cron.EntryID
	disabled bool
}

// New создаёт Scheduler. Если SCHEDULER_DISABLE_ALL=true — все
// последующие Register становятся no-op (логируется), Start ничего
// не запускает.
func New(cfg *Config, pool *pgxpool.Pool, log *slog.Logger) *Scheduler {
	if log == nil {
		log = slog.Default()
	}
	return &Scheduler{
		cfg:      cfg,
		pool:     pool,
		log:      log,
		cron:     cron.New(cron.WithParser(cron.NewParser(cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor))),
		jobs:     map[string]cron.EntryID{},
		disabled: DisabledByEnv(),
	}
}

// Register регистрирует job'у с именем name и функцией fn. Если в
// конфиге нет записи с этим именем — ошибка (так разработчик увидит,
// что забыл добавить в YAML). Если job отключена (enabled: false) —
// тихо пропускается.
//
// fn оборачивается так:
//  1. metrics.SchedulerJobLastRun (gauge) обновляется в начале tick'а.
//  2. locks.TryRun("scheduler:"+name, ...) — берёт advisory lock.
//  3. Если acquired=false → counter "skip", выход.
//  4. Иначе fn(ctx); ошибка → counter "error" + log; success → "ok".
//  5. metrics.SchedulerJobDuration измеряет только не-skip запуски.
func (s *Scheduler) Register(name string, fn JobFunc) error {
	if name == "" {
		return fmt.Errorf("scheduler: empty job name")
	}
	if fn == nil {
		return fmt.Errorf("scheduler: job %q: nil fn", name)
	}
	if s.disabled {
		s.log.Info("scheduler disabled, job skipped", slog.String("job", name))
		return nil
	}

	jc, ok := s.cfg.Jobs[name]
	if !ok {
		return fmt.Errorf("scheduler: job %q not in config (add to schedule.yaml)", name)
	}
	if !jc.Enabled {
		s.log.Info("scheduler job disabled in config",
			slog.String("job", name), slog.String("schedule", jc.Schedule))
		return nil
	}

	wrapped := s.wrap(name, fn)

	s.mu.Lock()
	defer s.mu.Unlock()
	id, err := s.cron.AddFunc(jc.Schedule, wrapped)
	if err != nil {
		return fmt.Errorf("scheduler: register %q: %w", name, err)
	}
	s.jobs[name] = id
	s.log.Info("scheduler job registered",
		slog.String("job", name),
		slog.String("schedule", jc.Schedule),
		slog.String("description", jc.Description))
	return nil
}

// wrap — closure, которую cron вызывает на каждом tick'е. Без ctx
// от cron, поэтому делаем свой Background; deadline не ставим (job'ы
// могут быть длинными типа score_recalc — пусть отрабатывают до конца
// или до Stop).
func (s *Scheduler) wrap(name string, fn JobFunc) func() {
	return func() {
		ctx := context.Background()
		s.runJob(ctx, name, fn)
	}
}

// runJob выделена для тестируемости (можно дёрнуть напрямую без cron-tick'а).
func (s *Scheduler) runJob(ctx context.Context, name string, fn JobFunc) {
	now := time.Now()
	if metrics.SchedulerJobLastRun != nil {
		metrics.SchedulerJobLastRun.WithLabelValues(name).Set(float64(now.Unix()))
	}

	lockName := "scheduler:" + name
	acquired, err := locks.TryRun(ctx, s.pool, lockName, func(ctx context.Context) error {
		return fn(ctx)
	})

	if !acquired {
		// err может быть != nil (например, не удалось взять connection)
		// — это не "skip из-за чужого lock'а", а реальный сбой. Логируем
		// и считаем как error, чтобы алерты не прятали проблемы.
		if err != nil {
			s.log.Error("scheduler job pre-lock error",
				slog.String("job", name), slog.String("err", err.Error()))
			if metrics.SchedulerJobRuns != nil {
				metrics.SchedulerJobRuns.WithLabelValues(name, "error").Inc()
			}
			return
		}
		s.log.Debug("scheduler job skipped (lock held by another instance)",
			slog.String("job", name))
		if metrics.SchedulerJobRuns != nil {
			metrics.SchedulerJobRuns.WithLabelValues(name, "skip").Inc()
		}
		return
	}

	dur := time.Since(now)
	if metrics.SchedulerJobDuration != nil {
		metrics.SchedulerJobDuration.WithLabelValues(name).Observe(dur.Seconds())
	}

	if err != nil {
		s.log.Error("scheduler job failed",
			slog.String("job", name),
			slog.Duration("duration", dur),
			slog.String("err", err.Error()))
		if metrics.SchedulerJobRuns != nil {
			metrics.SchedulerJobRuns.WithLabelValues(name, "error").Inc()
		}
		return
	}
	s.log.Info("scheduler job ok",
		slog.String("job", name),
		slog.Duration("duration", dur))
	if metrics.SchedulerJobRuns != nil {
		metrics.SchedulerJobRuns.WithLabelValues(name, "ok").Inc()
	}
}

// Start запускает cron-тикер. Не блокирует. Если scheduler отключён
// (SCHEDULER_DISABLE_ALL) или ни одна job не зарегистрирована —
// логирует и возвращает nil.
func (s *Scheduler) Start(ctx context.Context) error {
	if s.disabled {
		s.log.Info("scheduler disabled by env (SCHEDULER_DISABLE_ALL)")
		return nil
	}
	s.mu.Lock()
	count := len(s.jobs)
	s.mu.Unlock()
	if count == 0 {
		s.log.Info("scheduler started with 0 jobs (no-op)")
		return nil
	}
	s.cron.Start()
	s.log.Info("scheduler started", slog.Int("jobs", count))
	return nil
}

// Stop останавливает cron-тикер и ждёт завершения активных job'ов
// (cron возвращает ctx, который Done() при завершении).
//
// Должен вызываться при shutdown'е worker'а. Возвращает context.Canceled
// если grace period истёк до завершения job'ы.
func (s *Scheduler) Stop(ctx context.Context) error {
	if s.disabled {
		return nil
	}
	stopCtx := s.cron.Stop()
	select {
	case <-stopCtx.Done():
		s.log.Info("scheduler stopped")
		return nil
	case <-ctx.Done():
		s.log.Warn("scheduler stop: grace exceeded, forcing exit",
			slog.String("err", ctx.Err().Error()))
		return ctx.Err()
	}
}

// JobNames возвращает зарегистрированные имена. Используется в тестах
// и для health-endpoint'а.
func (s *Scheduler) JobNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.jobs))
	for n := range s.jobs {
		names = append(names, n)
	}
	return names
}
