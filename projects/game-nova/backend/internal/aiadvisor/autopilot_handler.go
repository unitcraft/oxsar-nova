package aiadvisor

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

// AutopilotHandler — HTTP-адаптер для AutopilotService.
//
// Ответственности handler-а:
//   - извлечь userID из контекста (auth middleware);
//   - распарсить body / path params;
//   - вызвать сервис;
//   - смапить доменные ошибки в HTTP-коды.
type AutopilotHandler struct {
	svc *AutopilotService
}

// NewAutopilotHandler — конструктор.
func NewAutopilotHandler(svc *AutopilotService) *AutopilotHandler {
	return &AutopilotHandler{svc: svc}
}

type adviseRequest struct {
	Strategy string `json:"strategy"`
}

// Advise POST /api/ai-advisor/autopilot/advise
//
// Ставит задание автопилота в очередь воркера, списывает
// AutopilotCostCredits и возвращает {job_id} для последующего polling.
func (h *AutopilotHandler) Advise(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "unauthorized"))
		return
	}
	var req adviseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if req.Strategy == "" {
		req.Strategy = string(StrategyAuto)
	}

	res, err := h.svc.Enqueue(r.Context(), uid, Strategy(req.Strategy))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStrategy):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid strategy"))
		case errors.Is(err, ErrNotEnoughCredit):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not enough credits"))
		case errors.Is(err, ErrRateLimitReached):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "daily limit reached"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, res)
}

// Result GET /api/ai-advisor/autopilot/result/{jobID}
//
// Polling-эндпоинт. Возвращает текущее состояние job'а. Фронт повторяет
// запрос каждые 1–2 сек, пока status != "pending".
func (h *AutopilotHandler) Result(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "unauthorized"))
		return
	}
	jobID := chi.URLParam(r, "jobID")
	if jobID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "job_id is required"))
		return
	}

	res, err := h.svc.Result(r.Context(), uid, jobID)
	if err != nil {
		switch {
		case errors.Is(err, ErrJobNotFound):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "job not found"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, res)
}

type executeRequest struct {
	JobID string `json:"job_id"`
	RecID string `json:"rec_id"`
}

// Execute POST /api/ai-advisor/autopilot/execute
//
// Игрок выбрал одну из топ-3 рекомендаций. Backend валидирует и
// вызывает соответствующий сервисный метод (building.Enqueue / ...).
func (h *AutopilotHandler) Execute(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "unauthorized"))
		return
	}
	var req executeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if req.JobID == "" || req.RecID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "job_id and rec_id are required"))
		return
	}

	res, err := h.svc.Execute(r.Context(), uid, req.JobID, req.RecID)
	if err != nil {
		switch {
		case errors.Is(err, ErrJobNotFound):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "job not found"))
		case errors.Is(err, ErrJobNotReady):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "job not ready"))
		case errors.Is(err, ErrRecommendationNotFound):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "recommendation not found"))
		case errors.Is(err, ErrUnsupportedCategory):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "category not supported yet"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, res)
}
