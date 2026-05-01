package auth

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
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

// IGNUnitID — research-технология «Intergalactic Research Network» (план 20 Ф.8).
// Уровень считывается из таблицы research; отсутствие записи трактуется как 0.
// См. configs/units.yml id=113.
const IGNUnitID = 113

// MinerNeedPoints возвращает порог `of_points` для следующего уровня
// шахтёра. Формула из legacy `Functions.inc.php:1642`:
//
//	level<1 → 100, иначе round(pow(1.5, level-1) * 200).
//
// Public функция — переиспользуется в score/recalc и фронт-тестах.
func MinerNeedPoints(level int) int64 {
	if level < 1 {
		return 100
	}
	return int64(math.Round(math.Pow(1.5, float64(level-1)) * 200))
}

// Me GET /api/me — профиль текущего пользователя.
//
// План 72.1 ч.17 (pixel-perfect MainScreen): расширен до полного набора
// stats-полей легаси `main.tpl`:
//
//   - points / rank / max_points    — рейтинг и исторический пик,
//   - combat_experience (e_points)  — тотальный бой-опыт,
//   - accumulated_experience (be_points) — резервуар активных техов,
//   - miner_level / miner_points / miner_need_points — система Шахтёра,
//   - dm_points                      — derived-метрика для альт-рейтинга,
//   - intergalactic_research_level  — уровень research IGN (id=113),
//   - battles                        — счётчик сражений.
//
// Требует Middleware (Bearer / ?token=). План 52: users.role удалён,
// роли приходят в JWT (claims.Roles) от identity-service.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	ctx := r.Context()

	var (
		username, profession                 string
		credit                               float64
		points, ePoints, bePoints, maxPoints float64
		ofPoints, dmPoints                   float64
		ofLevel, battles                     int
		vacationSince, vacationLastEnd       *time.Time
		deleteAt                             *time.Time // план 72.1.30: pending удаление
	)
	if err := h.db.QueryRow(ctx,
		`SELECT username, credit, COALESCE(profession, 'none'),
		        points, e_points, be_points, max_points,
		        of_points, of_level, dm_points, battles,
		        vacation_since, vacation_last_end, delete_at
		 FROM users WHERE id=$1`,
		uid,
	).Scan(
		&username, &credit, &profession,
		&points, &ePoints, &bePoints, &maxPoints,
		&ofPoints, &ofLevel, &dmPoints, &battles,
		&vacationSince, &vacationLastEnd, &deleteAt,
	); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	// Ранг и общее число игроков (для отображения "rank из total").
	var rank, totalUsers int
	if err := h.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)+1 FROM users WHERE points > $1 AND umode = false AND is_observer = false),
			(SELECT COUNT(*)   FROM users WHERE umode = false AND is_observer = false)
	`, points).Scan(&rank, &totalUsers); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	// IGN level — research id=113. Отсутствие записи = уровень 0.
	var ignLevel int
	if err := h.db.QueryRow(ctx,
		`SELECT COALESCE(level, 0) FROM research WHERE user_id=$1 AND unit_id=$2`,
		uid, IGNUnitID,
	).Scan(&ignLevel); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	var roles []string
	if claims, ok := RSAClaims(r.Context()); ok && claims != nil {
		roles = claims.Roles
	}
	resp := map[string]any{
		"user_id":                      uid,
		"username":                     username,
		"roles":                        roles,
		"credit":                       credit,
		"profession":                   profession,
		"points":                       points,
		"rank":                         rank,
		"total_users":                  totalUsers,
		"max_points":                   maxPoints,
		"combat_experience":            ePoints,
		"accumulated_experience":       bePoints,
		"miner_level":                  ofLevel,
		"miner_points":                 ofPoints,
		"miner_need_points":            MinerNeedPoints(ofLevel),
		"dm_points":                    dmPoints,
		"intergalactic_research_level": ignLevel,
		"battles":                      battles,
	}
	if vacationSince != nil {
		resp["vacation_since"] = vacationSince.UTC().Format(time.RFC3339)
		// Минимально можно выйти через 48h от vacation_since.
		resp["vacation_unlock_at"] = vacationSince.Add(VacationMinDuration).UTC().Format(time.RFC3339)
	}
	if vacationLastEnd != nil {
		resp["vacation_last_end"] = vacationLastEnd.UTC().Format(time.RFC3339)
	}
	// План 72.1.30: pending account deletion (grace 7 дней).
	if deleteAt != nil {
		resp["delete_at"] = deleteAt.UTC().Format(time.RFC3339)
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

