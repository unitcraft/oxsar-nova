package identitysvc

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"oxsar/identity/internal/httpx"
)

// RBACHandler — HTTP-адаптер для RBAC API. Все endpoints требуют JWT
// (через middleware) + permission-check внутри handler-метода.
type RBACHandler struct {
	rbac *RBACService
}

// NewRBACHandler создаёт handler поверх RBACService.
func NewRBACHandler(rbac *RBACService) *RBACHandler {
	return &RBACHandler{rbac: rbac}
}

// listRolesResp — DTO для GET /api/admin/roles.
type listRolesResp struct {
	Roles []Role `json:"roles"`
}

// ListRoles GET /api/admin/roles — требует permission "roles:read".
func (h *RBACHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "roles:read") {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusForbidden, Code: "forbidden", Message: "missing permission: roles:read"})
		return
	}
	roles, err := h.rbac.ListRoles(r.Context())
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusInternalServerError, Code: "internal", Message: err.Error()})
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, listRolesResp{Roles: roles})
}

// GetRolePermissions GET /api/admin/roles/{id}/permissions
func (h *RBACHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "roles:read") {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusForbidden, Code: "forbidden", Message: "missing permission: roles:read"})
		return
	}
	idStr := chi.URLParam(r, "id")
	roleID, err := strconv.Atoi(idStr)
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "invalid role id"})
		return
	}
	perms, err := h.rbac.GetRolePermissions(r.Context(), roleID)
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusInternalServerError, Code: "internal", Message: err.Error()})
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"permissions": perms})
}

// ListUserRoles GET /api/admin/users/{id}/roles
func (h *RBACHandler) ListUserRoles(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "roles:read") {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusForbidden, Code: "forbidden", Message: "missing permission: roles:read"})
		return
	}
	uid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "invalid user id"})
		return
	}
	assignments, err := h.rbac.ListUserRoles(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusInternalServerError, Code: "internal", Message: err.Error()})
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"assignments": assignments})
}

// grantRoleReq — body для POST /api/admin/users/{id}/roles.
type grantRoleReq struct {
	Role      string     `json:"role"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Reason    string     `json:"reason"`
}

// GrantUserRole POST /api/admin/users/{id}/roles — требует "roles:grant".
func (h *RBACHandler) GrantUserRole(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "roles:grant") {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusForbidden, Code: "forbidden", Message: "missing permission: roles:grant"})
		return
	}
	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "invalid user id"})
		return
	}
	var req grantRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "invalid json"})
		return
	}
	if req.Role == "" || req.Reason == "" {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "role and reason are required"})
		return
	}

	actorID, ok := actorIDFromContext(r)
	if !ok {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusUnauthorized, Code: "unauthorized", Message: "no actor in context"})
		return
	}

	ip := remoteIP(r)
	err = h.rbac.GrantUserRole(r.Context(), actorID, targetID, req.Role, GrantOptions{
		ExpiresAt: req.ExpiresAt,
		Reason:    req.Reason,
		IPAddress: ip,
		UserAgent: r.UserAgent(),
	})
	switch {
	case errors.Is(err, ErrRoleNotFound):
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusNotFound, Code: "role_not_found", Message: err.Error()})
	case errors.Is(err, ErrUserNotFound):
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusNotFound, Code: "user_not_found", Message: err.Error()})
	case err != nil:
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusInternalServerError, Code: "internal", Message: err.Error()})
	default:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "granted"})
	}
}

// RevokeUserRole DELETE /api/admin/users/{id}/roles/{role}?reason=...
func (h *RBACHandler) RevokeUserRole(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "roles:revoke") {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusForbidden, Code: "forbidden", Message: "missing permission: roles:revoke"})
		return
	}
	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "invalid user id"})
		return
	}
	roleName := chi.URLParam(r, "role")
	reason := strings.TrimSpace(r.URL.Query().Get("reason"))
	if roleName == "" || reason == "" {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "role and reason are required"})
		return
	}

	actorID, ok := actorIDFromContext(r)
	if !ok {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusUnauthorized, Code: "unauthorized", Message: "no actor in context"})
		return
	}
	ip := remoteIP(r)
	err = h.rbac.RevokeUserRole(r.Context(), actorID, targetID, roleName, reason, ip, r.UserAgent())
	switch {
	case errors.Is(err, ErrRoleNotFound):
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusNotFound, Code: "role_not_found", Message: err.Error()})
	case errors.Is(err, ErrNotGranted):
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusNotFound, Code: "not_granted", Message: err.Error()})
	case err != nil:
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusInternalServerError, Code: "internal", Message: err.Error()})
	default:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "revoked"})
	}
}

// QueryAudit GET /api/admin/audit/role-changes
func (h *RBACHandler) QueryAudit(w http.ResponseWriter, r *http.Request) {
	if !hasPermission(r, "audit:read") {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusForbidden, Code: "forbidden", Message: "missing permission: audit:read"})
		return
	}
	q := AuditQuery{
		Limit:  parseIntDefault(r.URL.Query().Get("limit"), 50),
		Offset: parseIntDefault(r.URL.Query().Get("offset"), 0),
		Action: r.URL.Query().Get("action"),
	}
	if v := r.URL.Query().Get("actor_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			q.ActorID = &id
		}
	}
	if v := r.URL.Query().Get("target_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			q.TargetID = &id
		}
	}
	if v := r.URL.Query().Get("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.Since = &t
		}
	}
	if v := r.URL.Query().Get("until"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.Until = &t
		}
	}

	events, err := h.rbac.QueryAuditChanges(r.Context(), q)
	if err != nil {
		httpx.WriteError(w, r, &httpx.Error{Status: http.StatusInternalServerError, Code: "internal", Message: err.Error()})
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"events": events})
}

// === Helpers ===

// hasPermission читает permissions из request context (выставляются JWT
// middleware) и проверяет наличие нужного permission.
func hasPermission(r *http.Request, perm string) bool {
	perms := permissionsFromContext(r)
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// permissionsFromContext извлекает []string permissions из request context.
// Заполняется в JWT middleware (план 52 Ф.2 — middleware расширен).
func permissionsFromContext(r *http.Request) []string {
	v := r.Context().Value(ctxKeyPermissions{})
	if perms, ok := v.([]string); ok {
		return perms
	}
	return nil
}

// actorIDFromContext извлекает uuid юзера из JWT-claims в context.
// Используется ключ ctxKeyUserID, который выставляет middleware при
// успешной валидации JWT.
func actorIDFromContext(r *http.Request) (uuid.UUID, bool) {
	v := r.Context().Value(ctxKeyUserID{})
	if id, ok := v.(uuid.UUID); ok {
		return id, true
	}
	if s, ok := v.(string); ok {
		if id, err := uuid.Parse(s); err == nil {
			return id, true
		}
	}
	return uuid.Nil, false
}

// remoteIP извлекает client IP из request.
func remoteIP(r *http.Request) *net.IP {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		// Берём первый IP (clientside).
		if i := strings.Index(v, ","); i != -1 {
			host = strings.TrimSpace(v[:i])
		} else {
			host = strings.TrimSpace(v)
		}
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}
	return &ip
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

// ctxKeyPermissions — context key для permissions в request context.
// Ставится в JWT middleware при валидации токена.
type ctxKeyPermissions struct{}

// ctxKeyUserID — context key для actor uuid в request context.
type ctxKeyUserID struct{}
