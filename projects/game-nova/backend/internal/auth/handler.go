package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/httpx"
)

// Handler — HTTP-адаптер к /api/me и /api/me/vacation.
//
// План 36 Ф.12: Register/Login/Refresh переехали в identity-service. Здесь
// остаются только эндпоинты текущего пользователя (профиль, vacation).
type Handler struct {
	db       *pgxpool.Pool
	vacation *VacationService
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db, vacation: NewVacationService(db)}
}

// Me GET /api/me — возвращает user_id, username, roles, credit, profession текущего пользователя.
// Требует Middleware (Bearer / ?token=). План 52: users.role удалён, роли
// приходят в JWT (claims.Roles) от identity-service.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var username, profession string
	var credit float64
	var vacationSince, vacationLastEnd *time.Time
	if err := h.db.QueryRow(context.Background(),
		`SELECT username, credit, COALESCE(profession, 'none'),
		        vacation_since, vacation_last_end
		 FROM users WHERE id=$1`,
		uid,
	).Scan(&username, &credit, &profession, &vacationSince, &vacationLastEnd); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	var roles []string
	if claims, ok := RSAClaims(r.Context()); ok && claims != nil {
		roles = claims.Roles
	}
	resp := map[string]any{
		"user_id":    uid,
		"username":   username,
		"roles":      roles,
		"credit":     credit,
		"profession": profession,
	}
	if vacationSince != nil {
		resp["vacation_since"] = vacationSince.UTC().Format(time.RFC3339)
		// Минимально можно выйти через 48h от vacation_since.
		resp["vacation_unlock_at"] = vacationSince.Add(VacationMinDuration).UTC().Format(time.RFC3339)
	}
	if vacationLastEnd != nil {
		resp["vacation_last_end"] = vacationLastEnd.UTC().Format(time.RFC3339)
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

// SetVacation POST /api/me/vacation — включить режим отпуска.
func (h *Handler) SetVacation(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	if err := h.vacation.SetVacation(r.Context(), uid); err != nil {
		switch {
		case errors.Is(err, ErrVacationAlreadyActive):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "vacation already active"))
		case errors.Is(err, ErrVacationIntervalNotMet):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "vacation interval not met (20 days)"))
		case errors.Is(err, ErrVacationBlocked):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "vacation blocked by pending events (build/fleet/research) — wait for them to finish"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UnsetVacation DELETE /api/me/vacation — выйти из режима отпуска.
func (h *Handler) UnsetVacation(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	if err := h.vacation.UnsetVacation(r.Context(), uid); err != nil {
		switch {
		case errors.Is(err, ErrVacationNotActive):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "vacation not active"))
		case errors.Is(err, ErrVacationTooEarly):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "vacation must last at least 48h before you can exit"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

