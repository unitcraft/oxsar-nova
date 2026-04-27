package aiadvisor

import (
	"encoding/json"
	"errors"
	"net/http"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type askRequest struct {
	Model    string `json:"model"`
	Question string `json:"question"`
}

// Ask POST /api/ai-advisor/ask
func (h *Handler) Ask(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "unauthorized"))
		return
	}
	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if req.Question == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "question is required"))
		return
	}
	if req.Model == "" {
		req.Model = "claude-haiku-4-5-20251001"
	}

	result, err := h.svc.Ask(r.Context(), uid, req.Model, req.Question)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotEnoughCredit):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not enough credits"))
		case errors.Is(err, ErrRateLimitReached):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "daily limit reached"))
		case errors.Is(err, ErrUnknownModel):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown model"))
		case errors.Is(err, ErrNoBackend):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "ai advisor not configured"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

// Estimate GET /api/ai-advisor/estimate?model=<model>
func (h *Handler) Estimate(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "unauthorized"))
		return
	}
	model := r.URL.Query().Get("model")
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}

	result, err := h.svc.Estimate(r.Context(), uid, model)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnknownModel):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown model"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}
