package report

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// Create POST /api/reports — игрок подаёт жалобу.
// Body: {"target_type":"user","target_id":"...","reason":"spam","comment":"..."}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Reason     string `json:"reason"`
		Comment    string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	rep, err := h.svc.Create(r.Context(), uid, CreateInput{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Reason:     req.Reason,
		Comment:    req.Comment,
	})
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, map[string]any{"report": rep})
	case errors.Is(err, ErrInvalidTarget):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid target"))
	case errors.Is(err, ErrEmptyReason):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "reason required"))
	case errors.Is(err, ErrTooLong):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "comment too long"))
	case errors.Is(err, ErrSelfReport):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "cannot report yourself"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// AdminList GET /api/admin/reports?status=new&limit=50
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	list, err := h.svc.List(r.Context(), status, limit)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"reports": list})
}

// AdminResolve POST /api/admin/reports/{id}/resolve
// Body: {"status":"resolved","note":"warning issued"}
func (h *Handler) AdminResolve(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	err := h.svc.Resolve(r.Context(), id, uid, req.Status, req.Note)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrAlreadyClosed):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "already resolved"))
	case errors.Is(err, ErrTooLong):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "note too long"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	}
}
