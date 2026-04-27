package referral

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

type referredUser struct {
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Points   float64 `json:"points"`
	RegTime  string  `json:"reg_time"`
}

type response struct {
	InvitedCount   int            `json:"invited_count"`
	BonusPoints    float64        `json:"bonus_points"`
	MaxBonusPoints int            `json:"max_bonus_points"`
	CreditPercent  float64        `json:"credit_percent"`
	Referred       []referredUser `json:"referred"`
}

// Mine GET /api/referrals — список приглашённых пользователей + бонусы.
func (h *Handler) Mine(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT id, username, points, COALESCE(regtime, created_at)
		FROM users
		WHERE referred_by = $1
		ORDER BY COALESCE(regtime, created_at) DESC
	`, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	var referred []referredUser
	for rows.Next() {
		var u referredUser
		var regTime time.Time
		if err := rows.Scan(&u.UserID, &u.Username, &u.Points, &regTime); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		u.RegTime = regTime.UTC().Format(time.RFC3339)
		referred = append(referred, u)
	}
	if err := rows.Err(); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	if referred == nil {
		referred = []referredUser{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, response{
		InvitedCount:   len(referred),
		BonusPoints:    float64(len(referred) * BonusPoints),
		MaxBonusPoints: MaxBonusPoints,
		CreditPercent:  CreditPercent,
		Referred:       referred,
	})
}
