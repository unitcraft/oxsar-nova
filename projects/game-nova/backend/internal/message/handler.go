package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// Inbox GET /api/messages
func (h *Handler) Inbox(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	list, err := h.svc.Inbox(r.Context(), uid, 100)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"messages": list})
}

// Sent GET /api/messages/sent
func (h *Handler) Sent(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	list, err := h.svc.Sent(r.Context(), uid, 100)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"messages": list})
}

// UnreadCount GET /api/messages/unread-count
func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	n, err := h.svc.UnreadCount(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]int{"unread": n})
}

// MarkRead POST /api/messages/{id}/read
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}
	err := h.svc.MarkRead(r.Context(), uid, id)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrMessageNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwned):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Compose POST /api/messages
// Body: {"to": "username", "subject": "...", "body": "..."}
func (h *Handler) Compose(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if req.To == "" || req.Subject == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "to and subject required"))
		return
	}
	err := h.svc.Compose(r.Context(), uid, req.To, req.Subject, req.Body)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrRecipientNotFound):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "recipient not found"))
	case errors.Is(err, ErrSelfMessage):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "cannot send to yourself"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// DeleteAll DELETE /api/messages?folder=N — удаляет все (или все в папке).
func (h *Handler) DeleteAll(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	folder := 0
	if f := r.URL.Query().Get("folder"); f != "" {
		fmt.Sscanf(f, "%d", &folder) //nolint:errcheck
	}
	if err := h.svc.DeleteAll(r.Context(), uid, folder); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete DELETE /api/messages/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}
	err := h.svc.Delete(r.Context(), uid, id)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrMessageNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// GetExpeditionReport GET /api/expedition-reports/{id}
func (h *Handler) GetExpeditionReport(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}
	rep, err := h.svc.GetExpeditionReport(r.Context(), uid, id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, rep)
	case errors.Is(err, ErrMessageNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwned):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// GetEspionageReport GET /api/espionage-reports/{id}
func (h *Handler) GetEspionageReport(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}
	rep, err := h.svc.GetEspionageReport(r.Context(), uid, id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, rep)
	case errors.Is(err, ErrMessageNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwned):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// GetReport GET /api/battle-reports/{id}
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}
	rep, err := h.svc.GetBattleReport(r.Context(), uid, id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, rep)
	case errors.Is(err, ErrMessageNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwned):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
