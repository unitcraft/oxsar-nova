package report

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"oxsar/portal/internal/httpx"
	"oxsar/portal/internal/portalsvc"
)

// Handler — HTTP-адаптер для пакета report.
type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// Create — POST /api/reports — игрок подаёт жалобу.
//
// Body: {"target_type":"user","target_id":"...","reason":"spam","comment":"..."}
//
// Аутентификация: JWT уже распарсен portalsvc.Middleware'ом, reporter_id
// берётся из ctx (subject claim).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := portalsvc.UserIDFromContext(r.Context())
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
	if err := decodeJSON(r, &req); err != nil {
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

// AdminList — GET /api/admin/reports?status=new&limit=50
//
// Доступ: admin (через portalsvc.AdminMiddleware) — модератор смотрит
// очередь жалоб. Авторизация делается на уровне роутинга в main.go,
// здесь лишь читаем query-параметры.
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 0
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
	if list == nil {
		list = []Report{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"reports": list})
}

// AdminResolve — POST /api/admin/reports/{id}/resolve
//
// Body: {"status":"resolved","note":"warning issued"}
//
// status: 'resolved' | 'rejected'. После успеха — 204 No Content.
func (h *Handler) AdminResolve(w http.ResponseWriter, r *http.Request) {
	uid, ok := portalsvc.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := decodeJSON(r, &req); err != nil {
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

func decodeJSON(r *http.Request, into any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(into)
}
