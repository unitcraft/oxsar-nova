// Package admin — RBAC для административных эндпоинтов.
//
// План 52 (RBAC unification): единый источник правды для ролей —
// identity-service. Game-nova больше НЕ хранит users.role локально,
// а читает roles/permissions из JWT-claims (выпущенных identity-сервисом
// при логине/refresh).
//
// История:
//   - План 14 Ф.8.1: ввели локальную ENUM users.role (player/support/
//     admin/superadmin) с RequireRole(min) проверкой по иерархии.
//   - План 52: ENUM users.role удалён миграцией; RequireRole теперь
//     обёртка над permission-проверкой по фиксированному mapping
//     (legacy role → set of required permissions).
//
// Использование (legacy API, оставлен для совместимости):
//     ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/credit", …)
//
// Использование (новое, предпочтительное):
//     ar.With(admin.RequirePermission("game:credits:grant")).Post("/credit", …)

package admin

import (
	"net/http"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/repo"
)

// Role — legacy enum, оставлен для backward-compat вызовов
// RequireRole(db, RoleX). Внутри маппится на permission-set.
type Role int

const (
	RolePlayer Role = iota
	RoleSupport
	RoleAdmin
	RoleSuperadmin
)

// rolePermissions — маппинг legacy-роли на минимальный permission-set.
// Любая из перечисленных permissions у юзера — пропускает middleware.
//
// Это «или» (any-of), а не «и» — для совместимости с прежней семантикой
// иерархии (admin покрывает support и т.д.). Identity-service грантит
// permissions через role_permissions (план 52 Ф.1 seed), так что юзер
// с ролью admin получит весь admin-set.
var rolePermissions = map[Role][]string{
	RolePlayer:     nil, // Никаких ограничений — но и доступ только если что-то есть.
	RoleSupport:    {"users:read"}, // Минимум support — read юзеров.
	RoleAdmin:      {"users:delete", "game:events:retry", "game:planets:transfer"},
	RoleSuperadmin: {"roles:grant", "system:config"},
}

// RequireRole возвращает middleware, который пропускает запрос, если у
// юзера в JWT есть хотя бы одна permission из rolePermissions[min].
// Параметр db оставлен в сигнатуре для backward-compat (более не
// используется — load из JWT, не из БД).
func RequireRole(db repo.Exec, min Role) func(http.Handler) http.Handler {
	required := rolePermissions[min]
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid, _ := auth.UserID(r.Context())
			if uid == "" {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			claims, ok := auth.RSAClaims(r.Context())
			if !ok {
				httpx.WriteError(w, r, httpx.ErrForbidden)
				return
			}
			if !hasAnyPermission(claims.Permissions, required) {
				httpx.WriteError(w, r, httpx.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission — новый идиоматичный middleware: проверяет наличие
// конкретной permission в JWT. Предпочтительный вариант для новых
// endpoints.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.RSAClaims(r.Context())
			if !ok {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			if !hasPermission(claims.Permissions, permission) {
				httpx.WriteError(w, r, httpx.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// hasPermission возвращает true, если perm присутствует в списке.
func hasPermission(perms []string, perm string) bool {
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// hasAnyPermission возвращает true, если у юзера есть хотя бы одна
// из required permissions. Если required пустой/nil — пропускает всех
// (RolePlayer-семантика).
func hasAnyPermission(userPerms []string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	for _, req := range required {
		if hasPermission(userPerms, req) {
			return true
		}
	}
	return false
}
