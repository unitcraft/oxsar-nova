// Package admin — RBAC для административных эндпоинтов (план 14 Ф.8.1).
//
// Иерархия ролей (по возрастанию прав):
//   player (0)   — обычный игрок, доступа к /api/admin нет
//   support (1)  — модератор: read-only + базовая модерация (ban/unban,
//                  view профилей, view events, audit-журнал)
//   admin (2)    — полный доступ к /api/admin, кроме role change
//   superadmin (3) — всё, включая смену ролей других админов
//
// Использование: оборачиваем роут middleware'ом
//     ar.With(admin.RequireRole(admin.RoleAdmin)).Post("/credit", …)
//
// Обратная совместимость: AdminOnly оставлен (работает как
// RequireRole(RoleAdmin)), чтобы не ломать существующие ссылки.

package admin

import (
	"context"
	"net/http"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/repo"
)

// Role — строгий enum для проверок прав.
type Role int

const (
	RolePlayer Role = iota
	RoleSupport
	RoleAdmin
	RoleSuperadmin
)

// roleFromDB переводит строку из users.role в Role. Неизвестные
// значения → RolePlayer (самый низкий уровень, безопасный fallback).
func roleFromDB(s string) Role {
	switch s {
	case "superadmin":
		return RoleSuperadmin
	case "admin":
		return RoleAdmin
	case "support":
		return RoleSupport
	default:
		return RolePlayer
	}
}

// RequireRole возвращает middleware, который пропускает запрос только
// если у пользователя роль >= min. Иначе 403.
func RequireRole(db repo.Exec, min Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid, _ := auth.UserID(r.Context())
			if uid == "" {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			role, ok := loadRole(r.Context(), db, uid)
			if !ok {
				httpx.WriteError(w, r, httpx.ErrForbidden)
				return
			}
			if role < min {
				httpx.WriteError(w, r, httpx.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func loadRole(ctx context.Context, db repo.Exec, uid string) (Role, bool) {
	var s string
	err := db.Pool().QueryRow(ctx, `SELECT role FROM users WHERE id = $1`, uid).Scan(&s)
	if err != nil {
		return RolePlayer, false
	}
	return roleFromDB(s), true
}
