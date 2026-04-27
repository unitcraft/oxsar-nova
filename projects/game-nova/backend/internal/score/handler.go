package score

import (
	"net/http"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/repo"
)

// Handler — HTTP-эндпоинты рейтинга.
type Handler struct {
	svc *Service
	db  repo.Exec
}

// NewHandler создаёт Handler.
func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// NewHandlerWithDB создаёт Handler с доступом к БД для функций Stats.
func NewHandlerWithDB(s *Service, db repo.Exec) *Handler {
	return &Handler{svc: s, db: db}
}

// Highscore GET /api/highscore?type=total|b|r|u|a&limit=N
//
// Возвращает топ игроков. type по умолчанию = "total", limit = 100 (max 200).
func (h *Handler) Highscore(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	scoreType := r.URL.Query().Get("type")
	if scoreType == "" {
		scoreType = "total"
	}
	limit := 100
	entries, err := h.svc.Top(r.Context(), scoreType, limit)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"highscore": entries})
}

// MyRank GET /api/highscore/me?type=total|b|r|u|a
//
// Возвращает позицию текущего игрока в рейтинге.
func (h *Handler) MyRank(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	scoreType := r.URL.Query().Get("type")
	if scoreType == "" {
		scoreType = "total"
	}
	rank, err := h.svc.PlayerRank(r.Context(), uid, scoreType)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	pts, err := h.svc.PlayerScore(r.Context(), uid, scoreType)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	ePts, err := h.svc.PlayerScore(r.Context(), uid, "e")
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"rank": rank, "type": scoreType, "points": pts, "e_points": ePts,
	})
}

// Alliances GET /api/highscore/alliances — рейтинг альянсов по суммарным очкам.
func (h *Handler) Alliances(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	entries, err := h.svc.TopAlliances(r.Context(), 100)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"alliances": entries})
}

// Vacation GET /api/highscore/vacation — список игроков в режиме отпуска.
func (h *Handler) Vacation(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	entries, err := h.svc.VacationPlayers(r.Context())
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"players": entries})
}

type transferRow struct {
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Total    float64 `json:"total"`
	Metal    float64 `json:"metal"`
	Silicon  float64 `json:"silicon"`
	Hydrogen float64 `json:"hydrogen"`
}

// ResourceTransfers GET /api/stats/resource-transfers?direction=sent|received&period=week|month|all
func (h *Handler) ResourceTransfers(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	if h.db == nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, "db unavailable"))
		return
	}
	direction := r.URL.Query().Get("direction")
	if direction != "sent" && direction != "received" {
		direction = "received"
	}
	period := r.URL.Query().Get("period")
	var periodClause string
	switch period {
	case "week":
		periodClause = "AND at > now() - interval '7 days'"
	case "month":
		periodClause = "AND at > now() - interval '30 days'"
	default:
		periodClause = ""
	}
	col := "to_user_id"
	if direction == "sent" {
		col = "from_user_id"
	}
	query := `
		SELECT u.id, u.username,
		       SUM(rt.metal)::float + 2*SUM(rt.silicon)::float + 4*SUM(rt.hydrogen)::float AS total,
		       SUM(rt.metal)::float, SUM(rt.silicon)::float, SUM(rt.hydrogen)::float
		FROM resource_transfers rt
		JOIN users u ON u.id = rt.` + col + ` AND u.deleted_at IS NULL
		WHERE rt.` + col + ` IS NOT NULL ` + periodClause + `
		GROUP BY u.id, u.username
		ORDER BY total DESC
		LIMIT 20
	`
	rows, err := h.db.Pool().Query(r.Context(), query)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()
	out := []transferRow{}
	for rows.Next() {
		var row transferRow
		if err := rows.Scan(&row.UserID, &row.Username, &row.Total, &row.Metal, &row.Silicon, &row.Hydrogen); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		out = append(out, row)
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"transfers": out, "direction": direction, "period": period})
}

// Stats GET /api/stats — счётчик онлайна.
//
// Возвращает количество игроков, игравших за последние 24 часа и прямо сейчас.
// Доступна без авторизации (публичная статистика).
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, "stats unavailable"))
		return
	}

	var online24h int64
	// Игроки, заходившие в течение последних 24 часов (не забанены).
	err := h.db.Pool().QueryRow(r.Context(), `
		SELECT COUNT(*) FROM users
		WHERE umode = false AND last_seen > now() - interval '24 hours'
	`).Scan(&online24h)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	var onlineNow int64
	// Игроки, активные в последние 5 минут (приблизительный онлайн).
	err = h.db.Pool().QueryRow(r.Context(), `
		SELECT COUNT(*) FROM users
		WHERE umode = false AND last_seen > now() - interval '5 minutes'
	`).Scan(&onlineNow)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"online_now": onlineNow,
		"online_24h": online24h,
	})
}
