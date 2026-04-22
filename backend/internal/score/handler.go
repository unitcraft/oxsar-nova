package score

import (
	"net/http"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/repo"
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
