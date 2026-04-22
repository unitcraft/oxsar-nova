package score

import (
	"net/http"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-эндпоинты рейтинга.
type Handler struct {
	svc *Service
}

// NewHandler создаёт Handler.
func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

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
