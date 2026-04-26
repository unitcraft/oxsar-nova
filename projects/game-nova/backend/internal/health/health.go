// Package health — runtime-состояние процесса (live/draining/ready)
// + HTTP-handlers для /api/health и /api/ready.
//
// План 31 Ф.1. Используется для zero-downtime deploy: при SIGTERM
// процесс переходит в draining-state, /api/health начинает возвращать
// 503, nginx (или другой балансировщик) убирает upstream из пула,
// затем — обычный srv.Shutdown.
//
// Использование (server):
//
//	state := health.NewState("backend", buildVersion)
//	state.SetReady() // после успешной инициализации БД и т.п.
//	r.Get("/api/health", state.HealthHandler())
//	r.Get("/api/ready",  state.ReadyHandler(pool))
//
//	<-ctx.Done() // SIGTERM
//	state.SetDraining()
//	time.Sleep(10 * time.Second)
//	srv.Shutdown(...)
//
// Worker использует тот же State, но без HTTP — только State.IsDraining()
// в event-loop, чтобы перестать брать новые события.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// State — runtime-состояние процесса. Безопасен для concurrent-доступа.
type State struct {
	component string
	version   string
	startedAt time.Time

	ready    atomic.Bool
	draining atomic.Bool
}

// NewState создаёт state. component — имя процесса (backend|worker).
// version — git-tag/sha (для /api/health output).
func NewState(component, version string) *State {
	return &State{
		component: component,
		version:   version,
		startedAt: time.Now(),
	}
}

// SetReady — процесс готов принимать запросы. Вызывается после
// успешной инициализации зависимостей.
func (s *State) SetReady() { s.ready.Store(true) }

// SetDraining — процесс в shutdown-state, отказывается от новых запросов.
// /api/health начинает отвечать 503. Идемпотентно.
func (s *State) SetDraining() { s.draining.Store(true) }

// IsReady — true если SetReady() был вызван.
func (s *State) IsReady() bool { return s.ready.Load() }

// IsDraining — true если SetDraining() был вызван.
func (s *State) IsDraining() bool { return s.draining.Load() }

// Status — JSON-структура для health/ready ответов.
type Status struct {
	Component string `json:"component"`
	Status    string `json:"status"`
	Version   string `json:"version,omitempty"`
	UptimeSec int64  `json:"uptime_sec"`
	Reason    string `json:"reason,omitempty"`
}

// HealthHandler возвращает HTTP-handler для liveness check.
//
//	200 OK — процесс жив, готов принимать запросы.
//	503    — draining (shutdown в процессе).
//
// Не делает БД-вызовов: даже при упавшей БД health=200, чтобы
// orchestrator не убил pod на временную проблему БД.
func (s *State) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		st := Status{
			Component: s.component,
			Version:   s.version,
			UptimeSec: int64(time.Since(s.startedAt).Seconds()),
		}
		code := http.StatusOK
		st.Status = "ok"
		if s.draining.Load() {
			code = http.StatusServiceUnavailable
			st.Status = "draining"
			st.Reason = "shutdown in progress"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(st)
	}
}

// Pinger — узкий интерфейс к БД-пулу для readiness-check.
type Pinger interface {
	Ping(ctx context.Context) error
}

// ReadyHandler — readiness check.
//
//	200 OK — БД доступна, процесс ready.
//	503    — draining ИЛИ БД недоступна ИЛИ ещё не SetReady().
//
// Используется orchestrator'ом для решения «слать ли запросы на этот
// instance». Отличается от HealthHandler тем, что проверяет deps.
func (s *State) ReadyHandler(p Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		st := Status{
			Component: s.component,
			Version:   s.version,
			UptimeSec: int64(time.Since(s.startedAt).Seconds()),
		}
		code := http.StatusOK
		st.Status = "ready"

		switch {
		case s.draining.Load():
			code = http.StatusServiceUnavailable
			st.Status = "draining"
			st.Reason = "shutdown in progress"
		case !s.ready.Load():
			code = http.StatusServiceUnavailable
			st.Status = "starting"
			st.Reason = "initialization in progress"
		default:
			if p != nil {
				ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
				defer cancel()
				if err := p.Ping(ctx); err != nil {
					code = http.StatusServiceUnavailable
					st.Status = "db_unhealthy"
					st.Reason = err.Error()
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(st)
	}
}
